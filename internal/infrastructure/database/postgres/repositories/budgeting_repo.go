package repositories

import (
	"context"
	"budget-planner/internal/common/errors"
	"budget-planner/internal/domain/budgeting"
	"budget-planner/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresBudgetingRepository implements the budgeting.Repository interface
type PostgresBudgetingRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresBudgetingRepository creates a new PostgreSQL-backed budgeting repository
func NewPostgresBudgetingRepository(pool *pgxpool.Pool, logger *logger.Logger) budgeting.Repository {
	return &PostgresBudgetingRepository{
		pool:   pool,
		logger: logger,
	}
}

// CreateItem creates a new item
func (r *PostgresBudgetingRepository) CreateItem(ctx context.Context, item *budgeting.Item) error {
	const query = `
		INSERT INTO budgeting_schema.items (id, user_id, name, description, price, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		item.ID, item.UserID, item.Name, item.Description, item.Price, item.Category, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("creating item", err)
	}
	return nil
}

// GetItemByID retrieves an item by ID
func (r *PostgresBudgetingRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*budgeting.Item, error) {
	const query = `
		SELECT id, user_id, name, description, price, category, created_at, updated_at
		FROM budgeting_schema.items
		WHERE id = $1
	`

	item := &budgeting.Item{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.UserID, &item.Name, &item.Description, &item.Price, &item.Category, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("item not found", map[string]interface{}{"id": id})
		}
		return nil, errors.NewDatabaseError("fetching item", err)
	}
	return item, nil
}

// GetItemsByUserID retrieves items for a user with pagination
func (r *PostgresBudgetingRepository) GetItemsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*budgeting.Item, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM budgeting_schema.items WHERE user_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("counting items", err)
	}

	// Get items
	const query = `
		SELECT id, user_id, name, description, price, category, created_at, updated_at
		FROM budgeting_schema.items
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("fetching items", err)
	}
	defer rows.Close()

	var items []*budgeting.Item
	for rows.Next() {
		item := &budgeting.Item{}
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Name, &item.Description, &item.Price, &item.Category, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewDatabaseError("scanning item", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

// UpdateItem updates an existing item
func (r *PostgresBudgetingRepository) UpdateItem(ctx context.Context, item *budgeting.Item) error {
	const query = `
		UPDATE budgeting_schema.items
		SET name = $2, description = $3, price = $4, category = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		item.ID, item.Name, item.Description, item.Price, item.Category, item.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("updating item", err)
	}
	return nil
}

// DeleteItem deletes an item
func (r *PostgresBudgetingRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM budgeting_schema.items WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("deleting item", err)
	}
	return nil
}

// CreateTransaction creates a new transaction
func (r *PostgresBudgetingRepository) CreateTransaction(ctx context.Context, transaction *budgeting.Transaction) error {
	const query = `
		INSERT INTO budgeting_schema.transactions (
			id, user_id, item_id, type, amount, category, description, transaction_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		transaction.ID, transaction.UserID, transaction.ItemID, transaction.Type,
		transaction.Amount, transaction.Category, transaction.Description,
		transaction.TransactionDate, transaction.CreatedAt, transaction.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("creating transaction", err)
	}
	return nil
}

// GetTransactionByID retrieves a transaction by ID
func (r *PostgresBudgetingRepository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*budgeting.Transaction, error) {
	const query = `
		SELECT id, user_id, item_id, type, amount, category, description, transaction_date, created_at, updated_at
		FROM budgeting_schema.transactions
		WHERE id = $1
	`

	transaction := &budgeting.Transaction{}
	var itemID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&transaction.ID, &transaction.UserID, &itemID, &transaction.Type,
		&transaction.Amount, &transaction.Category, &transaction.Description,
		&transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("transaction not found", map[string]interface{}{"id": id})
		}
		return nil, errors.NewDatabaseError("fetching transaction", err)
	}
	transaction.ItemID = itemID
	return transaction, nil
}

// GetTransactionsByUserID retrieves transactions for a user with pagination
func (r *PostgresBudgetingRepository) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*budgeting.Transaction, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM budgeting_schema.transactions WHERE user_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("counting transactions", err)
	}

	// Get transactions
	const query = `
		SELECT id, user_id, item_id, type, amount, category, description, transaction_date, created_at, updated_at
		FROM budgeting_schema.transactions
		WHERE user_id = $1
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("fetching transactions", err)
	}
	defer rows.Close()

	var transactions []*budgeting.Transaction
	for rows.Next() {
		transaction := &budgeting.Transaction{}
		var itemID *uuid.UUID
		err := rows.Scan(
			&transaction.ID, &transaction.UserID, &itemID, &transaction.Type,
			&transaction.Amount, &transaction.Category, &transaction.Description,
			&transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewDatabaseError("scanning transaction", err)
		}
		transaction.ItemID = itemID
		transactions = append(transactions, transaction)
	}

	return transactions, total, nil
}

// GetTransactionsByUserIDAndDateRange retrieves transactions for a user within a date range
func (r *PostgresBudgetingRepository) GetTransactionsByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, offset, limit int) ([]*budgeting.Transaction, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM budgeting_schema.transactions WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date <= $3`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID, startDate, endDate).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("counting transactions", err)
	}

	// Get transactions
	const query = `
		SELECT id, user_id, item_id, type, amount, category, description, transaction_date, created_at, updated_at
		FROM budgeting_schema.transactions
		WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date <= $3
		ORDER BY transaction_date DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, userID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("fetching transactions", err)
	}
	defer rows.Close()

	var transactions []*budgeting.Transaction
	for rows.Next() {
		transaction := &budgeting.Transaction{}
		var itemID *uuid.UUID
		err := rows.Scan(
			&transaction.ID, &transaction.UserID, &itemID, &transaction.Type,
			&transaction.Amount, &transaction.Category, &transaction.Description,
			&transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewDatabaseError("scanning transaction", err)
		}
		transaction.ItemID = itemID
		transactions = append(transactions, transaction)
	}

	return transactions, total, nil
}

// UpdateTransaction updates an existing transaction
func (r *PostgresBudgetingRepository) UpdateTransaction(ctx context.Context, transaction *budgeting.Transaction) error {
	const query = `
		UPDATE budgeting_schema.transactions
		SET item_id = $2, type = $3, amount = $4, category = $5, description = $6, transaction_date = $7, updated_at = $8
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		transaction.ID, transaction.ItemID, transaction.Type, transaction.Amount,
		transaction.Category, transaction.Description, transaction.TransactionDate, transaction.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("updating transaction", err)
	}
	return nil
}

// DeleteTransaction deletes a transaction
func (r *PostgresBudgetingRepository) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM budgeting_schema.transactions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("deleting transaction", err)
	}
	return nil
}

