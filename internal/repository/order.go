package repository

import (
	"context"
	"errors"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) Save(ctx context.Context, order model.Order) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders (number, status, accrual, uploaded_at, user_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		order.Number, order.Status, order.Accrual, order.UploadedAt, order.UserID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (r *OrderRepo) GetByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	var order model.Order
	var accrual *float64
	err := r.pool.QueryRow(ctx,
		`SELECT number, status, accrual, uploaded_at, user_id FROM orders WHERE number = $1`,
		orderNumber,
	).Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt, &order.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if accrual != nil {
		order.Accrual = *accrual
	}
	return &order, nil
}

func (r *OrderRepo) GetByUser(ctx context.Context, userID int) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT number, status, accrual, uploaded_at, user_id
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		var accrual *float64
		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt, &order.UserID); err != nil {
			return nil, err
		}
		if accrual != nil {
			order.Accrual = *accrual
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepo) GetProcessedOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT number, status, accrual, uploaded_at, user_id
		 FROM orders
		 WHERE status IN ('NEW', 'PROCESSING')`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		var accrual *float64
		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt, &order.UserID); err != nil {
			return nil, err
		}
		if accrual != nil {
			order.Accrual = *accrual
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepo) UpdateOrder(ctx context.Context, order model.Order) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $2
		 WHERE number = $1 AND status IN ('NEW', 'PROCESSING')`,
		order.Number, order.Status,
	)
	return err
}

func (r *OrderRepo) AccrueProcessed(ctx context.Context, order model.Order, userID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE orders
		 SET status = 'PROCESSED', accrual = $2
		 WHERE number = $1 AND status IN ('NEW', 'PROCESSING')`,
		order.Number, order.Accrual,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET balance = balance + $1 WHERE id = $2`,
		order.Accrual, userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
