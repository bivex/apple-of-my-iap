package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	iap "github.com/meetup/iap-api"
)

// Handlers contains the HTTP handlers for the IAP service.
type Handlers struct {
	biller     *iap.Biller
	webhookURL string // backend URL to push S2S notifications to (WEBHOOK_URL env)
}

// New creates a new Handlers instance.
func New(biller *iap.Biller) *Handlers {
	return &Handlers{
		biller:     biller,
		webhookURL: os.Getenv("WEBHOOK_URL"),
	}
}

// RegisterRoutes registers all routes with the Gin engine.
func (h *Handlers) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.indexHandler)
	r.GET("/plans", h.getPlans)
	r.GET("/subs", h.getSubs)
	r.POST("/subs", h.createSub)
	r.POST("/subs/clear", h.clearSubs)
	r.POST("/subs/:receiptToken", h.setSubStatus)
	r.POST("/subs/:receiptToken/renew", h.renewSub)
	r.POST("/subs/:receiptToken/cancel", h.cancelSub)
	r.POST("/subs/:receiptToken/refund/:transactionId", h.refundTransaction)
	r.POST("/subs/:receiptToken/notify/:notificationType", h.sendNotification)
	r.POST("/verifyReceipt", h.verifyReceipt)
}

// indexHandler returns a simple index page.
func (h *Handlers) indexHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Apple IAP Mock Service",
		"version": "1.0.0",
	})
}

// getPlans returns all available plans.
func (h *Handlers) getPlans(c *gin.Context) {
	plans := h.biller.GetPlans()
	c.JSON(http.StatusOK, plans)
}

// getSubs returns all subscriptions, sorted by purchase date (newest first).
func (h *Handlers) getSubs(c *gin.Context) {
	subs := h.biller.GetSubscriptions()

	subsList := make([]*iap.Subscription, 0, len(subs))
	for _, sub := range subs {
		subsList = append(subsList, sub)
	}

	for i := 0; i < len(subsList)-1; i++ {
		for j := i + 1; j < len(subsList); j++ {
			if subsList[i].OriginalReceiptInfo().OriginalPurchaseDate.Before(
				subsList[j].OriginalReceiptInfo().OriginalPurchaseDate.Time) {
				subsList[i], subsList[j] = subsList[j], subsList[i]
			}
		}
	}

	c.JSON(http.StatusOK, subsList)
}

// createSubRequest represents the request to create a subscription.
type createSubRequest struct {
	ProductID string `json:"productId"`
}

// createSub creates a new subscription.
func (h *Handlers) createSub(c *gin.Context) {
	var req createSubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, ok := h.biller.GetPlanByProductID(req.ProductID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	sub, err := h.biller.CreateSub(plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// clearSubs clears all subscriptions.
func (h *Handlers) clearSubs(c *gin.Context) {
	if err := h.biller.ClearSubs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// setSubStatusRequest represents the request to set subscription status.
type setSubStatusRequest struct {
	Status int `json:"status"`
}

// setSubStatus updates the status of a subscription.
func (h *Handlers) setSubStatus(c *gin.Context) {
	receiptToken := c.Param("receiptToken")

	var req setSubStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, ok := h.biller.GetSubscription(receiptToken)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if err := h.biller.SetSubStatus(sub, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// renewSub renews a subscription.
func (h *Handlers) renewSub(c *gin.Context) {
	receiptToken := c.Param("receiptToken")

	sub, ok := h.biller.GetSubscription(receiptToken)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if err := h.biller.RenewSub(sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// cancelSub cancels a subscription.
func (h *Handlers) cancelSub(c *gin.Context) {
	receiptToken := c.Param("receiptToken")

	sub, ok := h.biller.GetSubscription(receiptToken)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if err := h.biller.CancelSub(sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// refundTransaction refunds a specific transaction.
func (h *Handlers) refundTransaction(c *gin.Context) {
	receiptToken := c.Param("receiptToken")
	transactionID := c.Param("transactionId")

	sub, ok := h.biller.GetSubscription(receiptToken)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	transactionMap := sub.TransactionMap()
	receiptInfo, ok := transactionMap[transactionID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	if err := h.biller.RefundTransaction(sub, *receiptInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// verifyReceiptRequest represents the request to verify a receipt.
type verifyReceiptRequest struct {
	ReceiptData string `json:"receipt-data"`
}

// verifyReceipt verifies a receipt and returns the response in Apple's format.
func (h *Handlers) verifyReceipt(c *gin.Context) {
	var req verifyReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.biller.VerifyReceipt(req.ReceiptData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, response)
}

// validNotificationTypes is the set of Apple S2S v2 notification types the mock accepts.
var validNotificationTypes = map[string]bool{
	"SUBSCRIBED":             true,
	"DID_RENEW":              true,
	"DID_FAIL_TO_RENEW":      true,
	"EXPIRED":                true,
	"GRACE_PERIOD_EXPIRED":   true,
	"CANCEL":                 true,
	"REFUND":                 true,
	"REVOKE":                 true,
	"PRICE_INCREASE":         true,
}

// sendNotification fires an Apple S2S-style notification to the configured WEBHOOK_URL.
// POST /subs/:receiptToken/notify/:notificationType
func (h *Handlers) sendNotification(c *gin.Context) {
	receiptToken := c.Param("receiptToken")
	notificationType := c.Param("notificationType")

	if !validNotificationTypes[notificationType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unknown notificationType",
			"valid": []string{"SUBSCRIBED", "DID_RENEW", "DID_FAIL_TO_RENEW", "EXPIRED",
				"GRACE_PERIOD_EXPIRED", "CANCEL", "REFUND", "REVOKE", "PRICE_INCREASE"},
		})
		return
	}

	if h.webhookURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WEBHOOK_URL not configured"})
		return
	}

	sub, ok := h.biller.GetSubscription(receiptToken)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if err := h.biller.SendNotification(sub, notificationType, h.webhookURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "sent",
		"notificationType": notificationType,
		"webhookURL":       h.webhookURL,
	})
}
