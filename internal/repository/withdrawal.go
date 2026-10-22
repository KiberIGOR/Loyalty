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


func (r *WithdrawalRepo) 	GetBalance(ctx context.Context, userID int) (int64, error){
	
}

func (r *WithdrawalRepo) 	GetByUser(ctx context.Context, userID int) ([]model.Withdrawal, error){
	
}