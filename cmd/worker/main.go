package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	natsadp "payment-system/internal/adapters/nats"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/app"
	"payment-system/pkg/logger"
	"payment-system/sql/sqlcgen"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
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

func main() {
	l := logger.New()
	if err := run(l); err != nil {
		log.Fatal(err)
	}
}

func run(l *logger.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	defer pool.Close()

	walletRepo := postgres.NewWalletRepoPostgreSQL(sqlcgen.New(pool))
	paymentWorkerSvc := app.NewPaymentWorkerService(walletRepo)
	messageHandler, err := natsadp.NewPaymentMessageHandler(paymentWorkerSvc, l)
	if err != nil {
		return fmt.Errorf("init payment message handler: %w", err)
	}

	natsClient, err := natsadp.New(cfg.NatsURL, cfg.NatsStream)
	if err != nil {
		return fmt.Errorf("failed to initialize nats js client: %w", err)
	}
	defer natsClient.Close()

	consumeCtx, err := natsClient.Subscribe(ctx, cfg.NatsStream, durableName, messageHandler.HandleMessage)
	if err != nil {
		return fmt.Errorf("subscribe to nats stream: %w", err)
	}
	defer func() {
		if consumeCtx != nil {
			consumeCtx.Stop()
		}
	}()

	l.Info("started worker")

	<-ctx.Done()
	l.Info("shutdown signal received, stopping worker")
	return nil
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
