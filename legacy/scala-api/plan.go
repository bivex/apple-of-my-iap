package iap

// Plan represents a subscription billing plan.
type Plan struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	BillInterval      int    `json:"billInterval"`
	BillIntervalUnit  string `json:"billIntervalUnit"`
	TrialInterval     int    `json:"trialInterval"`
	TrialIntervalUnit string `json:"trialIntervalUnit"`
	ProductID         string `json:"productId"`
}
