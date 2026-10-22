package repository

import (
	"context"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}


	func (r *OrderRepo) Save(ctx context.Context, res model.Order) error {
		
	}

	func (r *OrderRepo) GetByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
		
	}

	func (r *OrderRepo) GetByUser(ctx context.Context, userID int) ([]model.Order, error) {
		
	}

	func (r *OrderRepo) GetProcessedOrders(ctx context.Context) ([]model.Order, error) {
		
	}

	func (r *OrderRepo) UpdateOrder(ctx context.Context, res model.Order)  error {
		
	}

	func (r *OrderRepo) AccrueProcessed(ctx context.Context, res model.Order, userID int) error {
		
	}