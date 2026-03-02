package iap

import (
	"encoding/json"
	"fmt"
	"time"
)

// AppleDateFormat is the format used by Apple for receipt dates: "yyyy-MM-dd HH:mm:ss 'Etc/GMT'"
// Note: The actual Apple format is "yyyy-MM-dd HH:mm:ss 'Etc/GMT'" but for parsing
// we use a simpler format without the timezone string since times are already in UTC
const AppleDateFormat = "2006-01-02 15:04:05 MST"

// AppleTime is a custom time type that marshals/unmarshals in Apple's date format.
type AppleTime struct {
	time.Time
}

// MarshalJSON implements json.Marshaler for AppleTime.
func (t AppleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(AppleDateFormat))
}

// UnmarshalJSON implements json.Unmarshaler for AppleTime.
func (t *AppleTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(AppleDateFormat, s)
	if err != nil {
		return fmt.Errorf("failed to parse apple time: %w", err)
	}
	t.Time = parsed
	return nil
}

// ReceiptInfo contains information about a single receipt/transaction.
type ReceiptInfo struct {
	OriginalPurchaseDate   AppleTime `json:"original_purchase_date"`
	OriginalTransactionID  string    `json:"original_transaction_id"`
	TransactionID          string    `json:"transaction_id"`
	PurchaseDate           AppleTime `json:"purchase_date"`
	ExpiresDate            AppleTime `json:"expires_date"`
	ProductID              string    `json:"product_id"`
	IsTrialPeriod          bool      `json:"is_trial_period"`
	IsInIntroOfferPeriod   *bool     `json:"is_in_intro_offer_period,omitempty"`
	CancellationDate       *AppleTime `json:"cancellation_date,omitempty"`
	Quantity               int       `json:"quantity"`
}

// ReceiptResponse is the response structure for Apple's verifyReceipt endpoint.
type ReceiptResponse struct {
	LatestReceipt      *string       `json:"latest_receipt,omitempty"`
	LatestReceiptInfo  []ReceiptInfo `json:"latest_receipt_info"`
	StatusCode         int           `json:"status"`
}

// LatestInfo returns the most recent receipt info by purchase date.
func (r *ReceiptResponse) LatestInfo() *ReceiptInfo {
	if len(r.LatestReceiptInfo) == 0 {
		return nil
	}
	latest := r.LatestReceiptInfo[0]
	for i := 1; i < len(r.LatestReceiptInfo); i++ {
		if r.LatestReceiptInfo[i].PurchaseDate.After(latest.PurchaseDate.Time) {
			latest = r.LatestReceiptInfo[i]
		}
	}
	return &latest
}
