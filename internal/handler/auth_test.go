package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KiberIGOR/Loyalty/internal/mocks"
	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/KiberIGOR/Loyalty/internal/repository"
	"github.com/KiberIGOR/Loyalty/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const testSecretKey = "test-secret"

func testPasswordHash(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func TestPostRegister(t *testing.T) {
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
		setup       func(m *mocks.MockUserRepository)
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
			data:        `{"login":"user","password":"password"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Create(gomock.Any(), "user", testPasswordHash("password")).
					Return(1, nil)
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
			name: "negative test #3 empty login",
			want: want{
				code:        http.StatusBadRequest,
				response:    http.StatusText(http.StatusBadRequest) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"","password":"password"}`,
		},
		{
			name: "negative test #4 empty password",
			want: want{
				code:        http.StatusBadRequest,
				response:    http.StatusText(http.StatusBadRequest) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"user","password":""}`,
		},
		{
			name: "negative test #5 login taken",
			want: want{
				code:        http.StatusConflict,
				response:    http.StatusText(http.StatusConflict) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"user","password":"password"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Create(gomock.Any(), "user", testPasswordHash("password")).
					Return(0, repository.ErrConflict)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUserRepository(ctrl)
			if test.setup != nil {
				test.setup(m)
			}

			authService := service.NewAuthService(m, testSecretKey)
			h := NewAuth(authService)

			request := httptest.NewRequest(test.method, "/api/user/register", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			w := httptest.NewRecorder()
			h.PostRegister(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))

			if test.want.code == http.StatusOK {
				authHeader := res.Header.Get("Authorization")
				require.True(t, strings.HasPrefix(authHeader, "Bearer "),
					"Authorization should start with Bearer, got %q", authHeader)
				token := strings.TrimPrefix(authHeader, "Bearer ")
				assert.NotEmpty(t, token)
				assert.Equal(t, 1, authService.GetUserID(request.Context(), token))
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}

func TestPostAuth(t *testing.T) {
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
		setup       func(m *mocks.MockUserRepository)
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
			data:        `{"login":"user","password":"password"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByLogin(gomock.Any(), "user").
					Return(&model.User{
						ID:           1,
						Login:        "user",
						PasswordHash: testPasswordHash("password"),
					}, nil)
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
			name: "negative test #3 empty login",
			want: want{
				code:        http.StatusBadRequest,
				response:    http.StatusText(http.StatusBadRequest) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"","password":"password"}`,
		},
		{
			name: "negative test #4 user not found",
			want: want{
				code:        http.StatusUnauthorized,
				response:    http.StatusText(http.StatusUnauthorized) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"user","password":"password"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByLogin(gomock.Any(), "user").
					Return(nil, repository.ErrNotFound)
			},
		},
		{
			name: "negative test #5 wrong password",
			want: want{
				code:        http.StatusUnauthorized,
				response:    http.StatusText(http.StatusUnauthorized) + "\n",
				contentType: "text/plain; charset=utf-8",
			},
			method:      http.MethodPost,
			contentType: "application/json",
			data:        `{"login":"user","password":"wrong"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByLogin(gomock.Any(), "user").
					Return(&model.User{
						ID:           1,
						Login:        "user",
						PasswordHash: testPasswordHash("password"),
					}, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUserRepository(ctrl)
			if test.setup != nil {
				test.setup(m)
			}

			authService := service.NewAuthService(m, testSecretKey)
			h := NewAuth(authService)

			request := httptest.NewRequest(test.method, "/api/user/login", strings.NewReader(test.data))
			request.Header.Set("content-type", test.contentType)
			w := httptest.NewRecorder()
			h.PostAuth(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))

			if test.want.code == http.StatusOK {
				authHeader := res.Header.Get("Authorization")
				require.True(t, strings.HasPrefix(authHeader, "Bearer "),
					"Authorization should start with Bearer, got %q", authHeader)
				token := strings.TrimPrefix(authHeader, "Bearer ")
				assert.NotEmpty(t, token)
				assert.Equal(t, 1, authService.GetUserID(request.Context(), token))
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("content-type"))
			}
		})
	}
}
