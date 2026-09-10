package main

import (
	"context"
	"log"
	natsadp "payment-system/internal/adapters/nats"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const durableName = "PAYMENT_PIDAR"

func main() {
	ctx := context.Background()

	natsClient, err := natsadp.New(natsadp.NatsURL, natsadp.StreamPayments)
	if err != nil {
		log.Fatal(err)
	}
	defer natsClient.Close()

	natsClient.Subscribe(ctx, natsadp.StreamPayments, durableName, func(msg jetstream.Msg) {
		log.Printf("Received message: %s", string(msg.Data()))
		msg.Ack()
	})

	log.Printf("Started worker")

	time.Sleep(time.Hour)
	select {}
}
