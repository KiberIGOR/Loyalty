package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/auth"
	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/KiberIGOR/Loyalty/internal/service"
)

type Handler struct {
	loyalty LoyaltyService
}

func New(loyalty LoyaltyService) *Handler {
	return &Handler{loyalty: loyalty}
}

type LoyaltyService interface {
	SaveOrder(ctx context.Context, orderNumber string, userID int) error
	GetOrders(ctx context.Context, userID int) ([]model.OrderResponce, error)
	GetBalance(ctx context.Context, userID int) (model.BalanceResponce, error)
	ReduceBalance(ctx context.Context, orderNumber string, sum float64, userID int) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.WithdrawalsResponce, error)
}

func (h *Handler) PostSaveOrder(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostSaveOrder error: no user id"))
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		respondBadRequest(res, err)
		return
	}
	defer req.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		respondBadRequest(res, errors.New("empty order number"))
		return
	}

	err = h.loyalty.SaveOrder(ctx, orderNumber, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(res, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrOrderAlreadyExists):
			res.WriteHeader(http.StatusOK)
		case errors.Is(err, service.ErrOrderConflict):
			http.Error(res, http.StatusText(http.StatusConflict), http.StatusConflict)
		default:
			respondInternalError(res, err)
		}
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("GetOrders error: no user id"))
		return
	}

	orders, err := h.loyalty.GetOrders(ctx, userID)
	if err != nil {
		respondInternalError(res, err)
		return
	}
	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(res).Encode(orders); err != nil {
		respondInternalError(res, err)
	}
}

func (h *Handler) GetBalance(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("GetBalance error: no user id"))
		return
	}

	balance, err := h.loyalty.GetBalance(ctx, userID)
	if err != nil {
		respondInternalError(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(res).Encode(balance); err != nil {
		respondInternalError(res, err)
	}
}

func (h *Handler) PostBalance(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostBalance error: no user id"))
		return
	}

	var body model.ReduceBalanceRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondBadRequest(res, err)
		return
	}
	body.Order = strings.TrimSpace(body.Order)

	err := h.loyalty.ReduceBalance(ctx, body.Order, body.Sum, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(res, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrInsufficientFunds):
			http.Error(res, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		default:
			respondInternalError(res, err)
		}
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("GetWithdrawals error: no user id"))
		return
	}

	withdrawals, err := h.loyalty.GetWithdrawals(ctx, userID)
	if err != nil {
		respondInternalError(res, err)
		return
	}
	if len(withdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(res).Encode(withdrawals); err != nil {
		respondInternalError(res, err)
	}
}
