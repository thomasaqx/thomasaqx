package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// CategoryRepo is a PostgreSQL-backed category repository.
type CategoryRepo struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	q := `INSERT INTO categories (user_id, name, description, color, icon, created_at)
	      VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	return r.db.QueryRowContext(ctx, q,
		c.UserID, c.Name, c.Description, c.Color, c.Icon, c.CreatedAt,
	).Scan(&c.ID)
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	c := &domain.Category{}
	q := `SELECT id, user_id, name, description, color, icon, created_at FROM categories WHERE id=$1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Description, &c.Color, &c.Icon, &c.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category %d not found", id)
	}
	return c, err
}

func (r *CategoryRepo) GetByUserID(ctx context.Context, userID int64) ([]*domain.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, description, color, icon, created_at FROM categories WHERE user_id=$1 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Category
	for rows.Next() {
		c := &domain.Category{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Description, &c.Color, &c.Icon, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// BudgetRepo is a PostgreSQL-backed budget repository.
type BudgetRepo struct {
	db *sql.DB
}

func NewBudgetRepo(db *sql.DB) *BudgetRepo {
	return &BudgetRepo{db: db}
}

func (r *BudgetRepo) Create(ctx context.Context, b *domain.Budget) error {
	q := `INSERT INTO budgets (user_id, category_id, amount, spent, period_year, period_month, created_at, updated_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
	return r.db.QueryRowContext(ctx, q,
		b.UserID, b.CategoryID, b.Amount, b.Spent, b.PeriodYear, b.PeriodMonth, b.CreatedAt, b.UpdatedAt,
	).Scan(&b.ID)
}

func (r *BudgetRepo) GetByID(ctx context.Context, id int64) (*domain.Budget, error) {
	b := &domain.Budget{}
	q := `SELECT id, user_id, category_id, amount, spent, period_year, period_month, created_at, updated_at
	      FROM budgets WHERE id=$1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Spent,
		&b.PeriodYear, &b.PeriodMonth, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("budget %d not found", id)
	}
	return b, err
}

func (r *BudgetRepo) GetByUserIDAndPeriod(ctx context.Context, userID int64, year, month int) ([]*domain.Budget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, category_id, amount, spent, period_year, period_month, created_at, updated_at
		 FROM budgets WHERE user_id=$1 AND period_year=$2 AND period_month=$3`,
		userID, year, month,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Budget
	for rows.Next() {
		b := &domain.Budget{}
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Spent,
			&b.PeriodYear, &b.PeriodMonth, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (r *BudgetRepo) Update(ctx context.Context, b *domain.Budget) error {
	q := `UPDATE budgets SET amount=$1, spent=$2, updated_at=$3 WHERE id=$4`
	_, err := r.db.ExecContext(ctx, q, b.Amount, b.Spent, b.UpdatedAt, b.ID)
	return err
}
