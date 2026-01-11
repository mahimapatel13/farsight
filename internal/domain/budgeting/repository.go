package budgeting

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the data access interface for budgeting
type Repository interface {
	// Item operations
	CreateItem(ctx context.Context, item *Item) error
	GetItemByID(ctx context.Context, id uuid.UUID) (*Item, error)
	GetItemsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Item, int, error)
	UpdateItem(ctx context.Context, item *Item) error
	DeleteItem(ctx context.Context, id uuid.UUID) error

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Transaction, int, error)
	GetTransactionsByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, offset, limit int) ([]*Transaction, int, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
}

