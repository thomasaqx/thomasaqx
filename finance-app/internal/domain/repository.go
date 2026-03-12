package domain

import "context"

// AccountRepository defines persistence operations for accounts.
type AccountRepository interface {
	Create(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id int64) (*Account, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Account, error)
	Update(ctx context.Context, a *Account) error
	Delete(ctx context.Context, id int64) error
	UpdateBalance(ctx context.Context, id int64, balance float64) error
}

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, t *Transaction) error
	GetByID(ctx context.Context, id int64) (*Transaction, error)
	List(ctx context.Context, filter TransactionFilter) ([]*Transaction, error)
	Update(ctx context.Context, t *Transaction) error
	SumByType(ctx context.Context, accountID int64, txType TransactionType) (float64, error)
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id int64) (*Category, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Category, error)
}

// BudgetRepository defines persistence operations for budgets.
type BudgetRepository interface {
	Create(ctx context.Context, b *Budget) error
	GetByID(ctx context.Context, id int64) (*Budget, error)
	GetByUserIDAndPeriod(ctx context.Context, userID int64, year, month int) ([]*Budget, error)
	Update(ctx context.Context, b *Budget) error
}

// CacheRepository defines cache operations.
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
}

// EventPublisher defines async event publishing.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event interface{}) error
}
