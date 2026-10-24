package repository

import (
	"context"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawalRepo struct {
	pool *pgxpool.Pool
}

func NewWithdrawalRepo(pool *pgxpool.Pool) *WithdrawalRepo {
	return &WithdrawalRepo{pool: pool}
}

func (r *WithdrawalRepo) GetBalance(ctx context.Context, userID int) (float64, error) {
	var withdrawn float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return 0, err
	}
	return withdrawn, nil
}

func (r *WithdrawalRepo) GetByUser(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT order_number, COALESCE(sum, 0), processed_at
		 FROM withdrawals
		 WHERE user_id = $1
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var item model.Withdrawal
		if err := rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}
