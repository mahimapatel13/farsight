package budgeting

import (
	"context"
	"budget-planner/internal/common/errors"
	"budget-planner/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// Service defines the business logic for budgeting
type Service interface {
	CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error)
	GetItem(ctx context.Context, id uuid.UUID) (*Item, error)
	GetItemsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Item, int, error)
	UpdateItem(ctx context.Context, req *UpdateItemRequest) (*Item, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error

	CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error)
	GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Transaction, int, error)
	GetTransactionsByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, offset, limit int) ([]*Transaction, int, error)
	UpdateTransaction(ctx context.Context, req *UpdateTransactionRequest) (*Transaction, error)
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
}

// service is the concrete implementation of the Service interface
type service struct {
	repo   Repository
	logger *logger.Logger
}

// NewService creates a new budgeting service
func NewService(
	repo Repository,
	logger *logger.Logger,
) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateItem creates a new budget item
func (s *service) CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	s.logger.Debug("Creating new item", "userID", req.UserID, "name", req.Name)

	now := time.Now()
	item := &Item{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		s.logger.Error("Failed to create item", "error", err)
		return nil, errors.NewDatabaseError("creating item", err)
	}

	s.logger.Info("Item created successfully", "itemID", item.ID)
	return item, nil
}

// GetItem retrieves an item by ID
func (s *service) GetItem(ctx context.Context, id uuid.UUID) (*Item, error) {
	item, err := s.repo.GetItemByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch item", "itemID", id, "error", err)
		return nil, errors.NewDatabaseError("fetching item", err)
	}
	return item, nil
}

// GetItemsByUserID retrieves items for a user
func (s *service) GetItemsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Item, int, error) {
	items, total, err := s.repo.GetItemsByUserID(ctx, userID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to fetch items", "userID", userID, "error", err)
		return nil, 0, errors.NewDatabaseError("fetching items", err)
	}
	return items, total, nil
}

// UpdateItem updates an existing item
func (s *service) UpdateItem(ctx context.Context, req *UpdateItemRequest) (*Item, error) {
	s.logger.Debug("Updating item", "itemID", req.ID)

	// Get existing item
	item, err := s.repo.GetItemByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("Failed to fetch item for update", "itemID", req.ID, "error", err)
		return nil, errors.NewDatabaseError("fetching item", err)
	}

	// Update fields if provided
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Category != nil {
		item.Category = *req.Category
	}
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		s.logger.Error("Failed to update item", "error", err)
		return nil, errors.NewDatabaseError("updating item", err)
	}

	s.logger.Info("Item updated successfully", "itemID", item.ID)
	return item, nil
}

// DeleteItem deletes an item
func (s *service) DeleteItem(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("Deleting item", "itemID", id)

	if err := s.repo.DeleteItem(ctx, id); err != nil {
		s.logger.Error("Failed to delete item", "itemID", id, "error", err)
		return errors.NewDatabaseError("deleting item", err)
	}

	s.logger.Info("Item deleted successfully", "itemID", id)
	return nil
}

// CreateTransaction creates a new transaction
func (s *service) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error) {
	s.logger.Debug("Creating new transaction", "userID", req.UserID, "type", req.Type, "amount", req.Amount)

	now := time.Now()
	transaction := &Transaction{
		ID:              uuid.New(),
		UserID:          req.UserID,
		ItemID:          req.ItemID,
		Type:            req.Type,
		Amount:          req.Amount,
		Category:        req.Category,
		Description:     req.Description,
		TransactionDate: req.TransactionDate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		s.logger.Error("Failed to create transaction", "error", err)
		return nil, errors.NewDatabaseError("creating transaction", err)
	}

	s.logger.Info("Transaction created successfully", "transactionID", transaction.ID)
	return transaction, nil
}

// GetTransaction retrieves a transaction by ID
func (s *service) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch transaction", "transactionID", id, "error", err)
		return nil, errors.NewDatabaseError("fetching transaction", err)
	}
	return transaction, nil
}

// GetTransactionsByUserID retrieves transactions for a user
func (s *service) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Transaction, int, error) {
	transactions, total, err := s.repo.GetTransactionsByUserID(ctx, userID, offset, limit)
	if err != nil {
		s.logger.Error("Failed to fetch transactions", "userID", userID, "error", err)
		return nil, 0, errors.NewDatabaseError("fetching transactions", err)
	}
	return transactions, total, nil
}

// GetTransactionsByUserIDAndDateRange retrieves transactions for a user within a date range
func (s *service) GetTransactionsByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, offset, limit int) ([]*Transaction, int, error) {
	transactions, total, err := s.repo.GetTransactionsByUserIDAndDateRange(ctx, userID, startDate, endDate, offset, limit)
	if err != nil {
		s.logger.Error("Failed to fetch transactions by date range", "userID", userID, "error", err)
		return nil, 0, errors.NewDatabaseError("fetching transactions", err)
	}
	return transactions, total, nil
}

// UpdateTransaction updates an existing transaction
func (s *service) UpdateTransaction(ctx context.Context, req *UpdateTransactionRequest) (*Transaction, error) {
	s.logger.Debug("Updating transaction", "transactionID", req.ID)

	// Get existing transaction
	transaction, err := s.repo.GetTransactionByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("Failed to fetch transaction for update", "transactionID", req.ID, "error", err)
		return nil, errors.NewDatabaseError("fetching transaction", err)
	}

	// Update fields if provided
	if req.ItemID != nil {
		transaction.ItemID = req.ItemID
	}
	if req.Type != nil {
		transaction.Type = *req.Type
	}
	if req.Amount != nil {
		transaction.Amount = *req.Amount
	}
	if req.Category != nil {
		transaction.Category = *req.Category
	}
	if req.Description != nil {
		transaction.Description = *req.Description
	}
	if req.TransactionDate != nil {
		transaction.TransactionDate = *req.TransactionDate
	}
	transaction.UpdatedAt = time.Now()

	if err := s.repo.UpdateTransaction(ctx, transaction); err != nil {
		s.logger.Error("Failed to update transaction", "error", err)
		return nil, errors.NewDatabaseError("updating transaction", err)
	}

	s.logger.Info("Transaction updated successfully", "transactionID", transaction.ID)
	return transaction, nil
}

// DeleteTransaction deletes a transaction
func (s *service) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("Deleting transaction", "transactionID", id)

	if err := s.repo.DeleteTransaction(ctx, id); err != nil {
		s.logger.Error("Failed to delete transaction", "transactionID", id, "error", err)
		return errors.NewDatabaseError("deleting transaction", err)
	}

	s.logger.Info("Transaction deleted successfully", "transactionID", id)
	return nil
}

