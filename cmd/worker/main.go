package main

import (
	"context"
	"fmt"
	"log"
	"time"

	natsadp "payment-system/internal/adapters/nats"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/app"
	"payment-system/sql/sqlcgen"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
)

type EnvConfig struct {
	DbUser     string `env:"WALLET_DB_USER,notEmpty"`
	DbPass     string `env:"WALLET_DB_PASS,notEmpty"`
	DbHost     string `env:"WALLET_DB_HOST,notEmpty"`
	DbPort     string `env:"WALLET_DB_PORT,notEmpty"`
	DbName     string `env:"WALLET_DB_NAME,notEmpty"`
	NatsURL    string `env:"NATS_URL,notEmpty"`
	NatsStream string `env:"NATS_STREAM,notEmpty"`
}

const durableName = "PAYMENT_WORKER"

const (
	PendingPaymentStatus = "pending"
	FailurePaymentStatus = "failure"
	SuccessPaymentStatus = "success"
)

// 1. app.Service usage (save card to DB)
// 2. app.Service usage (pay with card, and save payment into DB OR update its status in DB)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	cfg := EnvConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return fmt.Errorf("parsing config failed: %w", err)
	}

	pgConfig, err := pgxpool.ParseConfig(dbURL(cfg))
	if err != nil {
		return fmt.Errorf("failed to parse db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	walletRepo := postgres.NewWalletRepoPostgreSQL(sqlcgen.New(pool))
	paymentWorkerSvc := app.NewPaymentWorkerService(walletRepo)
	messageHandler := natsadp.NewPaymentMessageHandler(paymentWorkerSvc)

	natsClient, err := natsadp.New(cfg.NatsURL, cfg.NatsStream)
	if err != nil {
		return fmt.Errorf("failed to initialize nats js client: %w", err)
	}

	natsClient.Subscribe(ctx, cfg.NatsStream, durableName, func(msg jetstream.Msg) {
		messageHandler.HandleMessage(ctx, msg)
	})

	log.Printf("Started worker")

	time.Sleep(time.Hour)
	select {}
}

func dbURL(cfg EnvConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&pool_max_conn_lifetime=%s",
		cfg.DbUser,
		cfg.DbPass,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
		"disable",
		10,
		"1h30m")
}
