package domain

import (
	"errors"
	"time"
)

// TransactionType distinguishes income from expense.
type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus represents the lifecycle state of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction represents a financial movement.
type Transaction struct {
	ID              int64             `json:"id" db:"id"`
	AccountID       int64             `json:"account_id" db:"account_id"`
	ToAccountID     *int64            `json:"to_account_id,omitempty" db:"to_account_id"`
	CategoryID      *int64            `json:"category_id,omitempty" db:"category_id"`
	Type            TransactionType   `json:"type" db:"type"`
	Status          TransactionStatus `json:"status" db:"status"`
	Amount          float64           `json:"amount" db:"amount"`
	Description     string            `json:"description" db:"description"`
	TransactionDate time.Time         `json:"transaction_date" db:"transaction_date"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// Validate checks mandatory fields.
func (t *Transaction) Validate() error {
	if t.AccountID == 0 {
		return errors.New("account_id is required")
	}
	if t.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	switch t.Type {
	case TransactionTypeIncome, TransactionTypeExpense, TransactionTypeTransfer:
	default:
		return errors.New("invalid transaction type")
	}
	if t.Type == TransactionTypeTransfer && t.ToAccountID == nil {
		return errors.New("to_account_id is required for transfer")
	}
	if t.TransactionDate.IsZero() {
		t.TransactionDate = time.Now()
	}
	return nil
}

// MarkCompleted transitions the transaction to completed state.
func (t *Transaction) MarkCompleted() {
	t.Status = TransactionStatusCompleted
	t.UpdatedAt = time.Now()
}

// MarkFailed transitions the transaction to failed state.
func (t *Transaction) MarkFailed() {
	t.Status = TransactionStatusFailed
	t.UpdatedAt = time.Now()
}

// TransactionFilter holds filtering options for listing transactions.
type TransactionFilter struct {
	AccountID  int64
	Type       TransactionType
	CategoryID *int64
	DateFrom   *time.Time
	DateTo     *time.Time
	MinAmount  *float64
	MaxAmount  *float64
	Limit      int
	Offset     int
}
