package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// TransactionRepo is a PostgreSQL-backed transaction repository.
type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, t *domain.Transaction) error {
	q := `
		INSERT INTO transactions
		  (account_id, to_account_id, category_id, type, status, amount, description, transaction_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id`
	return r.db.QueryRowContext(ctx, q,
		t.AccountID, t.ToAccountID, t.CategoryID, t.Type, t.Status,
		t.Amount, t.Description, t.TransactionDate, t.CreatedAt, t.UpdatedAt,
	).Scan(&t.ID)
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	q := `SELECT id, account_id, to_account_id, category_id, type, status, amount,
	             description, transaction_date, created_at, updated_at
	      FROM transactions WHERE id = $1`
	t := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.AccountID, &t.ToAccountID, &t.CategoryID, &t.Type, &t.Status,
		&t.Amount, &t.Description, &t.TransactionDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction %d not found", id)
	}
	return t, err
}

// List retrieves transactions with dynamic filtering.
func (r *TransactionRepo) List(ctx context.Context, f domain.TransactionFilter) ([]*domain.Transaction, error) {
	args := []interface{}{}
	conds := []string{"account_id = $1"}
	args = append(args, f.AccountID)
	idx := 2

	if f.Type != "" {
		conds = append(conds, fmt.Sprintf("type = $%d", idx))
		args = append(args, f.Type)
		idx++
	}
	if f.CategoryID != nil {
		conds = append(conds, fmt.Sprintf("category_id = $%d", idx))
		args = append(args, *f.CategoryID)
		idx++
	}
	if f.DateFrom != nil {
		conds = append(conds, fmt.Sprintf("transaction_date >= $%d", idx))
		args = append(args, *f.DateFrom)
		idx++
	}
	if f.DateTo != nil {
		conds = append(conds, fmt.Sprintf("transaction_date <= $%d", idx))
		args = append(args, *f.DateTo)
		idx++
	}

	limit := f.Limit
	if limit == 0 {
		limit = 20
	}

	q := fmt.Sprintf(`
		SELECT id, account_id, to_account_id, category_id, type, status, amount,
		       description, transaction_date, created_at, updated_at
		FROM transactions
		WHERE %s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(conds, " AND "), idx, idx+1)
	args = append(args, limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := rows.Scan(
			&t.ID, &t.AccountID, &t.ToAccountID, &t.CategoryID, &t.Type, &t.Status,
			&t.Amount, &t.Description, &t.TransactionDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *TransactionRepo) Update(ctx context.Context, t *domain.Transaction) error {
	q := `UPDATE transactions SET status=$1, updated_at=$2 WHERE id=$3`
	_, err := r.db.ExecContext(ctx, q, t.Status, t.UpdatedAt, t.ID)
	return err
}

func (r *TransactionRepo) SumByType(ctx context.Context, accountID int64, txType domain.TransactionType) (float64, error) {
	var total sql.NullFloat64
	q := `SELECT COALESCE(SUM(amount), 0) FROM transactions
	      WHERE account_id=$1 AND type=$2 AND status='completed'`
	err := r.db.QueryRowContext(ctx, q, accountID, txType).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Float64, nil
}
