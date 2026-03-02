package iap

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReceiptGenerator handles generation of mock receipts and receipt tokens.
type ReceiptGenerator struct{}

// NewReceiptGenerator creates a new ReceiptGenerator.
func NewReceiptGenerator() *ReceiptGenerator {
	return &ReceiptGenerator{}
}

// GenEncoding generates a unique receipt token for a plan.
func (g *ReceiptGenerator) GenEncoding(plan *Plan, existingEncodings map[string]bool) string {
	for {
		id := uuid.New().String()
		parts := id[:8]
		first4 := parts[0:4]
		second4 := parts[4:8]
		receipt := fmt.Sprintf("%s_%s-%s", plan.Name, first4, second4)

		if !existingEncodings[receipt] {
			return receipt
		}
	}
}

// GenerateReceipt creates a new receipt info and token for a plan.
func (g *ReceiptGenerator) GenerateReceipt(plan *Plan, sub *Subscription) (string, *ReceiptInfo) {
	purchaseDateTime := time.Now()
	purchaseDate := AppleTime{Time: purchaseDateTime}
	productID := plan.ProductID
	transactionID := fmt.Sprintf("%s-%d", productID, purchaseDateTime.UnixMilli())

	expiresDate := g.calculateEndDate(purchaseDateTime, plan.BillInterval, plan.BillIntervalUnit)

	var origPurchaseDate AppleTime
	var origTransID string
	var receiptToken string

	if sub == nil {
		origPurchaseDate = purchaseDate
		origTransID = transactionID
		receiptToken = ""
	} else {
		orig := sub.OriginalReceiptInfo()
		origPurchaseDate = orig.OriginalPurchaseDate
		origTransID = orig.OriginalTransactionID

		origToken, ok := sub.ReceiptTokenMap[orig.TransactionID]
		if !ok {
			origToken = "ERROR_no_receipt_token_found"
		}
		receiptToken = fmt.Sprintf("%s-%03d", origToken, len(sub.ReceiptsList))
	}

	return receiptToken, &ReceiptInfo{
		OriginalPurchaseDate:  origPurchaseDate,
		OriginalTransactionID: origTransID,
		TransactionID:         transactionID,
		PurchaseDate:          purchaseDate,
		ExpiresDate:           AppleTime{Time: expiresDate},
		ProductID:             productID,
		CancellationDate:      nil,
		IsTrialPeriod:         false,
		IsInIntroOfferPeriod:  nil,
		Quantity:              1,
	}
}

// GenerateReceiptResponse creates a ReceiptResponse from a subscription.
func (g *ReceiptGenerator) GenerateReceiptResponse(sub *Subscription) *ReceiptResponse {
	latestToken := sub.LatestReceiptToken()

	return &ReceiptResponse{
		LatestReceipt:     latestToken,
		LatestReceiptInfo: sub.ReceiptsList,
		StatusCode:        sub.Status,
	}
}

// calculateEndDate calculates the expiration date based on the billing interval.
func (g *ReceiptGenerator) calculateEndDate(startDate time.Time, interval int, intervalUnit string) time.Time {
	switch intervalUnit {
	case "seconds":
		return startDate.Add(time.Duration(interval) * time.Second)
	case "minutes":
		return startDate.Add(time.Duration(interval) * time.Minute)
	case "hours":
		return startDate.Add(time.Duration(interval) * time.Hour)
	case "days":
		return startDate.AddDate(0, 0, interval)
	case "weeks":
		return startDate.AddDate(0, 0, interval*7)
	case "months":
		return startDate.AddDate(0, interval, 0)
	case "years":
		return startDate.AddDate(interval, 0, 0)
	default:
		panic(fmt.Sprintf("unknown interval unit: %s", intervalUnit))
	}
}
