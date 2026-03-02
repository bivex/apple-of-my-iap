package iap

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNewBiller(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	biller := NewBiller(cache)
	if biller == nil {
		t.Fatal("NewBiller() returned nil")
	}

	if err := biller.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
}

func TestBillerCreateSub(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		Description:       "test plan",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		TrialInterval:     0,
		TrialIntervalUnit: "days",
		ProductID:         "com.test.plan",
	}

	sub, err := biller.CreateSub(plan)
	if err != nil {
		t.Fatalf("CreateSub() error = %v", err)
	}

	if sub == nil {
		t.Fatal("CreateSub() returned nil")
	}

	if sub.ReceiptToken == "" {
		t.Error("CreateSub() ReceiptToken is empty")
	}

	if len(sub.ReceiptsList) != 1 {
		t.Errorf("CreateSub() ReceiptsList len = %d, want 1", len(sub.ReceiptsList))
	}

	if sub.SubStatus != SubscriptionStatusActive {
		t.Errorf("CreateSub() SubStatus = %s, want %s", sub.SubStatus, SubscriptionStatusActive)
	}
}

func TestBillerGetSubscription(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller.CreateSub(plan)

	// Get by token
	found, ok := biller.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Error("GetSubscription() returned ok=false")
	}
	if found == nil {
		t.Fatal("GetSubscription() returned nil subscription")
	}

	if found.ReceiptToken != sub.ReceiptToken {
		t.Errorf("GetSubscription() token = %s, want %s", found.ReceiptToken, sub.ReceiptToken)
	}
}

func TestBillerRenewSub(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	// Create a plan file first to ensure the plan is loaded
	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}
	plans := []*Plan{plan}
	plansData, _ := json.Marshal(plans)
	plansFile := tmpDir + "/plans.json"
	os.WriteFile(plansFile, plansData, 0644)

	// Restart biller to load plans
	biller2 := NewBiller(cache)
	biller2.Start()

	sub, _ := biller2.CreateSub(plan)

	err := biller2.RenewSub(sub)
	if err != nil {
		t.Fatalf("RenewSub() error = %v", err)
	}

	// Get updated subscription
	updated, ok := biller2.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Fatal("GetSubscription() after renewal returned ok=false")
	}

	if len(updated.ReceiptsList) != 2 {
		t.Errorf("RenewSub() ReceiptsList len = %d, want 2", len(updated.ReceiptsList))
	}
}

func TestBillerCancelSub(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller.CreateSub(plan)

	err := biller.CancelSub(sub)
	if err != nil {
		t.Fatalf("CancelSub() error = %v", err)
	}

	// Get updated subscription
	updated, ok := biller.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Fatal("GetSubscription() after cancel returned ok=false")
	}

	if updated.SubStatus != SubscriptionStatusCancelled {
		t.Errorf("CancelSub() SubStatus = %s, want %s", updated.SubStatus, SubscriptionStatusCancelled)
	}
}

func TestBillerRefundTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller.CreateSub(plan)

	err := biller.RefundTransaction(sub, sub.ReceiptsList[0])
	if err != nil {
		t.Fatalf("RefundTransaction() error = %v", err)
	}

	// Get updated subscription
	updated, ok := biller.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Fatal("GetSubscription() after refund returned ok=false")
	}

	if updated.ReceiptsList[0].CancellationDate == nil {
		t.Error("RefundTransaction() CancellationDate is nil")
	}
}

func TestBillerClearSubs(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	biller.CreateSub(plan)
	biller.CreateSub(plan)

	err := biller.ClearSubs()
	if err != nil {
		t.Fatalf("ClearSubs() error = %v", err)
	}

	subs := biller.GetSubscriptions()
	if len(subs) != 0 {
		t.Errorf("ClearSubs() remaining subscriptions = %d, want 0", len(subs))
	}
}

func TestBillerVerifyReceipt(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller.CreateSub(plan)

	// Verify valid receipt
	response, err := biller.VerifyReceipt(sub.ReceiptToken)
	if err != nil {
		t.Fatalf("VerifyReceipt() error = %v", err)
	}

	if response == "" {
		t.Error("VerifyReceipt() returned empty string")
	}

	// Check that status is 0 (valid)
	if !contains(response, `"status": 0`) {
		t.Errorf("VerifyReceipt() should contain status 0, got: %s", response)
	}

	// Verify invalid receipt
	response, err = biller.VerifyReceipt("invalid-token")
	if err != nil {
		t.Fatalf("VerifyReceipt() with invalid token error = %v", err)
	}

	// Should have status 21003 (UnauthorizedReceipt)
	if !contains(response, `"status": 21003`) {
		t.Errorf("VerifyReceipt() with invalid token should contain status 21003, got: %s", response)
	}
}

func TestBillerSetSubStatus(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)
	biller := NewBiller(cache)
	biller.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller.CreateSub(plan)

	err := biller.SetSubStatus(sub, 21006) // SubscriptionExpired
	if err != nil {
		t.Fatalf("SetSubStatus() error = %v", err)
	}

	// Get updated subscription
	updated, ok := biller.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Fatal("GetSubscription() after SetSubStatus returned ok=false")
	}

	if updated.Status != 21006 {
		t.Errorf("SetSubStatus() Status = %d, want 21006", updated.Status)
	}
}

func TestBillerPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cache, _ := NewCache(tmpDir)

	// Create first biller and add subscription
	biller1 := NewBiller(cache)
	biller1.Start()

	plan := &Plan{
		Name:              "test",
		BillInterval:      1,
		BillIntervalUnit:  "months",
		ProductID:         "com.test.plan",
	}

	sub, _ := biller1.CreateSub(plan)

	// Shutdown first biller
	biller1.Shutdown()

	// Create second biller and verify subscription is loaded
	biller2 := NewBiller(cache)
	biller2.Start()

	loaded, ok := biller2.GetSubscription(sub.ReceiptToken)
	if !ok {
		t.Error("Persistence: GetSubscription() failed to find loaded subscription")
	}

	if loaded.ReceiptToken != sub.ReceiptToken {
		t.Errorf("Persistence: loaded token = %s, want %s", loaded.ReceiptToken, sub.ReceiptToken)
	}
}
