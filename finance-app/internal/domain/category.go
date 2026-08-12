package domain

import (
	"errors"
	"time"
)

// Category is used to classify transactions (e.g. Food, Rent, Salary).
type Category struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	Icon        string    `json:"icon" db:"icon"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Validate checks mandatory fields.
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("category name is required")
	}
	if c.UserID == 0 {
		return errors.New("user_id is required")
	}
	return nil
}

// Budget defines a spending limit per category within a period.
type Budget struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	CategoryID int64     `json:"category_id" db:"category_id"`
	Amount     float64   `json:"amount" db:"amount"`
	Spent      float64   `json:"spent" db:"spent"`
	PeriodYear int       `json:"period_year" db:"period_year"`
	PeriodMonth int      `json:"period_month" db:"period_month"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Remaining returns how much budget is still available.
func (b *Budget) Remaining() float64 {
	r := b.Amount - b.Spent
	if r < 0 {
		return 0
	}
	return r
}

// IsExceeded returns true when spending surpassed the limit.
func (b *Budget) IsExceeded() bool {
	return b.Spent > b.Amount
}

// Validate checks budget fields.
func (b *Budget) Validate() error {
	if b.UserID == 0 {
		return errors.New("user_id is required")
	}
	if b.CategoryID == 0 {
		return errors.New("category_id is required")
	}
	if b.Amount <= 0 {
		return errors.New("budget amount must be positive")
	}
	if b.PeriodYear == 0 || b.PeriodMonth == 0 {
		return errors.New("period_year and period_month are required")
	}
	return nil
}
