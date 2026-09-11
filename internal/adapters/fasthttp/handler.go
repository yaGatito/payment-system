package fasthttpadp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	natsadp "payment-system/internal/adapters/nats"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type CallbackHandler struct {
	natsClient  natsadp.NatsClient
	natsSubject string
}

func NewCallbackHandler(natsClient natsadp.NatsClient, natsSubject string) *CallbackHandler {
	return &CallbackHandler{
		natsClient:  natsClient,
		natsSubject: natsSubject,
	}
}

func (ch *CallbackHandler) NotifyHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	logReq(ctx, body)

	var response rozetkaApiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error unmarshaling JSON: %v\n", err)
		writeResponse(ctx, fasthttp.StatusBadRequest, `{"error":"invalid json"}`)
		return
	}

	log.Printf("Received API response: %+v", response)

	paymentEvent, err := toPaymentEvent(response)
	if err != nil {
		log.Printf("Error mapping to payment event: %v", err)
		writeResponse(ctx, fasthttp.StatusBadRequest, `{"error":"invalid callback payload"}`)
		return
	}
	eventData, err := json.Marshal(paymentEvent)

	log.Printf("Mapped to event: %+v", string(eventData))

	if err != nil {
		log.Printf("Error marshaling payment event: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}

	publishCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = ch.natsClient.Publish(publishCtx, ch.natsSubject, eventData)
	if err != nil {
		log.Printf("Error publishing payment event: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}

	writeResponse(ctx, fasthttp.StatusOK, `{"status":"ok"}`)
}

func writeResponse(ctx *fasthttp.RequestCtx, statusCode int, message string) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(message)
}

func toPaymentEvent(response rozetkaApiResponse) (natsadp.PaymentEvent, error) {
	if response.ExternalID == "" {
		return natsadp.PaymentEvent{}, fmt.Errorf("missing external_id")
	}
	if response.Details.TransactionID == "" {
		return natsadp.PaymentEvent{}, fmt.Errorf("missing transaction_id")
	}
	if response.PaymentMethod.CcToken.Token == "" || response.PaymentMethod.CcToken.Mask == "" {
		return natsadp.PaymentEvent{}, fmt.Errorf("missing card token or mask")
	}

	compositeID := strings.Split(response.ExternalID, "_")
	if len(compositeID) != 3 || compositeID[0] == "" || compositeID[1] == "" || compositeID[2] == "" {
		return natsadp.PaymentEvent{}, fmt.Errorf("invalid external_id format")
	}
	eventType := compositeID[0]
	customerID := compositeID[1]
	paymentID := compositeID[2]

	return natsadp.PaymentEvent{
		CustomerID:    customerID,
		PaymentID:     paymentID,
		EventType:     eventType,
		TransactionID: response.Details.TransactionID,
		Status:        response.Details.Status,
		CcToken:       response.PaymentMethod.CcToken.Token,
		CardMask:      response.PaymentMethod.CcToken.Mask,
		PaymentSystem: response.PaymentMethod.CcToken.PaymentSystem,
	}, nil
}

func logReq(ctx *fasthttp.RequestCtx, rawBody []byte) {
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
	log.Printf("%s", rawBody)
	log.Printf("============================")
}
