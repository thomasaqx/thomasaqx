package domain_test

import (
	"testing"

	"github.com/thomasaqx/finance-app/internal/domain"
)

func TestAccount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		account domain.Account
		wantErr bool
	}{
		{
			name:    "valid account",
			account: domain.Account{UserID: 1, Name: "My Checking", Type: domain.AccountTypeChecking},
			wantErr: false,
		},
		{
			name:    "missing name",
			account: domain.Account{UserID: 1, Type: domain.AccountTypeChecking},
			wantErr: true,
		},
		{
			name:    "missing user_id",
			account: domain.Account{Name: "My Account", Type: domain.AccountTypeChecking},
			wantErr: true,
		},
		{
			name:    "invalid type",
			account: domain.Account{UserID: 1, Name: "Test", Type: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccount_Debit(t *testing.T) {
	a := &domain.Account{Balance: 100}

	if err := a.Debit(30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Balance != 70 {
		t.Errorf("expected balance 70, got %f", a.Balance)
	}

	if err := a.Debit(200); err == nil {
		t.Error("expected insufficient funds error")
	}

	if err := a.Debit(0); err == nil {
		t.Error("expected error for zero debit")
	}
}

func TestAccount_Credit(t *testing.T) {
	a := &domain.Account{Balance: 50}

	if err := a.Credit(25); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Balance != 75 {
		t.Errorf("expected balance 75, got %f", a.Balance)
	}

	if err := a.Credit(-10); err == nil {
		t.Error("expected error for negative credit")
	}
}
