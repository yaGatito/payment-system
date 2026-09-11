package main

import (
	"log"
	fasthttpadp "payment-system/internal/adapters/fasthttp"
	natsadp "payment-system/internal/adapters/nats"

	"github.com/caarlos0/env/v11"
	"github.com/valyala/fasthttp"
)

type EnvConfig struct {
	ListenAddr  string `env:"CALLBACK_LISTEN_ADDR,notEmpty"`
	NatsURL     string `env:"NATS_URL,notEmpty"`
	NatsStream  string `env:"NATS_STREAM,notEmpty"`
	NatsSubject string `env:"NATS_SUBJECT,notEmpty"`
}

func main() {
	cfg := EnvConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("parsing callback config failed: %v", err)
	}

	natsClient, err := natsadp.New(cfg.NatsURL, cfg.NatsStream)
	if err != nil {
		log.Fatal(err)
	}
	defer natsClient.Close()

	callbackHandler := fasthttpadp.NewCallbackHandler(natsClient, cfg.NatsSubject)

	server := &fasthttp.Server{
		Handler: callbackHandler.NotifyHandler,
	}

	log.Printf("Notify server listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
