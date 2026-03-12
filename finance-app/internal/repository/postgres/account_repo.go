package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// AccountRepo is a PostgreSQL-backed account repository.
type AccountRepo struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) error {
	q := `
		INSERT INTO accounts (user_id, name, type, balance, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	return r.db.QueryRowContext(ctx, q,
		a.UserID, a.Name, a.Type, a.Balance, a.Currency, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
}

func (r *AccountRepo) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	q := `SELECT id, user_id, name, type, balance, currency, created_at, updated_at
	      FROM accounts WHERE id = $1`
	a := &domain.Account{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account %d not found", id)
	}
	return a, err
}

func (r *AccountRepo) GetByUserID(ctx context.Context, userID int64) ([]*domain.Account, error) {
	q := `SELECT id, user_id, name, type, balance, currency, created_at, updated_at
	      FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Account
	for rows.Next() {
		a := &domain.Account{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AccountRepo) Update(ctx context.Context, a *domain.Account) error {
	q := `UPDATE accounts SET name=$1, type=$2, currency=$3, updated_at=$4 WHERE id=$5`
	_, err := r.db.ExecContext(ctx, q, a.Name, a.Type, a.Currency, time.Now(), a.ID)
	return err
}

func (r *AccountRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE id=$1`, id)
	return err
}

func (r *AccountRepo) UpdateBalance(ctx context.Context, id int64, balance float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET balance=$1, updated_at=$2 WHERE id=$3`,
		balance, time.Now(), id,
	)
	return err
}
