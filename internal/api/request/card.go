package request

type CardSetSpendingLimit struct {
	SpendingLimit int `json:"spending_limit" binding:"required,min=1,max=50000"`
}

type CardUpdateSpendingControl struct {
	ControlType  string   `json:"control_type" binding:"required,oneof=merchant_category"`
	AllowedCategories []string `json:"allowed_categories"`
	BlockedCategories []string `json:"blocked_categories"`
}
