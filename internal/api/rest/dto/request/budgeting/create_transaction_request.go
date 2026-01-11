package budgeting

import "time"

// CreateTransactionRequest represents data needed to create a new transaction
type CreateTransactionRequest struct {
	Item            string    `json:"item_id,omitempty" validate:"omitempty,uuid4"`
	Type            string    `json:"type" validate:"required,oneof=income expense"`
	Amount          float64   `json:"amount" validate:"required,gt=0"`
	Category        string    `json:"category" validate:"required,oneof=food transport shopping bills entertainment health education other"`
	Description     string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
}
