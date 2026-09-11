package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	natsadp "payment-system/internal/adapters/nats"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/app"
	"payment-system/internal/domain"
	"payment-system/sql/sqlcgen"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
)

type EnvConfig struct {
	DbUser string `env:"WALLET_DB_USER,notEmpty"`
	DbPass string `env:"WALLET_DB_PASS,notEmpty"`
	DbHost string `env:"WALLET_DB_HOST,notEmpty"`
	DbPort string `env:"WALLET_DB_PORT,notEmpty"`
	DbName string `env:"WALLET_DB_NAME,notEmpty"`
}

const durableName = "PAYMENT_WORKER"

const (
	AddCardEventType = "add-card"
	PaymentEventType = "payment"

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

	natsClient, err := natsadp.New(natsadp.NatsURL, natsadp.StreamPayments)
	if err != nil {
		return fmt.Errorf("failed to initialize nats js client: %w", err)
	}

	natsClient.Subscribe(ctx, natsadp.StreamPayments, durableName, func(msg jetstream.Msg) {
		var event natsadp.PaymentEvent
		err := json.Unmarshal(msg.Data(), &event)
		if err != nil {
			log.Printf("failed to deserialize event data: %s", string(msg.Data()))
		}

		switch event.EventType {
		case AddCardEventType:
			if event.Status == SuccessPaymentStatus {
				err := paymentWorkerSvc.SaveCard(context.Background(), domain.Card{
					CustomerID: uuid.MustParse(event.CustomerID),
					Token:      event.CcToken,
					Type:       event.PaymentSystem,
					// TODO: cut last4
					Last4: event.CardMask,
				})
				if err != nil {
					log.Printf("failed to save card: %v", err)
				}
			}

		case PaymentEventType:
			// TODO: populate fields required to save payment in callback-service
			// paymentWorkerSvc.SavePayment()
		}

		msg.Ack()
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
