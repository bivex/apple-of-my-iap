package iap

import (
	"testing"
	"time"
)

func TestRenderReceiptResponse(t *testing.T) {
	renderer := NewRenderer()
	now := time.Date(2026, 3, 2, 14, 46, 25, 0, time.UTC)

	resp := &ReceiptResponse{
		StatusCode: 0,
		LatestReceipt: stringPtr("test-receipt-token"),
		LatestReceiptInfo: []ReceiptInfo{
			{
				Quantity:              1,
				ProductID:             "com.example.product",
				TransactionID:         "txn-123",
				OriginalTransactionID: "orig-txn-123",
				PurchaseDate:          AppleTime{Time: now},
				OriginalPurchaseDate:  AppleTime{Time: now.Add(-30 * 24 * time.Hour)},
				ExpiresDate:           AppleTime{Time: now.Add(30 * 24 * time.Hour)},
				IsTrialPeriod:         false,
				IsInIntroOfferPeriod:  boolPtr(false),
			},
		},
	}

	output, err := renderer.RenderReceiptResponse(resp)
	if err != nil {
		t.Fatalf("RenderReceiptResponse() error = %v", err)
	}

	if output == "" {
		t.Error("RenderReceiptResponse() returned empty string")
	}

	// Check for expected fields
	expectedFields := []string{
		`"status": 0`,
		`"latest_receipt"`,
		`"test-receipt-token"`,
		`"latest_receipt_info"`,
		`"product_id"`,
		`"com.example.product"`,
		`"quantity"`,
		`"is_trial_period"`,
	}

	for _, field := range expectedFields {
		if !contains(output, field) {
			t.Errorf("RenderReceiptResponse() missing field: %s", field)
		}
	}
}

func TestRenderReceiptInfo(t *testing.T) {
	renderer := NewRenderer()
	now := time.Date(2026, 3, 2, 14, 46, 25, 0, time.UTC)

	info := ReceiptInfo{
		Quantity:              1,
		ProductID:             "com.example.product",
		TransactionID:         "txn-123",
		OriginalTransactionID: "orig-txn-123",
		PurchaseDate:          AppleTime{Time: now},
		OriginalPurchaseDate:  AppleTime{Time: now.Add(-30 * 24 * time.Hour)},
		ExpiresDate:           AppleTime{Time: now.Add(30 * 24 * time.Hour)},
		IsTrialPeriod:         false,
		IsInIntroOfferPeriod:  nil,
		CancellationDate:      nil,
	}

	rendered := renderer.renderReceiptInfo(info)

	// Check that quantity is a string
	if q, ok := rendered["quantity"].(string); !ok || q != "1" {
		t.Errorf("renderReceiptInfo() quantity should be string \"1\", got %v", rendered["quantity"])
	}

	// Check that is_trial_period is a string
	if itp, ok := rendered["is_trial_period"].(string); !ok || itp != "false" {
		t.Errorf("renderReceiptInfo() is_trial_period should be string \"false\", got %v", rendered["is_trial_period"])
	}

	// Check date formats
	if pd, ok := rendered["purchase_date"].(string); !ok || pd == "" {
		t.Error("renderReceiptInfo() purchase_date is missing or not a string")
	}

	// Check that _ms fields exist and are numbers
	if _, ok := rendered["purchase_date_ms"].(int64); !ok {
		t.Error("renderReceiptInfo() purchase_date_ms is missing or not an int64")
	}
}

func TestRenderReceiptInfoWithCancellation(t *testing.T) {
	renderer := NewRenderer()
	now := time.Now()

	info := ReceiptInfo{
		Quantity:         1,
		ProductID:        "com.example.product",
		TransactionID:    "txn-123",
		PurchaseDate:     AppleTime{Time: now},
		ExpiresDate:      AppleTime{Time: now.Add(30 * 24 * time.Hour)},
		IsTrialPeriod:    false,
		CancellationDate: &AppleTime{Time: now},
	}

	rendered := renderer.renderReceiptInfo(info)

	if _, ok := rendered["cancellation_date"].(string); !ok {
		t.Error("renderReceiptInfo() cancellation_date should be present")
	}
}

func TestRenderReceiptInfoWithIntroOffer(t *testing.T) {
	renderer := NewRenderer()
	now := time.Now()

	trueVal := true
	info := ReceiptInfo{
		Quantity:             1,
		ProductID:            "com.example.product",
		TransactionID:        "txn-123",
		PurchaseDate:         AppleTime{Time: now},
		ExpiresDate:          AppleTime{Time: now.Add(30 * 24 * time.Hour)},
		IsTrialPeriod:        false,
		IsInIntroOfferPeriod: &trueVal,
	}

	rendered := renderer.renderReceiptInfo(info)

	if _, ok := rendered["is_in_intro_offer_period"].(string); !ok {
		t.Error("renderReceiptInfo() is_in_intro_offer_period should be present")
	}
}

func TestFormatDate(t *testing.T) {
	tm := time.Date(2026, 3, 2, 14, 46, 25, 123456789, time.UTC)
	result := formatDate(tm)

	expected := "2026-03-02 14:46:25"
	if result != expected {
		t.Errorf("formatDate() = %s, want %s", result, expected)
	}
}

func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsIn(s, substr))
}

func containsIn(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
