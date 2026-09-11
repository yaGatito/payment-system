package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v11"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v5/pgxpool"

	gqladp "payment-system/graphql"
	rozetkaclient "payment-system/internal/adapters/3rdparty"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/app"
	"payment-system/pkg/logger"
	"payment-system/sql/sqlcgen"
)

type EnvConfig struct {
	DbUser            string `env:"WALLET_DB_USER,notEmpty"`
	DbPass            string `env:"WALLET_DB_PASS,notEmpty"`
	DbHost            string `env:"WALLET_DB_HOST,notEmpty"`
	DbPort            string `env:"WALLET_DB_PORT,notEmpty"`
	DbName            string `env:"WALLET_DB_NAME,notEmpty"`
	RozetkaBaseURL    string `env:"ROZETKA_BASE_URL,notEmpty"`
	RozetkaUsername   string `env:"ROZETKA_USERNAME,notEmpty"`
	RozetkaPassword   string `env:"ROZETKA_PASSWORD,notEmpty"`
	RozetkaCallback   string `env:"ROZETKA_CALLBACK_URL,notEmpty"`
	GraphQLListenAddr string `env:"GRAPHQL_LISTEN_ADDR,notEmpty"`
}

func main() {
	l := logger.New()
	if err := run(l); err != nil {
		log.Fatal(err)
	}
}

func run(l *logger.Logger) error {
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
	rozetkaClient := rozetkaclient.NewClient(cfg.RozetkaBaseURL, cfg.RozetkaUsername, cfg.RozetkaPassword, cfg.RozetkaCallback)
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

	l.Info("wallet graphql listening on %s/graphql", cfg.GraphQLListenAddr)
	if err := http.ListenAndServe(cfg.GraphQLListenAddr, mux); err != nil {
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
