package main

import (
	"log"
	fasthttpadp "payment-system/internal/adapters/fasthttp"
	natsadp "payment-system/internal/adapters/nats"

	"github.com/valyala/fasthttp"
)

func main() {
	natsClient, err := natsadp.New(natsadp.NatsURL, natsadp.StreamPayments)
	if err != nil {
		log.Fatal(err)
	}
	defer natsClient.Close()

	callbackHandler := fasthttpadp.NewCallbackHandler(natsClient)

	server := &fasthttp.Server{
		Handler: callbackHandler.NotifyHandler,
	}

	log.Println("Notify server listening on :8081")

	if err := server.ListenAndServe(":8081"); err != nil {
		log.Fatal(err)
	}
}
