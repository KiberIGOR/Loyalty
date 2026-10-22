package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Address              string
	DataBaseDSN          string
	SecretKey            string
	AccrualSystemAddress string
}

func Parse() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port for redirect url")
	ps := fmt.Sprintf("postgres://%s:%s@localhost:5432/loyalty?sslmode=disable", "loyalty", "loyalty")
	flag.StringVar(&cfg.DataBaseDSN, "d", ps, "dsn string for database setting")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "localhost:8081", "Address of api charge calculation system")
	flag.StringVar(&cfg.SecretKey, "k", "J83YbUemqzZBG1iJoNz8", "Secret Key for JWT")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}

	if envRunAddr := os.Getenv("DATABASE_URI"); envRunAddr != "" {
		cfg.DataBaseDSN = envRunAddr
	}

	if envRunAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envRunAddr != "" {
		cfg.DataBaseDSN = envRunAddr
	}

	return cfg
}
