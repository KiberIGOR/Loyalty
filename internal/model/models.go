package model

import "time"

// AuthRequest — тело запросов POST /api/user/register и POST /api/user/login.
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// User — пользователь системы лояльности.
type User struct {
	ID           int
	Login        string
	PasswordHash string
}

type OrderResponce struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int64  `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}
type BalanceResponce struct {
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}

type ReduceBalanceRequest struct {
	Order string `json:"order"`
	Sum   int64  `json:"sum"`
}
type WithdrawalsResponce struct {
	Order       string `json:"order"`
	Sum         int64  `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
type AccrualSystemResponce struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int64  `json:"accrual"`
}

type Order struct {
	Number     string
	Status     string
	Accrual    int64
	UploadedAt time.Time
	UserID     int64
}

type Withdrawal struct {
	Order       string
	Sum         int64
	ProcessedAt time.Time
}
