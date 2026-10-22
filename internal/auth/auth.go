package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const KeyUserID ctxKey = "user_id"

type AuthService interface {
	GetUserID(ctx context.Context, tokenString string) int
}

var auth AuthService

func Initialize(a AuthService) {
	auth = a
}

func tokenFromHeader(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromHeader(r)
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		id := auth.GetUserID(r.Context(), token)
		if id <= 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), KeyUserID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
