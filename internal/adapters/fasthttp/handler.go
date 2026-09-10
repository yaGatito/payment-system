package fasthttpadp

import (
	"encoding/json"
	"log"
	natsadp "payment-system/internal/adapters/nats"
	"strings"

	"github.com/valyala/fasthttp"
)

type paymentEvent struct {
	TransactionID string `json:"transaction_id"`
	CustomerID    string `json:"customer_id"`
	PaymentID     string `json:"payment_id"`
	EventType     string `json:"event_type"`
	Status        string `json:"status"`
	CcToken       string `json:"cc_token"`
	CardMask      string `json:"mask"`
	PaymentSystem string `json:"payment_system"`
}

type apiResponse struct {
	ID                string `json:"id"`
	ExternalID        string `json:"external_id"`
	UnifiedExternalID string `json:"unified_external_id"`
	IsSuccess         bool   `json:"is_success"`
	Details           struct {
		TransactionID string `json:"transaction_id"`
		Amount        string `json:"amount"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
		StatusCode    string `json:"status_code"`
		StatusDesc    string `json:"status_description"`
		StatusDescEN  string `json:"status_description_en"`
		StatusDescUK  string `json:"status_description_uk"`
		CreatedAt     string `json:"created_at"`
		RecipientIban string `json:"recipient_iban"`
	} `json:"details"`
	ActionRequired bool `json:"action_required"`
	Action         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"action"`
	PaymentMethod struct {
		Type    string `json:"type"`
		CcToken struct {
			Token         string  `json:"token"`
			Mask          string  `json:"mask"`
			ExpiresAt     string  `json:"expires_at"`
			BankShortName *string `json:"bank_short_name"`
			PaymentSystem string  `json:"payment_system"`
			SavedCard     bool    `json:"saved_card"`
			BinCountry    string  `json:"bin_country"`
		} `json:"cc_token"`
	} `json:"payment_method"`
	Customer struct {
		ExternalID string `json:"external_id"`
	} `json:"customer"`
}

type CallbackHandler struct {
	natsClient natsadp.NatsClient
}

func NewCallbackHandler(natsClient natsadp.NatsClient) *CallbackHandler {
	return &CallbackHandler{
		natsClient: natsClient,
	}
}

func (ch *CallbackHandler) NotifyHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	logReq(ctx, body)

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error unmarshaling JSON: %v\n", err)
		writeResponse(ctx, fasthttp.StatusBadRequest, `{"error":"invalid json"}`)
		return
	}

	log.Printf("Received API response: %+v", response)

	paymentEvent, err := toPaymentEvent(response)
	if err != nil {
		log.Printf("Error mapping to payment event: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}
	eventData, err := json.Marshal(paymentEvent)

	log.Printf("Mapped to event: %+v", string(eventData))

	if err != nil {
		log.Printf("Error marshaling payment event: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}

	err = ch.natsClient.Publish(ctx, natsadp.SubjectPayments, eventData)
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

func toPaymentEvent(response apiResponse) (paymentEvent, error) {
	compositeID := strings.Split(response.ExternalID, "_")
	if len(compositeID) != 3 {
		log.Printf("Received not 3 fragments of composite external ID")
		return paymentEvent{}, nil
	}
	eventType := compositeID[0]
	customerID := compositeID[1]
	paymentID := compositeID[2]

	return paymentEvent{
		TransactionID: response.Details.TransactionID,
		PaymentID:     paymentID,
		CustomerID:    customerID,
		EventType:     eventType,
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
