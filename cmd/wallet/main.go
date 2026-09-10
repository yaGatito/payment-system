package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"

	gqladp "payment-system/graphql"
	rozetkaclient "payment-system/internal/3rdparty"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/app"
	"payment-system/sql/sqlcgen"
)

type EnvConfig struct {
	DbUser string `env:"WALLET_DB_USER,notEmpty"`
	DbPass string `env:"WALLET_DB_PASS,notEmpty"`
	DbHost string `env:"WALLET_DB_HOST,notEmpty"`
	DbPort string `env:"WALLET_DB_PORT,notEmpty"`
	DbName string `env:"WALLET_DB_NAME,notEmpty"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	testId1, _ := uuid.NewRandom()
	testId2, _ := uuid.NewRandom()
	testId3, _ := uuid.NewRandom()
	fmt.Println(testId1.String())
	fmt.Println(testId2.String())
	fmt.Println(testId3.String())

	baseURL := "https://api.rozetkapay.com"
	username := "a6a29002-dc68-4918-bc5d-51a6094b14a8"
	password := "XChz3J8qrr"
	redirectURL := "https://tidy-worsening-womanless.ngrok-free.dev/notify"

	cfg := EnvConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("parsing config failed: %v", err)
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
	rozetkaClient := rozetkaclient.NewClient(baseURL, username, password, redirectURL)
	walletService := app.NewWalletService(walletRepo, rozetkaClient)

	schema, err := gqladp.NewSchema(walletService)
	if err != nil {
		log.Fatalf("graphql schema: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	mux := http.NewServeMux()
	mux.Handle("/graphql", h)

	addr := ":8082"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	log.Printf("wallet graphql listening on %s/graphql", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}

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
