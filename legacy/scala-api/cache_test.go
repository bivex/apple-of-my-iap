package iap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}

	// Check that files were created
	subsFile := filepath.Join(tmpDir, "subscriptions.json")
	plansFile := filepath.Join(tmpDir, "plans.json")

	if _, err := os.Stat(subsFile); os.IsNotExist(err) {
		t.Error("subscriptions.json was not created")
	}
	if _, err := os.Stat(plansFile); os.IsNotExist(err) {
		t.Error("plans.json was not created")
	}
}

func TestCacheReadSubscriptionsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	subs, err := cache.ReadSubscriptions()
	if err != nil {
		t.Fatalf("ReadSubscriptions() error = %v", err)
	}

	if subs == nil {
		t.Fatal("ReadSubscriptions() returned nil")
	}
	if len(subs) != 0 {
		t.Errorf("ReadSubscriptions() len = %d, want 0", len(subs))
	}
}

func TestCacheWriteReadSubscriptions(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	// Create test subscriptions
	subs := map[string]*Subscription{
		"token1": {
			Status:       0,
			ReceiptsList: []ReceiptInfo{{TransactionID: "txn1"}},
			ReceiptToken: "token1",
			SubStatus:    SubscriptionStatusActive,
		},
		"token2": {
			Status:       0,
			ReceiptsList: []ReceiptInfo{{TransactionID: "txn2"}},
			ReceiptToken: "token2",
			SubStatus:    SubscriptionStatusActive,
		},
	}

	// Write
	err := cache.WriteSubscriptions(subs)
	if err != nil {
		t.Fatalf("WriteSubscriptions() error = %v", err)
	}

	// Create new cache instance to test file persistence
	cache2, _ := NewCache(tmpDir)
	readSubs, err := cache2.ReadSubscriptions()
	if err != nil {
		t.Fatalf("ReadSubscriptions() error = %v", err)
	}

	if len(readSubs) != 2 {
		t.Errorf("ReadSubscriptions() len = %d, want 2", len(readSubs))
	}

	if _, ok := readSubs["token1"]; !ok {
		t.Error("token1 not found in read subscriptions")
	}
	if _, ok := readSubs["token2"]; !ok {
		t.Error("token2 not found in read subscriptions")
	}
}

func TestCacheReadPlansEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	plans, err := cache.ReadPlans()
	if err != nil {
		t.Fatalf("ReadPlans() error = %v", err)
	}

	if plans == nil {
		t.Fatal("ReadPlans() returned nil")
	}
	if len(plans) != 0 {
		t.Errorf("ReadPlans() len = %d, want 0", len(plans))
	}
}

func TestCacheWriteReadPlans(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	// Write plans directly to file
	plansFile := filepath.Join(tmpDir, "plans.json")
	data := `[{"name":"test","description":"test plan","billInterval":1,"billIntervalUnit":"months","trialInterval":0,"trialIntervalUnit":"days","productId":"com.test"}]`
	err := os.WriteFile(plansFile, []byte(data), 0644)
	if err != nil {
		t.Fatalf("Failed to write plans file: %v", err)
	}

	// Read plans
	plans, err := cache.ReadPlans()
	if err != nil {
		t.Fatalf("ReadPlans() error = %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("ReadPlans() len = %d, want 1", len(plans))
	}

	if plans[0].ProductID != "com.test" {
		t.Errorf("ReadPlans()[0].ProductID = %s, want com.test", plans[0].ProductID)
	}
}
