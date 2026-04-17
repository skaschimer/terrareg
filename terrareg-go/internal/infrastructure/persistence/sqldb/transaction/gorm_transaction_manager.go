package transaction

import (
	"context"

	"gorm.io/gorm"

	domainTransaction "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/transaction"
)

// GormTransactionManager implements the domain TransactionManager interface
// using GORM and the existing SavepointHelper for transaction management.
//
// This adapter allows domain services to use transaction management without
// depending on GORM directly, enabling unit testing with mock implementations.
type GormTransactionManager struct {
	// helper provides the underlying GORM transaction/savepoint functionality
	helper *SavepointHelper
}

// NewGormTransactionManager creates a new GORM-based transaction manager.
// Returns an error if the underlying SavepointHelper cannot be created.
//
// The db parameter is the GORM database instance to use for transactions.
func NewGormTransactionManager(db *gorm.DB) (domainTransaction.TransactionManager, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}

	helper, err := NewSavepointHelper(db)
	if err != nil {
		return nil, err
	}

	return &GormTransactionManager{
		helper: helper,
	}, nil
}

// WithTransaction executes the given function within a transaction.
//
// If a transaction already exists in the context, it will create a savepoint
// for this operation (nested transaction support). Otherwise, it creates a
// new transaction.
//
// The function receives a context with the transaction active. If the function
// returns an error, the transaction (or savepoint) is rolled back. Otherwise,
// it is committed.
func (m *GormTransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return m.helper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Adapt the callback function to not expose GORM to the domain layer
		return fn(ctx)
	})
}

// WithNamedTransaction is like WithTransaction but allows specifying a name
// for the transaction/savepoint.
//
// The name is used for savepoint identifiers in databases that support
// nested transactions. If empty, a unique name is generated.
func (m *GormTransactionManager) WithNamedTransaction(ctx context.Context, name string, fn func(context.Context) error) error {
	return m.helper.WithNamedTransaction(ctx, name, func(ctx context.Context, tx *gorm.DB) error {
		// Adapt the callback function to not expose GORM to the domain layer
		return fn(ctx)
	})
}

// IsTransactionActive checks if a transaction is currently active in the
// given context.
//
// This can be used by domain services to determine if they're participating
// in an existing transaction or if a new one would be created.
func (m *GormTransactionManager) IsTransactionActive(ctx context.Context) bool {
	return m.helper.IsTransactionActive(ctx)
}

// GetDBFromContext retrieves the GORM DB instance from the context if a
// transaction is active.
//
// This is a convenience method for repository implementations that need
// access to the current transaction. This method is intentionally not part
// of the domain TransactionManager interface to keep the domain layer
// independent of GORM.
//
// Repository implementations in the infrastructure layer can type-assert
// the TransactionManager to *GormTransactionManager to access this method,
// or repositories can directly use the BaseRepository which handles this.
func (m *GormTransactionManager) GetDBFromContext(ctx context.Context) *gorm.DB {
	// Check if transaction is active in context
	if m.helper.IsTransactionActive(ctx) {
		if tx, exists := ctx.Value(TransactionDBKey).(*gorm.DB); exists {
			return tx
		}
	}
	// No active transaction, return db with context
	return m.helper.WithContext(ctx)
}

// WithGormTransaction executes the given function within a transaction and
// provides the GORM DB instance to the callback.
//
// This is a convenience method for infrastructure code that needs direct
// access to GORM (e.g., repository implementations). This method is not
// part of the domain TransactionManager interface and should only be used
// in the infrastructure layer.
func (m *GormTransactionManager) WithGormTransaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return m.helper.WithTransaction(ctx, fn)
}

// ErrNilDatabase is returned when attempting to create a transaction manager
// with a nil database.
var ErrNilDatabase = &TransactionManagerError{
	Message: "database cannot be nil",
}

// TransactionManagerError represents an error in transaction management.
type TransactionManagerError struct {
	Message string
}

func (e *TransactionManagerError) Error() string {
	return e.Message
}
