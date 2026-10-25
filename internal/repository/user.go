package repository

import (
	"context"
	"errors"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConflict          = errors.New("conflict")
	ErrNotFound          = errors.New("not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

const uniqueViolation = "23505"

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, login, passwordHash string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		login, passwordHash,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return 0, ErrConflict
		}
		return 0, err
	}
	return id, nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetBalance(ctx context.Context, userID int) (float64, error) {
	var balance float64
	err := r.pool.QueryRow(ctx,
		`SELECT balance FROM users WHERE id = $1`,
		userID,
	).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return balance, nil
}

func (r *UserRepo) ReduceBalance(ctx context.Context, orderNumber string, sum float64, userID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		sum, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO withdrawals (order_number, sum, processed_at, user_id)
		 VALUES ($1, $2, $3, $4)`,
		orderNumber, sum, time.Now(), userID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrConflict
		}
		return err
	}

	return tx.Commit(ctx)
}
