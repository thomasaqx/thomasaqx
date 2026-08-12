package service_test

import (
	"context"
	"testing"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// --- mocks ---

type memAccountRepo struct {
	accounts map[int64]*domain.Account
	nextID   int64
}

func newMemAccountRepo() *memAccountRepo {
	return &memAccountRepo{accounts: make(map[int64]*domain.Account)}
}

func (r *memAccountRepo) Create(_ context.Context, a *domain.Account) error {
	r.nextID++
	a.ID = r.nextID
	cp := *a
	r.accounts[a.ID] = &cp
	return nil
}

func (r *memAccountRepo) GetByID(_ context.Context, id int64) (*domain.Account, error) {
	a, ok := r.accounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *memAccountRepo) GetByUserID(_ context.Context, userID int64) ([]*domain.Account, error) {
	var list []*domain.Account
	for _, a := range r.accounts {
		if a.UserID == userID {
			cp := *a
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (r *memAccountRepo) Update(_ context.Context, a *domain.Account) error {
	r.accounts[a.ID] = a
	return nil
}

func (r *memAccountRepo) Delete(_ context.Context, id int64) error {
	delete(r.accounts, id)
	return nil
}

func (r *memAccountRepo) UpdateBalance(_ context.Context, id int64, balance float64) error {
	if a, ok := r.accounts[id]; ok {
		a.Balance = balance
	}
	return nil
}

type memCache struct{}

func (m *memCache) Set(_ context.Context, _ string, _ interface{}, _ int) error { return nil }
func (m *memCache) Get(_ context.Context, _ string, _ interface{}) error {
	return domain.ErrNotFound
}
func (m *memCache) Delete(_ context.Context, _ string) error { return nil }

type memPublisher struct{}

func (m *memPublisher) Publish(_ context.Context, _ string, _ interface{}) error { return nil }

// --- tests ---

func TestAccountService_CreateAndGet(t *testing.T) {
	repo := newMemAccountRepo()
	svc := service.NewAccountService(repo, &memCache{}, &memPublisher{})
	ctx := context.Background()

	a := &domain.Account{
		UserID:   1,
		Name:     "My Wallet",
		Type:     domain.AccountTypeWallet,
		Currency: "BRL",
	}
	if err := svc.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.ID == 0 {
		t.Fatal("expected ID to be set after create")
	}

	got, err := svc.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != a.Name {
		t.Errorf("expected name %q, got %q", a.Name, got.Name)
	}
}

func TestAccountService_Create_Validation(t *testing.T) {
	svc := service.NewAccountService(newMemAccountRepo(), &memCache{}, &memPublisher{})
	err := svc.Create(context.Background(), &domain.Account{Type: domain.AccountTypeChecking})
	if err == nil {
		t.Error("expected validation error for missing name")
	}
}

func TestAccountService_Delete(t *testing.T) {
	repo := newMemAccountRepo()
	svc := service.NewAccountService(repo, &memCache{}, &memPublisher{})
	ctx := context.Background()

	a := &domain.Account{UserID: 1, Name: "Test", Type: domain.AccountTypeSavings}
	_ = svc.Create(ctx, a)

	if err := svc.Delete(ctx, a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := svc.GetByID(ctx, a.ID); err == nil {
		t.Error("expected error after delete")
	}
}
