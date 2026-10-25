package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/KiberIGOR/Loyalty/internal/logger"
	"github.com/KiberIGOR/Loyalty/internal/model"
	"github.com/KiberIGOR/Loyalty/internal/service"
	"go.uber.org/zap"
)

var errInvalidAuthRequest = errors.New("invalid auth request")

func respondInternalError(res http.ResponseWriter, err error) {
	logger.Log.Error("internal server error", zap.Error(err))
	http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func respondBadRequest(res http.ResponseWriter, err error) {
	logger.Log.Error("bad request", zap.Error(err))
	http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

type AuthService interface {
	BuildJWTString(ctx context.Context, id int) (string, error)
	Auth(ctx context.Context, login string, password string) (int, error)
	Register(ctx context.Context, login string, password string) (int, error)
}

type AuthHandler struct {
	auth AuthService
}

func NewAuth(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func parseAuthRequest(req *http.Request) (model.AuthRequest, error) {
	defer req.Body.Close()

	var creds model.AuthRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&creds); err != nil {
		return model.AuthRequest{}, err
	}

	creds.Login = strings.TrimSpace(creds.Login)
	if creds.Login == "" || creds.Password == "" {
		return model.AuthRequest{}, errInvalidAuthRequest
	}
	return creds, nil
}

func (h *AuthHandler) setAuthorization(res http.ResponseWriter, req *http.Request, userID int) error {
	token, err := h.auth.BuildJWTString(req.Context(), userID)
	if err != nil {
		return err
	}
	res.Header().Set("Authorization", "Bearer "+token)
	return nil
}

func (h *AuthHandler) PostRegister(res http.ResponseWriter, req *http.Request) {
	creds, err := parseAuthRequest(req)
	if err != nil {
		respondBadRequest(res, err)
		return
	}

	id, err := h.auth.Register(req.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrLoginTaken) {
			http.Error(res, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		respondInternalError(res, err)
		return
	}

	if err := h.setAuthorization(res, req, id); err != nil {
		logger.Log.Error("failed to build JWT for new user", zap.Int("user_id", id), zap.Error(err))
		respondInternalError(res, err)
		return
	}

	logger.Log.Info("new user created", zap.Int("user_id", id))
	res.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) PostAuth(res http.ResponseWriter, req *http.Request) {
	creds, err := parseAuthRequest(req)
	if err != nil {
		respondBadRequest(res, err)
		return
	}

	id, err := h.auth.Auth(req.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		respondInternalError(res, err)
		return
	}

	if err := h.setAuthorization(res, req, id); err != nil {
		logger.Log.Error("failed to build JWT", zap.Int("user_id", id), zap.Error(err))
		respondInternalError(res, err)
		return
	}

	res.WriteHeader(http.StatusOK)
}
