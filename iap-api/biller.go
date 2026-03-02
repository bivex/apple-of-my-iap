package iap

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Biller manages subscription business logic and state.
type Biller struct {
	plans             []*Plan
	plansByProductID  map[string]*Plan
	subscriptions     map[string]*Subscription
	mu                sync.RWMutex
	cache             *Cache
	generator         *ReceiptGenerator
	renderer          *Renderer
}

// NewBiller creates a new Biller with the given cache.
func NewBiller(cache *Cache) *Biller {
	return &Biller{
		plans:            make([]*Plan, 0),
		plansByProductID: make(map[string]*Plan),
		subscriptions:    make(map[string]*Subscription),
		cache:            cache,
		generator:        NewReceiptGenerator(),
		renderer:         NewRenderer(),
	}
}

// Start initializes the Biller by loading plans and subscriptions from cache.
func (b *Biller) Start() error {
	log.Println("Reading subs from cache...")

	subs, err := b.cache.ReadSubscriptions()
	if err != nil {
		log.Printf("Warning: failed to read subscriptions from cache: %v", err)
		subs = make(map[string]*Subscription)
	}
	b.subscriptions = subs

	plans, err := b.cache.ReadPlans()
	if err != nil {
		log.Printf("Warning: failed to read plans from cache: %v", err)
		plans = make([]*Plan, 0)
	}
	b.plans = plans
	b.plansByProductID = make(map[string]*Plan)
	for _, p := range plans {
		b.plansByProductID[p.ProductID] = p
	}

	log.Printf("Loaded %d plans and %d subscriptions", len(b.plans), len(b.subscriptions))
	return nil
}

// Shutdown saves the current state to cache.
func (b *Biller) Shutdown() error {
	return b.cache.WriteSubscriptions(b.subscriptions)
}

// GetPlans returns all available plans.
func (b *Biller) GetPlans() []*Plan {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.plans
}

// GetPlanByProductID returns a plan by its product ID.
func (b *Biller) GetPlanByProductID(productID string) (*Plan, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	plan, ok := b.plansByProductID[productID]
	return plan, ok
}

// GetSubscriptions returns all subscriptions.
func (b *Biller) GetSubscriptions() map[string]*Subscription {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make(map[string]*Subscription, len(b.subscriptions))
	for k, v := range b.subscriptions {
		result[k] = v
	}
	return result
}

// GetSubscription returns a subscription by its receipt token.
func (b *Biller) GetSubscription(receiptToken string) (*Subscription, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	sub, ok := b.subscriptions[receiptToken]
	return sub, ok
}

// CreateSub creates a new subscription for the given plan.
func (b *Biller) CreateSub(plan *Plan) (*Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	existingTokens := make(map[string]bool, len(b.subscriptions))
	for token := range b.subscriptions {
		existingTokens[token] = true
	}
	receiptToken := b.generator.GenEncoding(plan, existingTokens)

	_, receiptInfo := b.generator.GenerateReceipt(plan, nil)

	sub := NewSubscription(receiptToken, *receiptInfo)

	b.subscriptions[receiptToken] = sub

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return sub, nil
}

// SetSubStatus updates the status code of a subscription.
func (b *Biller) SetSubStatus(sub *Subscription, statusCode int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub.Status = statusCode
	b.subscriptions[sub.ReceiptToken] = sub

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return nil
}

// RenewSub renews a subscription by adding a new receipt.
func (b *Biller) RenewSub(sub *Subscription) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	plan, ok := b.plansByProductID[sub.LatestReceiptInfo().ProductID]
	if !ok {
		return ErrPlanNotFound
	}

	_, newReceiptInfo := b.generator.GenerateReceipt(plan, sub)

	newReceiptToken := b.generator.GenEncoding(plan, map[string]bool{
		sub.ReceiptToken: true,
	})

	updatedSub := sub.AddReceipt(*newReceiptInfo, newReceiptToken)
	b.subscriptions[sub.ReceiptToken] = updatedSub

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return nil
}

// CancelSub marks a subscription as cancelled.
func (b *Biller) CancelSub(sub *Subscription) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	updatedSub := sub.Cancel()
	b.subscriptions[sub.ReceiptToken] = updatedSub

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return nil
}

// RefundTransaction marks a transaction as refunded.
func (b *Biller) RefundTransaction(sub *Subscription, receiptInfo ReceiptInfo) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Printf("Refunding transaction: %s", receiptInfo.TransactionID)

	updatedSub := sub.Refund(receiptInfo)
	b.subscriptions[sub.ReceiptToken] = updatedSub

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return nil
}

// SendNotification fires an Apple S2S-style notification to webhookURL.
// It builds a fake JWS (unsigned, dev-only) in Apple's App Store Server Notifications v2 format.
// notificationType: SUBSCRIBED, DID_RENEW, DID_FAIL_TO_RENEW, EXPIRED, GRACE_PERIOD_EXPIRED,
//
//	CANCEL, REFUND, REVOKE, PRICE_INCREASE
func (b *Biller) SendNotification(sub *Subscription, notificationType, webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("WEBHOOK_URL not configured")
	}

	b.mu.RLock()
	latest := sub.LatestReceiptInfo()
	b.mu.RUnlock()

	if latest == nil {
		return fmt.Errorf("subscription has no receipt info")
	}

	now := time.Now()
	notificationUUID := uuid.New().String()

	// Inner JWS payload: signedTransactionInfo (transaction details)
	txPayload := map[string]interface{}{
		"originalTransactionId": latest.OriginalTransactionID,
		"transactionId":         latest.TransactionID,
		"productId":             latest.ProductID,
		"purchaseDate":          now.UnixMilli(),
		"expiresDate":           latest.ExpiresDate.UnixMilli(),
		"type":                  "Auto-Renewable Subscription",
		"inAppOwnershipType":    "PURCHASED",
	}
	signedTransactionInfo := buildFakeJWS(txPayload)

	// Inner JWS payload: signedRenewalInfo
	renewalPayload := map[string]interface{}{
		"originalTransactionId":   latest.OriginalTransactionID,
		"productId":               latest.ProductID,
		"autoRenewProductId":      latest.ProductID,
		"autoRenewStatus":         1,
		"signedDate":              now.UnixMilli(),
	}
	signedRenewalInfo := buildFakeJWS(renewalPayload)

	// Outer notification envelope
	notificationPayload := map[string]interface{}{
		"notificationType": notificationType,
		"notificationUUID": notificationUUID,
		"version":          "2.0",
		"signedDate":       now.UnixMilli(),
		"data": map[string]interface{}{
			"bundleId":              "com.example.app",
			"bundleVersion":         "1",
			"environment":           "Sandbox",
			"signedTransactionInfo": signedTransactionInfo,
			"signedRenewalInfo":     signedRenewalInfo,
		},
	}
	outerJWS := buildFakeJWS(notificationPayload)

	req, err := http.NewRequest(http.MethodPost, webhookURL, strings.NewReader(outerJWS))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}

	log.Printf("[notify] %s → %s (%s) → %d", notificationType, webhookURL, notificationUUID, resp.StatusCode)
	return nil
}

// buildFakeJWS encodes payload as a fake JWS compact token (unsigned, dev-only).
// Format: base64url(header) . base64url(payload) . fakesig
func buildFakeJWS(payload interface{}) string {
	header := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"ES256","x5c":["dev-mock"]}`),
	)
	payloadBytes, _ := json.Marshal(payload)
	var buf bytes.Buffer
	enc := base64.NewEncoder(base64.RawURLEncoding, &buf)
	_, _ = enc.Write(payloadBytes)
	enc.Close()
	return header + "." + buf.String() + ".dev-mock-sig"
}

// ClearSubs removes all subscriptions.
func (b *Biller) ClearSubs() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscriptions = make(map[string]*Subscription)

	if err := b.cache.WriteSubscriptions(b.subscriptions); err != nil {
		log.Printf("Warning: failed to write subscriptions to cache: %v", err)
	}

	return nil
}

// VerifyReceipt returns the receipt response for a given receipt token.
func (b *Biller) VerifyReceipt(receiptToken string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	sub, ok := b.subscriptions[receiptToken]
	if !ok {
		resp := &ReceiptResponse{
			StatusCode: UnauthorizedReceipt.Code,
		}
		return b.renderer.RenderReceiptResponse(resp)
	}

	receiptResp := b.generator.GenerateReceiptResponse(sub)
	return b.renderer.RenderReceiptResponse(receiptResp)
}

// Errors
var (
	ErrPlanNotFound = &BillerError{Message: "plan not found"}
	ErrSubNotFound  = &BillerError{Message: "subscription not found"}
)

// BillerError represents an error from the Biller.
type BillerError struct {
	Message string
}

func (e *BillerError) Error() string {
	return e.Message
}
