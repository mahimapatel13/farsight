package budgeting

import "time"

// UpdateTransactionRequest represents data needed to update a transaction
type UpdateTransactionRequest struct {
	ItemID          *string    `json:"item_id,omitempty" validate:"omitempty,uuid4"`
	Type            *string    `json:"type,omitempty" validate:"omitempty,oneof=income expense"`
	Amount          *float64   `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Category        *string    `json:"category,omitempty" validate:"omitempty,oneof=food transport shopping bills entertainment health education other"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	TransactionDate *time.Time `json:"transaction_date,omitempty"`
}


