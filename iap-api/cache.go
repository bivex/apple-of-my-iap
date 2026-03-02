package iap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Cache handles file-based persistence for subscriptions and plans.
type Cache struct {
	subscriptionsFile string
	plansFile         string
	mu                sync.RWMutex
}

// NewCache creates a new Cache with the given base directory.
func NewCache(baseDir string) (*Cache, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	subscriptionsFile := filepath.Join(baseDir, "subscriptions.json")
	plansFile := filepath.Join(baseDir, "plans.json")

	// Create files if they don't exist
	if _, err := os.Stat(subscriptionsFile); os.IsNotExist(err) {
		if err := os.WriteFile(subscriptionsFile, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create subscriptions file: %w", err)
		}
	}
	if _, err := os.Stat(plansFile); os.IsNotExist(err) {
		if err := os.WriteFile(plansFile, []byte("[]"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create plans file: %w", err)
		}
	}

	return &Cache{
		subscriptionsFile: subscriptionsFile,
		plansFile:         plansFile,
	}, nil
}

// ReadSubscriptions reads all subscriptions from the cache file.
func (c *Cache) ReadSubscriptions() (map[string]*Subscription, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.subscriptionsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscriptions file: %w", err)
	}

	var subs map[string]*Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	return subs, nil
}

// WriteSubscriptions writes subscriptions to the cache file.
func (c *Cache) WriteSubscriptions(subs map[string]*Subscription) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %w", err)
	}

	if err := os.WriteFile(c.subscriptionsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write subscriptions file: %w", err)
	}

	return nil
}

// ReadPlans reads all plans from the cache file.
func (c *Cache) ReadPlans() ([]*Plan, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.plansFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read plans file: %w", err)
	}

	var plans []*Plan
	if err := json.Unmarshal(data, &plans); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plans: %w", err)
	}

	return plans, nil
}
