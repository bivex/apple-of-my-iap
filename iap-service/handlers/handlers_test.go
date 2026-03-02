package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	iap "github.com/meetup/iap-api"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *iap.Biller) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	cache, _ := iap.NewCache(tmpDir)
	biller := iap.NewBiller(cache)
	biller.Start()

	router := gin.New()
	h := New(biller)
	h.RegisterRoutes(router)

	return router, biller
}

func TestIndexHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("indexHandler status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["message"] != "Apple IAP Mock Service" {
		t.Errorf("indexHandler message = %v, want 'Apple IAP Mock Service'", response["message"])
	}
}

func TestGetPlansHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/plans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getPlans status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGetSubsHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/subs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getSubs status = %d, want %d", w.Code, http.StatusOK)
	}

	var subs []iap.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &subs)
	if err != nil {
		t.Fatalf("Failed to parse subscriptions: %v", err)
	}

	if len(subs) != 0 {
		t.Errorf("getSubs len = %d, want 0", len(subs))
	}
}

func TestClearSubsHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/subs/clear", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("clearSubs status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestVerifyReceiptHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	reqBody := `{"receipt-data":"test-token-123"}`
	req := httptest.NewRequest("POST", "/verifyReceipt", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("verifyReceipt status = %d, want %d", w.Code, http.StatusOK)
	}

	responseBody := w.Body.String()
	if responseBody == "" {
		t.Error("verifyReceipt response is empty")
	}

	if !strings.Contains(responseBody, `"status"`) {
		t.Errorf("verifyReceipt response should contain status field: %s", responseBody)
	}
}

func TestVerifyReceiptHandlerInvalidJSON(t *testing.T) {
	router, _ := setupTestRouter(t)

	reqBody := `invalid json`
	req := httptest.NewRequest("POST", "/verifyReceipt", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("verifyReceipt with invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRenewSubHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/subs/nonexistent/renew", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("renewSub with invalid token status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCancelSubHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/subs/nonexistent/cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("cancelSub with invalid token status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSetSubStatusHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	reqBody := `{"status":21006}`
	req := httptest.NewRequest("POST", "/subs/nonexistent", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("setSubStatus with invalid token status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRefundTransactionHandler(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/subs/nonexistent/refund/txn-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("refundTransaction with invalid token status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
