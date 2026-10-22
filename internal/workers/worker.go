package workers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/go-resty/resty/v2"
)

var errOrderNotRegistered = errors.New("no order in acrual system")
var errTooManyRequests = errors.New("too many requests to accrual system")

type LoyaltyService interface {
	GetProcessedOrders(ctx context.Context) ([]model.Order, error)
	UpdateOrder(ctx context.Context, res model.AccrualSystemResponce, userID int64)  error
}

func Run(ctx context.Context, loyalty LoyaltyService, address string, ticker *time.Ticker) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            process(ctx, loyalty, address)
        }
    }
}

func process(ctx context.Context, loyalty LoyaltyService, address string) {
    tickCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
    defer cancel()
    orders, err := loyalty.GetProcessedOrders(tickCtx)
    if err != nil {
        return
    }
    for _, order := range orders {
        if tickCtx.Err() != nil {
            return
        }
        res, err := sendRequestToAccrualSystem(tickCtx, address, order.Number)
        if err != nil {
            continue
        }
        if res.Status != order.Status && res.Status != "REGISTERED" {
            _ = loyalty.UpdateOrder(tickCtx, res, order.UserID)
        }
    }
}

func sendRequestToAccrualSystem(ctx context.Context, address string, orderNumber string) (model.AccrualSystemResponce, error) {
	var out model.AccrualSystemResponce
	client := resty.New()
	client.
		SetRetryCount(5).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r != nil && r.StatusCode() == http.StatusTooManyRequests
		}).
		SetRetryAfter(func(_ *resty.Client, resp *resty.Response) (time.Duration, error) {
			if resp == nil {
				return 60 * time.Second, nil
			}
			sec, err := strconv.Atoi(resp.Header().Get("Retry-After"))
			if err != nil {
				return 60 * time.Second, nil
			}
			return time.Duration(sec) * time.Second, nil
		})

	url := fmt.Sprintf("http://%s/api/orders/%s", address, orderNumber)

	resp, err := client.R().SetContext(ctx).SetResult(&out).Get(url)

	if err != nil {
		return model.AccrualSystemResponce{}, err
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return out, nil
	case http.StatusNoContent:
		return model.AccrualSystemResponce{}, errOrderNotRegistered
	case http.StatusTooManyRequests:
		return model.AccrualSystemResponce{}, errTooManyRequests
	default:
		return model.AccrualSystemResponce{}, fmt.Errorf("accrual status %d", resp.StatusCode())
	}
}