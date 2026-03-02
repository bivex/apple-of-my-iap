package iap

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAppleTimeMarshalJSON(t *testing.T) {
	tm := time.Date(2026, 3, 2, 14, 46, 25, 0, time.UTC)
	at := AppleTime{Time: tm}

	data, err := json.Marshal(at)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// AppleTime marshals to a format with timezone
	result := string(data)
	if !contains(result, "2026-03-02 14:46:25") {
		t.Errorf("MarshalJSON() = %s, should contain 2026-03-02 14:46:25", result)
	}
}

func TestAppleTimeUnmarshalJSON(t *testing.T) {
	// Parse time with timezone abbreviation (UTC)
	data := []byte("\"2026-03-02 14:46:25 UTC\"")
	var at AppleTime

	err := json.Unmarshal(data, &at)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	expected := time.Date(2026, 3, 2, 14, 46, 25, 0, time.UTC)
	if !at.Time.Equal(expected) {
		t.Errorf("UnmarshalJSON() = %v, want %v", at.Time, expected)
	}
}

func TestAppleTimeUnmarshalJSONInvalid(t *testing.T) {
	data := []byte("\"invalid\"")
	var at AppleTime

	err := json.Unmarshal(data, &at)
	if err == nil {
		t.Error("UnmarshalJSON() expected error for invalid format, got nil")
	}
}

func TestReceiptInfoLatestInfo(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	resp := ReceiptResponse{
		LatestReceiptInfo: []ReceiptInfo{
			{PurchaseDate: AppleTime{Time: earlier}},
			{PurchaseDate: AppleTime{Time: now}},
		},
	}

	latest := resp.LatestInfo()
	if latest == nil {
		t.Fatal("LatestInfo() returned nil")
	}

	if !latest.PurchaseDate.Time.Equal(now) {
		t.Errorf("LatestInfo() = %v, want %v", latest.PurchaseDate.Time, now)
	}
}

func TestReceiptInfoLatestInfoEmpty(t *testing.T) {
	resp := &ReceiptResponse{
		LatestReceiptInfo: []ReceiptInfo{},
	}

	latest := resp.LatestInfo()
	if latest != nil {
		t.Errorf("LatestInfo() on empty list = %v, want nil", latest)
	}
}

func TestReceiptInfoJSONRoundTrip(t *testing.T) {
	original := ReceiptInfo{
		OriginalPurchaseDate:  AppleTime{Time: time.Now()},
		OriginalTransactionID: "orig-123",
		TransactionID:         "txn-456",
		PurchaseDate:          AppleTime{Time: time.Now()},
		ExpiresDate:           AppleTime{Time: time.Now().Add(30 * 24 * time.Hour)},
		ProductID:             "com.example.product",
		IsTrialPeriod:         true,
		IsInIntroOfferPeriod:  boolPtr(true),
		CancellationDate:      nil,
		Quantity:              1,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded ReceiptInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.ProductID != original.ProductID {
		t.Errorf("ProductID = %s, want %s", decoded.ProductID, original.ProductID)
	}
	if decoded.TransactionID != original.TransactionID {
		t.Errorf("TransactionID = %s, want %s", decoded.TransactionID, original.TransactionID)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
