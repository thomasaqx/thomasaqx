package service_test

import (
	"context"
	"testing"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// memTransactionRepo is an in-memory transaction repository.
type memTransactionRepo struct {
	txs    map[int64]*domain.Transaction
	nextID int64
}

func newMemTxRepo() *memTransactionRepo {
	return &memTransactionRepo{txs: make(map[int64]*domain.Transaction)}
}

func (r *memTransactionRepo) Create(_ context.Context, t *domain.Transaction) error {
	r.nextID++
	t.ID = r.nextID
	cp := *t
	r.txs[t.ID] = &cp
	return nil
}

func (r *memTransactionRepo) GetByID(_ context.Context, id int64) (*domain.Transaction, error) {
	t, ok := r.txs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *memTransactionRepo) List(_ context.Context, f domain.TransactionFilter) ([]*domain.Transaction, error) {
	var list []*domain.Transaction
	for _, t := range r.txs {
		if t.AccountID == f.AccountID {
			cp := *t
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (r *memTransactionRepo) Update(_ context.Context, t *domain.Transaction) error {
	r.txs[t.ID] = t
	return nil
}

func (r *memTransactionRepo) SumByType(_ context.Context, accountID int64, txType domain.TransactionType) (float64, error) {
	var sum float64
	for _, t := range r.txs {
		if t.AccountID == accountID && t.Type == txType && t.Status == domain.TransactionStatusCompleted {
			sum += t.Amount
		}
	}
	return sum, nil
}

func setupTxService(t *testing.T) (*service.TransactionService, *memAccountRepo) {
	t.Helper()
	accountRepo := newMemAccountRepo()
	txRepo := newMemTxRepo()
	svc := service.NewTransactionService(txRepo, accountRepo, &memCache{}, &memPublisher{}, 2)
	t.Cleanup(svc.Stop)
	return svc, accountRepo
}

func seedAccount(t *testing.T, repo *memAccountRepo, balance float64) *domain.Account {
	t.Helper()
	a := &domain.Account{
		UserID:   1,
		Name:     "Test Account",
		Type:     domain.AccountTypeChecking,
		Balance:  balance,
		Currency: "BRL",
	}
	_ = repo.Create(context.Background(), a)
	return a
}

func TestTransactionService_Income(t *testing.T) {
	svc, accountRepo := setupTxService(t)
	acc := seedAccount(t, accountRepo, 0)

	tx := &domain.Transaction{
		AccountID: acc.ID,
		Type:      domain.TransactionTypeIncome,
		Amount:    500,
	}
	if err := svc.Create(context.Background(), tx); err != nil {
		t.Fatalf("Create income: %v", err)
	}
	if tx.Status != domain.TransactionStatusCompleted {
		t.Errorf("expected completed, got %s", tx.Status)
	}

	updated, _ := accountRepo.GetByID(context.Background(), acc.ID)
	if updated.Balance != 500 {
		t.Errorf("expected balance 500, got %f", updated.Balance)
	}
}

func TestTransactionService_Expense_InsufficientFunds(t *testing.T) {
	svc, accountRepo := setupTxService(t)
	acc := seedAccount(t, accountRepo, 100)

	tx := &domain.Transaction{
		AccountID: acc.ID,
		Type:      domain.TransactionTypeExpense,
		Amount:    500,
	}
	err := svc.Create(context.Background(), tx)
	if err == nil {
		t.Error("expected insufficient funds error")
	}
}

func TestTransactionService_Summary(t *testing.T) {
	svc, accountRepo := setupTxService(t)
	acc := seedAccount(t, accountRepo, 1000)

	ctx := context.Background()
	_ = svc.Create(ctx, &domain.Transaction{AccountID: acc.ID, Type: domain.TransactionTypeIncome, Amount: 1000})
	_ = svc.Create(ctx, &domain.Transaction{AccountID: acc.ID, Type: domain.TransactionTypeExpense, Amount: 200})

	income, expense, net, err := svc.Summary(ctx, acc.ID)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if income != 1000 {
		t.Errorf("expected income 1000, got %f", income)
	}
	if expense != 200 {
		t.Errorf("expected expense 200, got %f", expense)
	}
	if net != 800 {
		t.Errorf("expected net 800, got %f", net)
	}
}

func TestTransactionService_CreateAsync(t *testing.T) {
	svc, accountRepo := setupTxService(t)
	acc := seedAccount(t, accountRepo, 0)

	resultCh := svc.CreateAsync(&domain.Transaction{
		AccountID: acc.ID,
		Type:      domain.TransactionTypeIncome,
		Amount:    100,
	})
	if err := <-resultCh; err != nil {
		t.Fatalf("CreateAsync: %v", err)
	}
}
