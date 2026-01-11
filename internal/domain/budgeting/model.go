package budgeting

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

// Category represents a budget category
type Category string

const (
	CategoryFood       Category = "food"
	CategoryTransport  Category = "transport"
	CategoryShopping   Category = "shopping"
	CategoryBills      Category = "bills"
	CategoryEntertainment Category = "entertainment"
	CategoryHealth     Category = "health"
	CategoryEducation  Category = "education"
	CategoryOther      Category = "other"
)

// Item represents a budget item (product/service) with price information
type Item struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	Price       float64
	Category    Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Transaction represents a financial transaction
type Transaction struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ItemID        *uuid.UUID // Optional: link to an item
	Type          TransactionType
	Amount        float64
	Category      Category
	Description   string
	TransactionDate time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateItemRequest represents data needed to create a new item
type CreateItemRequest struct {
	UserID      uuid.UUID
	Name        string
	Description string
	Price       float64
	Category    Category
}

// UpdateItemRequest represents data needed to update an item
type UpdateItemRequest struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Price       *float64
	Category    *Category
}

// CreateTransactionRequest represents data needed to create a new transaction
type CreateTransactionRequest struct {
	UserID          uuid.UUID
	ItemID          *uuid.UUID
	Type            TransactionType
	Amount          float64
	Category        Category
	Description     string
	TransactionDate time.Time
}

// UpdateTransactionRequest represents data needed to update a transaction
type UpdateTransactionRequest struct {
	ID              uuid.UUID
	ItemID          *uuid.UUID
	Type            *TransactionType
	Amount          *float64
	Category        *Category
	Description     *string
	TransactionDate *time.Time
}

