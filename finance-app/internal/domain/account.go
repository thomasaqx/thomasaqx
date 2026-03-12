package domain

import (
	"errors"
	"time"
)

// AccountType represents the type of financial account.
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeWallet   AccountType = "wallet"
	AccountTypeInvest   AccountType = "invest"
)

// Account is the core financial account entity.
type Account struct {
	ID        int64       `json:"id" db:"id"`
	UserID    int64       `json:"user_id" db:"user_id"`
	Name      string      `json:"name" db:"name"`
	Type      AccountType `json:"type" db:"type"`
	Balance   float64     `json:"balance" db:"balance"`
	Currency  string      `json:"currency" db:"currency"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// Validate checks that the account fields are correct.
func (a *Account) Validate() error {
	if a.Name == "" {
		return errors.New("account name is required")
	}
	if a.UserID == 0 {
		return errors.New("user_id is required")
	}
	if a.Currency == "" {
		a.Currency = "BRL"
	}
	switch a.Type {
	case AccountTypeChecking, AccountTypeSavings, AccountTypeWallet, AccountTypeInvest:
	default:
		return errors.New("invalid account type")
	}
	return nil
}

// Debit reduces the balance by amount. Returns error if funds are insufficient.
func (a *Account) Debit(amount float64) error {
	if amount <= 0 {
		return errors.New("debit amount must be positive")
	}
	if a.Balance < amount {
		return errors.New("insufficient funds")
	}
	a.Balance -= amount
	a.UpdatedAt = time.Now()
	return nil
}

// Credit increases the balance by amount.
func (a *Account) Credit(amount float64) error {
	if amount <= 0 {
		return errors.New("credit amount must be positive")
	}
	a.Balance += amount
	a.UpdatedAt = time.Now()
	return nil
}
