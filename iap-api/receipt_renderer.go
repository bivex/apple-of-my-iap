package iap

import (
	"encoding/json"
	"fmt"
	"time"
)

// Renderer handles rendering receipts in Apple's JSON format.
type Renderer struct{}

// NewRenderer creates a new Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderReceiptResponse renders a ReceiptResponse in Apple's JSON format.
func (r *Renderer) RenderReceiptResponse(resp *ReceiptResponse) (string, error) {
	latestReceiptInfo := make([]map[string]interface{}, len(resp.LatestReceiptInfo))
	for i, info := range resp.LatestReceiptInfo {
		latestReceiptInfo[i] = r.renderReceiptInfo(info)
	}

	result := map[string]interface{}{
		"status":              resp.StatusCode,
		"latest_receipt_info": latestReceiptInfo,
	}

	if resp.LatestReceipt != nil {
		result["latest_receipt"] = *resp.LatestReceipt
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal receipt response: %w", err)
	}

	return string(data), nil
}

// renderReceiptInfo converts a ReceiptInfo to Apple's JSON format.
func (r *Renderer) renderReceiptInfo(info ReceiptInfo) map[string]interface{} {
	result := map[string]interface{}{
		"quantity":                      "1",
		"product_id":                    info.ProductID,
		"transaction_id":                info.TransactionID,
		"original_transaction_id":       info.OriginalTransactionID,
		"purchase_date":                 formatDate(info.PurchaseDate.Time),
		"purchase_date_ms":              info.PurchaseDate.Time.UnixMilli(),
		"original_purchase_date":        formatDate(info.OriginalPurchaseDate.Time),
		"original_purchase_date_ms":     info.OriginalPurchaseDate.Time.UnixMilli(),
		"expires_date":                  formatDate(info.ExpiresDate.Time),
		"expires_date_ms":               info.ExpiresDate.Time.UnixMilli(),
		"is_trial_period":               fmt.Sprintf("%t", info.IsTrialPeriod),
	}

	if info.IsInIntroOfferPeriod != nil {
		result["is_in_intro_offer_period"] = fmt.Sprintf("%t", *info.IsInIntroOfferPeriod)
	}

	if info.CancellationDate != nil {
		result["cancellation_date"] = formatDate(info.CancellationDate.Time)
	}

	return result
}

// formatDate formats a time for Apple's date format.
func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}
