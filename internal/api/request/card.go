package request

type CardSetSpendingLimit struct {
	SpendingLimit int `json:"spending_limit" binding:"required,min=1,max=50000"`
}
