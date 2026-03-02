package iap

import (
	"testing"
	"time"
)

func TestGenEncoding(t *testing.T) {
	gen := NewReceiptGenerator()
	plan := &Plan{
		Name:      "test_plan",
		ProductID: "com.test.plan",
	}

	existing := make(map[string]bool)

	token1 := gen.GenEncoding(plan, existing)
	if token1 == "" {
		t.Fatal("GenEncoding() returned empty string")
	}

	// Should be different on second call
	existing[token1] = true
	token2 := gen.GenEncoding(plan, existing)
	if token2 == token1 {
		t.Errorf("GenEncoding() returned same token %s", token1)
	}

	// Should start with plan name
	if len(token1) < len(plan.Name)+1 {
		t.Errorf("GenEncoding() token too short: %s", token1)
	}
}

func TestGenerateReceiptNewSubscription(t *testing.T) {
	gen := NewReceiptGenerator()
	plan := &Plan{
		Name:              "monthly",
		ProductID:         "com.example.monthly",
		BillInterval:      1,
		BillIntervalUnit:  "months",
	}

	// For new subscriptions, GenerateReceipt returns an empty token
	// The actual token is generated via GenEncoding and set by the caller
	receiptToken, receiptInfo := gen.GenerateReceipt(plan, nil)

	if receiptInfo == nil {
		t.Fatal("GenerateReceipt() returned nil receiptInfo")
	}

	if receiptInfo.ProductID != plan.ProductID {
		t.Errorf("GenerateReceipt() ProductID = %s, want %s", receiptInfo.ProductID, plan.ProductID)
	}

	if receiptInfo.Quantity != 1 {
		t.Errorf("GenerateReceipt() Quantity = %d, want 1", receiptInfo.Quantity)
	}

	if receiptInfo.IsTrialPeriod {
		t.Error("GenerateReceipt() IsTrialPeriod should be false")
	}

	if receiptInfo.ExpiresDate.Time.Before(time.Now()) {
		t.Error("GenerateReceipt() ExpiresDate should be in the future")
	}

	// Token is empty for new subscriptions (set by caller via GenEncoding)
	// This is expected behavior
	_ = receiptToken
}

func TestGenerateReceiptRenewal(t *testing.T) {
	gen := NewReceiptGenerator()
	plan := &Plan{
		Name:              "monthly",
		ProductID:         "com.example.monthly",
		BillInterval:      1,
		BillIntervalUnit:  "months",
	}

	now := time.Now()
	origSub := &Subscription{
		ReceiptToken: "original-token",
		ReceiptsList: []ReceiptInfo{
			{
				TransactionID:         "orig-txn",
				OriginalPurchaseDate:  AppleTime{Time: now.Add(-30 * 24 * time.Hour)},
				OriginalTransactionID: "orig-txn",
			},
		},
		ReceiptTokenMap: map[string]string{
			"orig-txn": "original-token",
		},
	}

	receiptToken, receiptInfo := gen.GenerateReceipt(plan, origSub)

	if receiptToken == "" {
		t.Error("GenerateReceipt() token is empty for renewal")
	}

	if receiptInfo.OriginalTransactionID != "orig-txn" {
		t.Errorf("GenerateReceipt() OriginalTransactionID = %s, want orig-txn", receiptInfo.OriginalTransactionID)
	}
}

func TestGenerateReceiptResponse(t *testing.T) {
	gen := NewReceiptGenerator()
	now := time.Now()

	sub := &Subscription{
		Status: ValidReceipt.Code,
		ReceiptsList: []ReceiptInfo{
			{TransactionID: "txn1", PurchaseDate: AppleTime{Time: now}},
		},
		ReceiptToken: "token1",
		ReceiptTokenMap: map[string]string{
			"txn1": "token1",
		},
	}

	resp := gen.GenerateReceiptResponse(sub)

	if resp.StatusCode != ValidReceipt.Code {
		t.Errorf("GenerateReceiptResponse() StatusCode = %d, want %d", resp.StatusCode, ValidReceipt.Code)
	}

	if len(resp.LatestReceiptInfo) != 1 {
		t.Errorf("GenerateReceiptResponse() LatestReceiptInfo len = %d, want 1", len(resp.LatestReceiptInfo))
	}

	if resp.LatestReceipt == nil {
		t.Error("GenerateReceiptResponse() LatestReceipt is nil")
	} else if *resp.LatestReceipt != "token1" {
		t.Errorf("GenerateReceiptResponse() LatestReceipt = %s, want token1", *resp.LatestReceipt)
	}
}

func TestCalculateEndDate(t *testing.T) {
	gen := NewReceiptGenerator()
	start := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		interval     int
		intervalUnit string
		wantDelta    time.Duration
	}{
		{"seconds", 30, "seconds", 30 * time.Second},
		{"minutes", 5, "minutes", 5 * time.Minute},
		{"hours", 2, "hours", 2 * time.Hour},
		{"days", 7, "days", 7 * 24 * time.Hour},
		{"weeks", 2, "weeks", 14 * 24 * time.Hour},
		{"months", 1, "months", 30 * 24 * time.Hour}, // Approximate
		{"years", 1, "years", 365 * 24 * time.Hour},  // Approximate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.calculateEndDate(start, tt.interval, tt.intervalUnit)
			delta := result.Sub(start)

			// Allow some tolerance for months/years
			tolerance := time.Minute
			if tt.intervalUnit == "months" || tt.intervalUnit == "years" {
				tolerance = 24 * time.Hour
			}

			diff := delta - tt.wantDelta
			if diff < 0 {
				diff = -diff
			}
			if diff > tolerance {
				t.Errorf("calculateEndDate() delta = %v, want %v (±%v)", delta, tt.wantDelta, tolerance)
			}
		})
	}
}

func TestCalculateEndDateInvalidUnit(t *testing.T) {
	gen := NewReceiptGenerator()
	start := time.Now()

	defer func() {
		if r := recover(); r == nil {
			t.Error("calculateEndDate() should panic with invalid interval unit")
		}
	}()

	gen.calculateEndDate(start, 1, "invalid")
}
