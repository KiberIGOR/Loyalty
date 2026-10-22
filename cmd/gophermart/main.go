package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KiberIGOR/Loyalty/internal/auth"
	"github.com/KiberIGOR/Loyalty/internal/compress"
	"github.com/KiberIGOR/Loyalty/internal/config"
	"github.com/KiberIGOR/Loyalty/internal/handler"
	"github.com/KiberIGOR/Loyalty/internal/logger"
	"github.com/KiberIGOR/Loyalty/internal/repository"
	"github.com/KiberIGOR/Loyalty/internal/service"
	"github.com/KiberIGOR/Loyalty/internal/workers"
	"github.com/KiberIGOR/Loyalty/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

	cfg := config.Parse()

	if err := logger.Initialize("info"); err != nil {
		log.Fatal(err)
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := runMigrations(cfg.DataBaseDSN); err != nil {
		log.Fatal("migrate: ", err)
	}

	pool, err := newPool(initCtx, cfg.DataBaseDSN)
	if err != nil {
		log.Fatal("pool: ", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	authService := service.NewAuthService(userRepo, cfg.SecretKey)
	auth.Initialize(authService)
	h := handler.NewAuth(authService)

	ordersRepo := repository.NewOrderRepo(pool)
	withdrawalRepo := repository.NewWithdrawalRepo(pool)
	LoyaltyService := service.NewLoyaltyService(userRepo,ordersRepo,withdrawalRepo,cfg.AccrualSystemAddress)
	h2 := handler.New(LoyaltyService)

	ticker := time.NewTicker(20*time.Second)
	defer ticker.Stop()
	go workers.Run(ctx, LoyaltyService, cfg.AccrualSystemAddress, ticker)
	
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)

	r.Post("/api/user/register", h.PostRegister)
	r.Post("/api/user/login", h.PostAuth)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/api/user/orders", h2.PostSaveOrder)
		r.Get("/api/user/orders", h2.GetOrders)
		r.Get("/api/user/balance", h2.GetBalance)
		r.Post("/api/user/balance/withdraw", h2.PostBalance)
		r.Get("/api/user/withdrawals", h2.GetWithdrawals)
	})

	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 10 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
