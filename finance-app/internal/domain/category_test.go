package domain_test

import (
	"testing"

	"github.com/thomasaqx/finance-app/internal/domain"
)

func TestBudget_Remaining(t *testing.T) {
	b := &domain.Budget{Amount: 1000, Spent: 400}
	if b.Remaining() != 600 {
		t.Errorf("expected 600, got %f", b.Remaining())
	}

	b.Spent = 1200
	if b.Remaining() != 0 {
		t.Error("remaining should be 0 when overspent")
	}
}

func TestBudget_IsExceeded(t *testing.T) {
	b := &domain.Budget{Amount: 500, Spent: 300}
	if b.IsExceeded() {
		t.Error("should not be exceeded")
	}
	b.Spent = 600
	if !b.IsExceeded() {
		t.Error("should be exceeded")
	}
}

func TestBudget_Validate(t *testing.T) {
	b := &domain.Budget{UserID: 1, CategoryID: 2, Amount: 100, PeriodYear: 2024, PeriodMonth: 3}
	if err := b.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	b.Amount = 0
	if err := b.Validate(); err == nil {
		t.Error("expected error for zero amount")
	}
}
