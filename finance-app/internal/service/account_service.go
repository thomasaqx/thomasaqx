package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// AccountService implements business logic for accounts.
type AccountService struct {
	repo      domain.AccountRepository
	cache     domain.CacheRepository
	publisher domain.EventPublisher
}

func NewAccountService(
	repo domain.AccountRepository,
	cache domain.CacheRepository,
	publisher domain.EventPublisher,
) *AccountService {
	return &AccountService{repo: repo, cache: cache, publisher: publisher}
}

func (s *AccountService) Create(ctx context.Context, a *domain.Account) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	if err := s.repo.Create(ctx, a); err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	_ = s.cache.Delete(ctx, cacheKeyUserAccounts(a.UserID))
	_ = s.publisher.Publish(ctx, "account.created", a)
	return nil
}

func (s *AccountService) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	var cached domain.Account
	key := cacheKeyAccount(id)
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	_ = s.cache.Set(ctx, key, a, 300)
	return a, nil
}

func (s *AccountService) GetByUserID(ctx context.Context, userID int64) ([]*domain.Account, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *AccountService) Update(ctx context.Context, a *domain.Account) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	a.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, a); err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	_ = s.cache.Delete(ctx, cacheKeyAccount(a.ID))
	_ = s.cache.Delete(ctx, cacheKeyUserAccounts(a.UserID))
	return nil
}

func (s *AccountService) Delete(ctx context.Context, id int64) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	_ = s.cache.Delete(ctx, cacheKeyAccount(id))
	_ = s.cache.Delete(ctx, cacheKeyUserAccounts(a.UserID))
	return nil
}

func cacheKeyAccount(id int64) string {
	return fmt.Sprintf("account:%d", id)
}

func cacheKeyUserAccounts(userID int64) string {
	return fmt.Sprintf("user:%d:accounts", userID)
}
