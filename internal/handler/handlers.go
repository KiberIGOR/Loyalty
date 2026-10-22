package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/auth"
	"github.com/KiberIGOR/Loyalty/internal/model"
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
	GetBalance(ctx context.Context, userID int) ([]model.BalanceResponce, error)
	ReduceBalance(ctx context.Context, orderNumber string, sum int64, userID int) error
	GetWithdrawals(ctx context.Context, userID int) ([]model.WithdrawalsResponce, error)
}

func (h *Handler) PostSaveOrder(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}

	var a string = "12345678903"

	err := h.loyalty.SaveOrder(ctx, a, userID)
	if err != nil {

	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) GetOrders(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}
	orders, err := h.loyalty.GetOrders(ctx, userID)
	if err != nil {

	}
	byteOrders, err := json.Marshal(orders)
	if err != nil {

	}
	if _, err = res.Write([]byte(byteOrders)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) GetBalance(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}
	balance, err := h.loyalty.GetBalance(ctx, userID)
	if err != nil {

	}
	bytebalance, err := json.Marshal(balance)
	if err != nil {

	}
	if _, err = res.Write([]byte(bytebalance)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) PostBalance(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}
	a := model.ReduceBalanceRequest{
		Order: "2377225624",
		Sum:   751,
	}
	err := h.loyalty.ReduceBalance(ctx, a.Order, a.Sum, userID)
	if err != nil {

	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	userID, ok := req.Context().Value(auth.KeyUserID).(int)
	if !ok {
		respondInternalError(res, errors.New("PostJSONURLHandler error: no user id"))
		return
	}
	withdrawals, err := h.loyalty.GetWithdrawals(ctx, userID)
	if err != nil {

	}
	byteWithdrawals, err := json.Marshal(withdrawals)
	if err != nil {

	}
	if _, err = res.Write([]byte(byteWithdrawals)); err != nil {
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}
