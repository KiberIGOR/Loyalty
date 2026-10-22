package service

import (
	"context"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/model"
)

type OrderRepository interface {
	// Save сохраняет заказ пользователя по данным системы расчёта начислений.
	Save(ctx context.Context, res model.Order) error
	// GetByNumber возвращает заказ по его номеру.
	GetByNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	// GetByUser возвращает все заказы пользователя, от новых к старым.
	GetByUser(ctx context.Context, userID int) ([]model.Order, error)
	// GetProcessedOrders возвращает все заказы в статусе "NEW" и "PROCESSING".
	GetProcessedOrders(ctx context.Context) ([]model.Order, error)
	//UpdateOrder обновляет статус order 
	UpdateOrder(ctx context.Context, res model.Order)  error
	//AccrueProcessed транзакция обновляет order и начисляет бонусы пользователю
	AccrueProcessed(ctx context.Context, res model.Order, userID int) error
}

type WithdrawalRepository interface {
	// GetBalance возвращает сумму всех списаний пользователя.
	GetBalance(ctx context.Context, userID int) (int64, error)
	// GetByUser возвращает все списания пользователя, от новых к старым.
	GetByUser(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

type LoyaltyService struct {
	userRepo             UserRepository
	orderRepo            OrderRepository
	withdrawalRepo       WithdrawalRepository
	AccrualSystemAddress string
}

func NewLoyaltyService(userRepo UserRepository, orderRepo OrderRepository, withdrawalRepo WithdrawalRepository, accrualSystemAddress string) *LoyaltyService {
	return &LoyaltyService{
		userRepo:             userRepo,
		orderRepo:            orderRepo,
		withdrawalRepo:       withdrawalRepo,
		AccrualSystemAddress: accrualSystemAddress,
	}
}

func (a *LoyaltyService) SaveOrder(ctx context.Context, orderNumber string, userID int) error {
	ok := a.checkNumber(orderNumber)
	if !ok {
	}
	data, err := a.orderRepo.GetByNumber(ctx, orderNumber)
	if data.Number != "" {
		if data.UserID != int64(userID){
		}
	}
	var order = model.Order{
		Number: orderNumber,
		Status: "NEW",
		Accrual: 0,
		UploadedAt: time.Now(),
		UserID: int64(userID),
	}
	if err != nil {
	}
	err = a.orderRepo.Save(ctx, order)
	if err != nil {
	}
}
func (a *LoyaltyService) GetOrders(ctx context.Context, userID int) ([]model.OrderResponce, error) {
	res, err := a.orderRepo.GetByUser(ctx, userID)
	out := make([]model.OrderResponce, len(res))
	for i, item := range res {
		out[i] = model.OrderResponce{
			Number: item.Number,
			Status: item.Status,
			Accrual: item.Accrual,
			UploadedAt: item.UploadedAt.Format(time.RFC3339),
		}
	}
}
func (a *LoyaltyService) GetBalance(ctx context.Context, userID int) ([]model.BalanceResponce, error) {
	res, err := a.userRepo.GetBalance(ctx, userID)
	res2, err := a.withdrawalRepo.GetBalance(ctx, userID)
}
func (a *LoyaltyService) ReduceBalance(ctx context.Context, orderNumber string, sum int64, userID int) error {
	ok := a.checkNumber(orderNumber)
	if !ok {
	}
	if sum < 0 {
	}
	res, err := a.userRepo.GetBalance(ctx, userID)
	if err != nil {
	}
	if res < sum {
	}
	//В одной транзакции будет и уменьшение баланса и информации о списании
	err = a.userRepo.ReduceBalance(ctx, orderNumber, sum, userID)
	if err != nil {
	}
}
func (a *LoyaltyService) GetWithdrawals(ctx context.Context, userID int) ([]model.WithdrawalsResponce, error) {
	res, err := a.withdrawalRepo.GetByUser(ctx, userID)
	if err != nil {
	}
	out := make([]model.WithdrawalsResponce, len(res))
	for i, item := range res {
		out[i] = model.WithdrawalsResponce{
			Order:       item.Order,
			Sum:         item.Sum,
			ProcessedAt: item.ProcessedAt.Format(time.RFC3339),
		}
	}

	return out, nil
}

func (a *LoyaltyService) checkNumber(orderNumber string) bool {
	if orderNumber == "" {
		return false
	}

	sum := 0
	double := false
	for i := len(orderNumber) - 1; i >= 0; i-- {
		c := orderNumber[i]
		if c < '0' || c > '9' {
			return false
		}

		n := int(c - '0')
		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}

	return sum%10 == 0
}

func (a *LoyaltyService) GetProcessedOrders(ctx context.Context) ([]model.Order, error) {
	res, err := a.orderRepo.GetProcessedOrders(ctx)

	return res, err 
}
func (a *LoyaltyService) UpdateOrder(ctx context.Context, res model.AccrualSystemResponce, userID int64) error {
	switch res.Status {
	case "REGISTERED":
		return nil
	case "INVALID", "PROCESSING":
		return a.orderRepo.UpdateOrder(ctx, model.Order{
			Number: res.Order,
			Status: res.Status,
		})
	case "PROCESSED":
		//транзакция на обновление order и на начисление бонусов пользовтелю 
		if err := a.orderRepo.AccrueProcessed(ctx, model.Order{
			Number:  res.Order,
			Status:  res.Status,
			Accrual: res.Accrual,
		}, int(userID)); err != nil {
			return err
		}
	default:
		return nil
	}
}


