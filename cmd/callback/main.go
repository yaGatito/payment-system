package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	fasthttpadp "payment-system/internal/adapters/fasthttp"
	natsadp "payment-system/internal/adapters/nats"
	"payment-system/pkg/logger"

	"github.com/caarlos0/env/v11"
	"github.com/valyala/fasthttp"
)

const shutdownTimeout = 10 * time.Second

type EnvConfig struct {
	CallbackPort   string `env:"CALLBACK_PORT,notEmpty"`
	NatsURL        string `env:"NATS_URL,notEmpty"`
	NatsStream     string `env:"NATS_STREAM,notEmpty"`
	NatsSubject    string `env:"NATS_SUBJECT,notEmpty"`
	CallbackSecret string `env:"ROZETKA_PASSWORD,notEmpty"`
}

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
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("parsing callback config failed: %w", err)
	}

	natsClient, err := natsadp.New(cfg.NatsURL, cfg.NatsStream)
	if err != nil {
		return fmt.Errorf("init nats client: %w", err)
	}
	defer natsClient.Close()

	callbackHandler := fasthttpadp.NewCallbackHandler(
		natsClient,
		cfg.NatsSubject,
		cfg.CallbackSecret,
		l,
	)

	server := &fasthttp.Server{
		Handler: callbackHandler.NotifyHandler,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe(":" + cfg.CallbackPort)
	}()

	l.Info("notify server listening on %s", cfg.CallbackPort)

	select {
	case <-ctx.Done():
		l.Info("shutdown signal received, stopping callback service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("callback shutdown failed: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("callback server failed: %w", err)
		}
		return nil
	}
}
