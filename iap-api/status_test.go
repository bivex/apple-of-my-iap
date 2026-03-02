package iap

import (
	"testing"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected Status
	}{
		{"Valid Receipt", 0, ValidReceipt},
		{"Bad Envelope", 21000, BadEnvelope},
		{"Bad Receipt", 21002, BadReceipt},
		{"Unauthorized", 21003, UnauthorizedReceipt},
		{"Shared Secret Mismatch", 21004, SharedSecretMismatch},
		{"Server Unavailable", 21005, ServerUnavailable},
		{"Subscription Expired", 21006, SubscriptionExpired},
		{"Test to Production", 21007, TestToProduction},
		{"Production to Test", 21008, ProductionToTest},
		{"Unknown Code", 99999, Status{Code: 99999, Description: "Unknown status.", Name: "Unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatus(tt.code)
			if result.Code != tt.expected.Code {
				t.Errorf("GetStatus(%d) code = %d, want %d", tt.code, result.Code, tt.expected.Code)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("GetStatus(%d) name = %s, want %s", tt.code, result.Name, tt.expected.Name)
			}
		})
	}
}

func TestStatusFields(t *testing.T) {
	status := ValidReceipt
	if status.Code != 0 {
		t.Errorf("ValidReceipt.Code = %d, want 0", status.Code)
	}
	if status.Description != "Valid receipt." {
		t.Errorf("ValidReceipt.Description = %s, want 'Valid receipt.'", status.Description)
	}
	if status.Name != "ValidReceipt" {
		t.Errorf("ValidReceipt.Name = %s, want 'ValidReceipt'", status.Name)
	}
}
