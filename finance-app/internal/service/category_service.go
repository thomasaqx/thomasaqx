package service

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// CategoryService handles category and budget operations.
type CategoryService struct {
	catRepo    domain.CategoryRepository
	budgetRepo domain.BudgetRepository
}

func NewCategoryService(catRepo domain.CategoryRepository, budgetRepo domain.BudgetRepository) *CategoryService {
	return &CategoryService{catRepo: catRepo, budgetRepo: budgetRepo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, c *domain.Category) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	c.CreatedAt = time.Now()
	return s.catRepo.Create(ctx, c)
}

func (s *CategoryService) GetCategoriesByUser(ctx context.Context, userID int64) ([]*domain.Category, error) {
	return s.catRepo.GetByUserID(ctx, userID)
}

func (s *CategoryService) CreateBudget(ctx context.Context, b *domain.Budget) error {
	if err := b.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return s.budgetRepo.Create(ctx, b)
}

func (s *CategoryService) GetBudgets(ctx context.Context, userID int64, year, month int) ([]*domain.Budget, error) {
	return s.budgetRepo.GetByUserIDAndPeriod(ctx, userID, year, month)
}
