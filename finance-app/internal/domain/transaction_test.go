package domain_test

import (
	"testing"
	"time"

	"github.com/thomasaqx/finance-app/internal/domain"
)

func TestTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      domain.Transaction
		wantErr bool
	}{
		{
			name: "valid income",
			tx: domain.Transaction{
				AccountID: 1,
				Type:      domain.TransactionTypeIncome,
				Amount:    500,
			},
			wantErr: false,
		},
		{
			name: "transfer without to_account",
			tx: domain.Transaction{
				AccountID: 1,
				Type:      domain.TransactionTypeTransfer,
				Amount:    100,
			},
			wantErr: true,
		},
		{
			name:    "zero amount",
			tx:      domain.Transaction{AccountID: 1, Type: domain.TransactionTypeExpense, Amount: 0},
			wantErr: true,
		},
		{
			name:    "missing account_id",
			tx:      domain.Transaction{Type: domain.TransactionTypeExpense, Amount: 10},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransaction_MarkCompleted(t *testing.T) {
	tx := &domain.Transaction{Status: domain.TransactionStatusPending}
	tx.MarkCompleted()
	if tx.Status != domain.TransactionStatusCompleted {
		t.Errorf("expected completed, got %s", tx.Status)
	}
	if tx.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestTransaction_DefaultDate(t *testing.T) {
	tx := &domain.Transaction{AccountID: 1, Type: domain.TransactionTypeIncome, Amount: 10}
	before := time.Now()
	_ = tx.Validate()
	after := time.Now()

	if tx.TransactionDate.Before(before) || tx.TransactionDate.After(after) {
		t.Errorf("transaction_date should default to now, got %v", tx.TransactionDate)
	}
}
