package repository

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
	"gorm.io/gorm"
)

// BaseRepository provides common database context handling for all repositories
// This eliminates duplicate getDBFromContext implementations across repositories
type BaseRepository struct {
	// db provides database access (required)
	db *gorm.DB
	// helper provides transaction savepoint management (required)
	helper *transaction.SavepointHelper
}

// NewBaseRepository creates a new base repository
// Returns an error if db is nil
func NewBaseRepository(db *gorm.DB) (*BaseRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	helper, err := transaction.NewSavepointHelper(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint helper: %w", err)
	}

	return &BaseRepository{
		db:     db,
		helper: helper,
	}, nil
}

// NewBaseRepositoryWithoutSavepoint creates a base repository without savepoint support
// Use this when you need a base repository but don't have a db instance yet
// This is used internally by other repositories that need to create a BaseRepository
func NewBaseRepositoryWithoutSavepoint(db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		db:     db,
		helper: nil, // Will be initialized later
	}
}

// GetDBFromContext returns the appropriate database instance for the given context
// If a transaction is active in the context, it uses that transaction
// Otherwise, it returns the database instance with the context applied
func (r *BaseRepository) GetDBFromContext(ctx context.Context) *gorm.DB {
	// Check if transaction is active in context
	if r.helper.IsTransactionActive(ctx) {
		if tx, exists := ctx.Value(transaction.TransactionDBKey).(*gorm.DB); exists {
			return tx
		}
	}

	// No active transaction, use db with context
	return r.helper.WithContext(ctx)
}

// IsTransactionActive checks if a transaction is active in the context
func (r *BaseRepository) IsTransactionActive(ctx context.Context) bool {
	return r.helper.IsTransactionActive(ctx)
}
