package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

func notifyHandler(ctx *fasthttp.RequestCtx) {
	log.Printf("========== NOTIFY ==========")
	log.Printf("Method: %s", ctx.Method())
	log.Printf("URI: %s", ctx.URI().String())
	log.Printf("RemoteAddr: %s", ctx.RemoteAddr().String())

	log.Printf("StatusCode: %d", ctx.Response.StatusCode())

	log.Printf("Headers:")
	hrds := ctx.Request.Header.All()

	for k, v := range hrds {
		log.Printf("%s: %s", string(k), string(v))
	}

	log.Printf("Body:")
	log.Printf("%s", ctx.PostBody())

	log.Printf("============================")

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(`{"status":"ok"}`)
}

func main() {
	server := &fasthttp.Server{
		Handler: notifyHandler,
	}

	log.Println("Notify server listening on :8081")

	if err := server.ListenAndServe(":8081"); err != nil {
		log.Fatal(err)
	}
}
