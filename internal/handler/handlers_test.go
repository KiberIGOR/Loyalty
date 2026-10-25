package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/auth"
	"github.com/KiberIGOR/Loyalty/internal/mocks"
	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/KiberIGOR/Loyalty/internal/repository"
	"github.com/KiberIGOR/Loyalty/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testUserID      = 1
	validOrder      = "12345678903"
	anotherValidOrd = "9278923470"
	invalidOrder    = "12345678900"
)

func requestWithUserID(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), auth.KeyUserID, userID)
	return r.WithContext(ctx)
}

type loyaltyMocks struct {
	user       *mocks.MockUserRepository
	order      *mocks.MockOrderRepository
	withdrawal *mocks.MockWithdrawalRepository
}

func newLoyaltyHandler(t *testing.T, setup func(m loyaltyMocks)) *Handler {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := loyaltyMocks{
		user:       mocks.NewMockUserRepository(ctrl),
		order:      mocks.NewMockOrderRepository(ctrl),
		withdrawal: mocks.NewMockWithdrawalRepository(ctrl),
	}
	if setup != nil {
		setup(m)
	}
	return New(service.NewLoyaltyService(m.user, m.order, m.withdrawal))
}

func TestPostSaveOrder(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		contentType string
		data        string
		setup       func(m loyaltyMocks)
	}{
		{
			name: "positive test #1",
			want: want{
				code:        http.StatusAccepted,
				response:    "",
				contentType: "",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        validOrder,
			setup: func(m loyaltyMocks) {
				m.order.EXPECT().
					GetByNumber(gomock.Any(), validOrder).
					Return(nil, repository.ErrNotFound)
				m.order.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "positive test #2 already uploaded by this user",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        validOrder,
			setup: func(m loyaltyMocks) {
				m.order.EXPECT().
					GetByNumber(gomock.Any(), validOrder).
					Return(&model.Order{Number: validOrder, UserID: int64(testUserID)}, nil)
			},
		},
		{
			name: "negative test #3 no body",
			want: want{
				code:        http.StatusBadRequest,
				response:    http.StatusText(http.StatusBadRequest) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        "",
		},
		{
			name: "negative test #4 invalid order number",
			want: want{
				code:        http.StatusUnprocessableEntity,
				response:    http.StatusText(http.StatusUnprocessableEntity) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        invalidOrder,
		},
		{
			name: "negative test #5 order uploaded by another user",
			want: want{
				code:        http.StatusConflict,
				response:    http.StatusText(http.StatusConflict) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "text/plain",
			data:        validOrder,
			setup: func(m loyaltyMocks) {
				m.order.EXPECT().
					GetByNumber(gomock.Any(), validOrder).
					Return(&model.Order{Number: validOrder, UserID: 2}, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newLoyaltyHandler(t, test.setup)

			request := httptest.NewRequest(test.method, "/api/user/orders", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			request = requestWithUserID(request, testUserID)
			w := httptest.NewRecorder()
			h.PostSaveOrder(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	uploadedAt := time.Date(2026, 12, 10, 15, 15, 45, 0, time.UTC)

	type want struct {
		code        int
		orders      []model.OrderResponce
		contentType string
	}
	tests := []struct {
		name   string
		want   want
		method string
		setup  func(m loyaltyMocks)
	}{
		{
			name: "positive test #1",
			want: want{
				code: http.StatusOK,
				orders: []model.OrderResponce{
					{
						Number:     anotherValidOrd,
						Status:     "PROCESSED",
						Accrual:    500,
						UploadedAt: uploadedAt.Format(time.RFC3339),
					},
				},
				contentType: "application/json",
			},
			method: http.MethodGet,
			setup: func(m loyaltyMocks) {
				m.order.EXPECT().
					GetByUser(gomock.Any(), testUserID).
					Return([]model.Order{
						{
							Number:     anotherValidOrd,
							Status:     "PROCESSED",
							Accrual:    500,
							UploadedAt: uploadedAt,
							UserID:     int64(testUserID),
						},
					}, nil)
			},
		},
		{
			name: "positive test #2 no content",
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
			},
			method: http.MethodGet,
			setup: func(m loyaltyMocks) {
				m.order.EXPECT().
					GetByUser(gomock.Any(), testUserID).
					Return(nil, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newLoyaltyHandler(t, test.setup)

			request := httptest.NewRequest(test.method, "/api/user/orders", nil)
			request = requestWithUserID(request, testUserID)
			w := httptest.NewRecorder()
			h.GetOrders(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if test.want.code == http.StatusOK {
				var got []model.OrderResponce
				require.NoError(t, json.Unmarshal(resBody, &got))
				assert.Equal(t, test.want.orders, got)
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			} else if test.want.contentType != "" {
				assert.Equal(t, http.StatusText(test.want.code)+"\n", string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	type want struct {
		code        int
		balance     model.BalanceResponce
		contentType string
	}
	tests := []struct {
		name   string
		want   want
		method string
		setup  func(m loyaltyMocks)
	}{
		{
			name: "positive test #1",
			want: want{
				code: http.StatusOK,
				balance: model.BalanceResponce{
					Current:   500.5,
					Withdrawn: 42,
				},
				contentType: "application/json",
			},
			method: http.MethodGet,
			setup: func(m loyaltyMocks) {
				m.user.EXPECT().
					GetBalance(gomock.Any(), testUserID).
					Return(500.5, nil)
				m.withdrawal.EXPECT().
					GetBalance(gomock.Any(), testUserID).
					Return(42.0, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newLoyaltyHandler(t, test.setup)

			request := httptest.NewRequest(test.method, "/api/user/balance", nil)
			request = requestWithUserID(request, testUserID)
			w := httptest.NewRecorder()
			h.GetBalance(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if test.want.code == http.StatusOK {
				var got model.BalanceResponce
				require.NoError(t, json.Unmarshal(resBody, &got))
				assert.Equal(t, test.want.balance, got)
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			} else if test.want.contentType != "" {
				assert.Equal(t, http.StatusText(test.want.code)+"\n", string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestPostBalance(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		want        want
		method      string
		contentType string
		data        string
		setup       func(m loyaltyMocks)
	}{
		{
			name: "positive test #1",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"order":"` + validOrder + `","sum":751}`,
			setup: func(m loyaltyMocks) {
				m.user.EXPECT().
					ReduceBalance(gomock.Any(), validOrder, 751.0, testUserID).
					Return(nil)
			},
		},
		{
			name: "negative test #2 no body",
			want: want{
				code:        http.StatusBadRequest,
				response:    http.StatusText(http.StatusBadRequest) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        "",
		},
		{
			name: "negative test #3 invalid order number",
			want: want{
				code:        http.StatusUnprocessableEntity,
				response:    http.StatusText(http.StatusUnprocessableEntity) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"order":"` + invalidOrder + `","sum":751}`,
		},
		{
			name: "negative test #4 insufficient funds",
			want: want{
				code:        http.StatusPaymentRequired,
				response:    http.StatusText(http.StatusPaymentRequired) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"order":"` + validOrder + `","sum":751}`,
			setup: func(m loyaltyMocks) {
				m.user.EXPECT().
					ReduceBalance(gomock.Any(), validOrder, 751.0, testUserID).
					Return(repository.ErrInsufficientFunds)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newLoyaltyHandler(t, test.setup)

			request := httptest.NewRequest(test.method, "/api/user/balance/withdraw", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			request = requestWithUserID(request, testUserID)
			w := httptest.NewRecorder()
			h.PostBalance(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	processedAt := time.Date(2020, 12, 9, 16, 9, 57, 0, time.UTC)

	type want struct {
		code        int
		items       []model.WithdrawalsResponce
		contentType string
	}
	tests := []struct {
		name   string
		want   want
		method string
		setup  func(m loyaltyMocks)
	}{
		{
			name: "positive test #1",
			want: want{
				code: http.StatusOK,
				items: []model.WithdrawalsResponce{
					{
						Order:       anotherValidOrd,
						Sum:         500,
						ProcessedAt: processedAt.Format(time.RFC3339),
					},
				},
				contentType: "application/json",
			},
			method: http.MethodGet,
			setup: func(m loyaltyMocks) {
				m.withdrawal.EXPECT().
					GetByUser(gomock.Any(), testUserID).
					Return([]model.Withdrawal{
						{
							Order:       anotherValidOrd,
							Sum:         500,
							ProcessedAt: processedAt,
						},
					}, nil)
			},
		},
		{
			name: "positive test #2 no content",
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
			},
			method: http.MethodGet,
			setup: func(m loyaltyMocks) {
				m.withdrawal.EXPECT().
					GetByUser(gomock.Any(), testUserID).
					Return(nil, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newLoyaltyHandler(t, test.setup)

			request := httptest.NewRequest(test.method, "/api/user/withdrawals", nil)
			request = requestWithUserID(request, testUserID)
			w := httptest.NewRecorder()
			h.GetWithdrawals(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if test.want.code == http.StatusOK {
				var got []model.WithdrawalsResponce
				require.NoError(t, json.Unmarshal(resBody, &got))
				assert.Equal(t, test.want.items, got)
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			} else if test.want.contentType != "" {
				assert.Equal(t, http.StatusText(test.want.code)+"\n", string(resBody))
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}
