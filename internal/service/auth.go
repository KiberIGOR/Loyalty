package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/KiberIGOR/Loyalty/internal/repository"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrLoginTaken         = errors.New("login already taken")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type UserRepository interface {
	//Create регистрация
	Create(ctx context.Context, login, passwordHash string) (int, error)
	//GetByLogin получение данных для авторизации
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	//GetBalance возвращает баланс накопленных балов пользователя 
	GetBalance(ctx context.Context, userID int) (int64, error)
	//ReduceBalance транзакция уменьшает баланс и записывает информацию о списании
	ReduceBalance(ctx context.Context, orderNumber string, sum int64, userID int) error
}

type AuthService struct {
	repo      UserRepository
	secretKey string
}

func NewAuthService(repo UserRepository, key string) *AuthService {
	return &AuthService{
		repo:      repo,
		secretKey: key,
	}
}

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TokenExp = time.Hour * 24

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (a *AuthService) BuildJWTString(ctx context.Context, id int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: id,
	})
	// создаём строку токена
	tokenString, err := token.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", err
	}
	// возвращаем строку токена
	return tokenString, nil
}

func (a *AuthService) GetUserID(ctx context.Context, tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.secretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		return -1
	}
	return claims.UserID
}

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func compareHashAndPassword(storedHash, password string) bool {
	computed := hashPassword(password)
	if len(storedHash) != len(computed) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(computed)) == 1
}

func (a *AuthService) Register(ctx context.Context, login, password string) (int, error) {
	id, err := a.repo.Create(ctx, login, hashPassword(password))
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return 0, ErrLoginTaken
		}
		return 0, err
	}
	return id, nil
}

func (a *AuthService) Auth(ctx context.Context, login, password string) (int, error) {
	user, err := a.repo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	if !compareHashAndPassword(user.PasswordHash, password) {
		return 0, ErrInvalidCredentials
	}
	return user.ID, nil
}
