package iap

import (
	"testing"
	"time"
)

func TestNewSubscription(t *testing.T) {
	now := time.Now()
	receiptToken := "test-token"
	receiptInfo := ReceiptInfo{
		OriginalPurchaseDate:  AppleTime{Time: now},
		OriginalTransactionID: "orig-txn",
		TransactionID:         "txn-123",
		PurchaseDate:          AppleTime{Time: now},
		ExpiresDate:           AppleTime{Time: now.Add(30 * 24 * time.Hour)},
		ProductID:             "com.example.product",
		IsTrialPeriod:         false,
		Quantity:              1,
	}

	sub := NewSubscription(receiptToken, receiptInfo)

	if sub.Status != ValidReceipt.Code {
		t.Errorf("NewSubscription() Status = %d, want %d", sub.Status, ValidReceipt.Code)
	}
	if sub.ReceiptToken != receiptToken {
		t.Errorf("NewSubscription() ReceiptToken = %s, want %s", sub.ReceiptToken, receiptToken)
	}
	if len(sub.ReceiptsList) != 1 {
		t.Errorf("NewSubscription() ReceiptsList len = %d, want 1", len(sub.ReceiptsList))
	}
	if sub.SubStatus != SubscriptionStatusActive {
		t.Errorf("NewSubscription() SubStatus = %s, want %s", sub.SubStatus, SubscriptionStatusActive)
	}
}

func TestSubscriptionOriginalReceiptInfo(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{
			{TransactionID: "new-txn", PurchaseDate: AppleTime{Time: now}},
			{TransactionID: "old-txn", PurchaseDate: AppleTime{Time: earlier}},
		},
	}

	orig := sub.OriginalReceiptInfo()
	if orig == nil {
		t.Fatal("OriginalReceiptInfo() returned nil")
	}
	if orig.TransactionID != "old-txn" {
		t.Errorf("OriginalReceiptInfo() TransactionID = %s, want old-txn", orig.TransactionID)
	}
}

func TestSubscriptionLatestReceiptInfo(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{
			{TransactionID: "new-txn", PurchaseDate: AppleTime{Time: now}},
			{TransactionID: "old-txn", PurchaseDate: AppleTime{Time: earlier}},
		},
	}

	latest := sub.LatestReceiptInfo()
	if latest == nil {
		t.Fatal("LatestReceiptInfo() returned nil")
	}
	if latest.TransactionID != "new-txn" {
		t.Errorf("LatestReceiptInfo() TransactionID = %s, want new-txn", latest.TransactionID)
	}
}

func TestSubscriptionAddReceipt(t *testing.T) {
	now := time.Now()
	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{
			{TransactionID: "txn1", PurchaseDate: AppleTime{Time: now.Add(-1 * time.Hour)}},
		},
		ReceiptToken: "token1",
		ReceiptTokenMap: map[string]string{
			"txn1": "token1",
		},
	}

	newReceipt := ReceiptInfo{
		TransactionID: "txn2",
		PurchaseDate:  AppleTime{Time: now},
	}
	newToken := "token2"

	updated := sub.AddReceipt(newReceipt, newToken)

	if len(updated.ReceiptsList) != 2 {
		t.Errorf("AddReceipt() ReceiptsList len = %d, want 2", len(updated.ReceiptsList))
	}
	if updated.ReceiptsList[0].TransactionID != "txn2" {
		t.Errorf("AddReceipt() first receipt TransactionID = %s, want txn2", updated.ReceiptsList[0].TransactionID)
	}
	if updated.ReceiptTokenMap["txn2"] != newToken {
		t.Errorf("AddReceipt() token for txn2 = %s, want %s", updated.ReceiptTokenMap["txn2"], newToken)
	}
}

func TestSubscriptionCancel(t *testing.T) {
	sub := &Subscription{
		SubStatus: SubscriptionStatusActive,
	}

	canceled := sub.Cancel()

	if canceled.SubStatus != SubscriptionStatusCancelled {
		t.Errorf("Cancel() SubStatus = %s, want %s", canceled.SubStatus, SubscriptionStatusCancelled)
	}
}

func TestSubscriptionRefund(t *testing.T) {
	now := time.Now()
	refundInfo := ReceiptInfo{
		TransactionID: "txn1",
		PurchaseDate:  AppleTime{Time: now},
	}

	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{refundInfo},
	}

	refunded := sub.Refund(refundInfo)

	if len(refunded.ReceiptsList) != 1 {
		t.Fatalf("Refund() ReceiptsList len = %d, want 1", len(refunded.ReceiptsList))
	}

	if refunded.ReceiptsList[0].CancellationDate == nil {
		t.Error("Refund() CancellationDate is nil")
	}
}

func TestSubscriptionTransactionMap(t *testing.T) {
	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{
			{TransactionID: "txn1"},
			{TransactionID: "txn2"},
		},
	}

	txnMap := sub.TransactionMap()

	if len(txnMap) != 2 {
		t.Errorf("TransactionMap() len = %d, want 2", len(txnMap))
	}
	if _, ok := txnMap["txn1"]; !ok {
		t.Error("TransactionMap() missing txn1")
	}
	if _, ok := txnMap["txn2"]; !ok {
		t.Error("TransactionMap() missing txn2")
	}
}

func TestSubscriptionEmpty(t *testing.T) {
	sub := &Subscription{
		ReceiptsList: []ReceiptInfo{},
	}

	if sub.OriginalReceiptInfo() != nil {
		t.Error("OriginalReceiptInfo() on empty list should return nil")
	}
	if sub.LatestReceiptInfo() != nil {
		t.Error("LatestReceiptInfo() on empty list should return nil")
	}
}
