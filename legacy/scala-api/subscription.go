package iap

import (
	"time"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

// Subscription represents an IAP subscription with its receipt history.
type Subscription struct {
	Status           int                       `json:"status"`
	ReceiptsList     []ReceiptInfo             `json:"receiptsList"`
	ReceiptToken     string                    `json:"receiptToken"`
	ReceiptTokenMap  map[string]string         `json:"receiptTokenMap"` // transactionID -> receiptToken
	Auto             bool                      `json:"auto"`
	SubStatus        SubscriptionStatus        `json:"subStatus"`
}

// NewSubscription creates a new subscription with a single receipt.
func NewSubscription(receiptToken string, originalReceiptInfo ReceiptInfo) *Subscription {
	return &Subscription{
		Status: ValidReceipt.Code,
		ReceiptsList: []ReceiptInfo{originalReceiptInfo},
		ReceiptToken: receiptToken,
		ReceiptTokenMap: map[string]string{
			originalReceiptInfo.TransactionID: receiptToken,
		},
		Auto:      false,
		SubStatus: SubscriptionStatusActive,
	}
}

// OriginalReceiptInfo returns the original (first) receipt info.
func (s *Subscription) OriginalReceiptInfo() *ReceiptInfo {
	if len(s.ReceiptsList) == 0 {
		return nil
	}
	return &s.ReceiptsList[len(s.ReceiptsList)-1]
}

// LatestReceiptInfo returns the latest (most recent) receipt info.
func (s *Subscription) LatestReceiptInfo() *ReceiptInfo {
	if len(s.ReceiptsList) == 0 {
		return nil
	}
	return &s.ReceiptsList[0]
}

// LatestReceiptToken returns the receipt token for the latest receipt.
func (s *Subscription) LatestReceiptToken() *string {
	latest := s.LatestReceiptInfo()
	if latest == nil {
		return nil
	}
	token, ok := s.ReceiptTokenMap[latest.TransactionID]
	if !ok {
		return nil
	}
	return &token
}

// TransactionMap returns a map of transaction IDs to receipt info.
func (s *Subscription) TransactionMap() map[string]*ReceiptInfo {
	result := make(map[string]*ReceiptInfo, len(s.ReceiptsList))
	for i := range s.ReceiptsList {
		result[s.ReceiptsList[i].TransactionID] = &s.ReceiptsList[i]
	}
	return result
}

// AddReceipt adds a new receipt to the subscription.
func (s *Subscription) AddReceipt(receipt ReceiptInfo, newReceiptToken string) *Subscription {
	newList := make([]ReceiptInfo, len(s.ReceiptsList)+1)
	newList[0] = receipt
	copy(newList[1:], s.ReceiptsList)

	newMap := make(map[string]string, len(s.ReceiptTokenMap)+1)
	for k, v := range s.ReceiptTokenMap {
		newMap[k] = v
	}
	newMap[receipt.TransactionID] = newReceiptToken

	return &Subscription{
		Status:          s.Status,
		ReceiptsList:    newList,
		ReceiptToken:    s.ReceiptToken,
		ReceiptTokenMap: newMap,
		Auto:            s.Auto,
		SubStatus:       s.SubStatus,
	}
}

// Cancel marks the subscription as cancelled.
func (s *Subscription) Cancel() *Subscription {
	return &Subscription{
		Status:          s.Status,
		ReceiptsList:    s.ReceiptsList,
		ReceiptToken:    s.ReceiptToken,
		ReceiptTokenMap: s.ReceiptTokenMap,
		Auto:            s.Auto,
		SubStatus:       SubscriptionStatusCancelled,
	}
}

// Refund marks a transaction as refunded by setting its cancellation date.
func (s *Subscription) Refund(receiptInfo ReceiptInfo) *Subscription {
	newList := make([]ReceiptInfo, len(s.ReceiptsList))
	now := AppleTime{Time: time.Now()}
	for i, r := range s.ReceiptsList {
		if r.TransactionID == receiptInfo.TransactionID {
			newList[i] = ReceiptInfo{
				OriginalPurchaseDate:   r.OriginalPurchaseDate,
				OriginalTransactionID:  r.OriginalTransactionID,
				TransactionID:          r.TransactionID,
				PurchaseDate:           r.PurchaseDate,
				ExpiresDate:            r.ExpiresDate,
				ProductID:              r.ProductID,
				IsTrialPeriod:          r.IsTrialPeriod,
				IsInIntroOfferPeriod:   r.IsInIntroOfferPeriod,
				CancellationDate:       &now,
				Quantity:               r.Quantity,
			}
		} else {
			newList[i] = r
		}
	}
	return &Subscription{
		Status:          s.Status,
		ReceiptsList:    newList,
		ReceiptToken:    s.ReceiptToken,
		ReceiptTokenMap: s.ReceiptTokenMap,
		Auto:            s.Auto,
		SubStatus:       s.SubStatus,
	}
}
