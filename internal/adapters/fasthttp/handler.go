package fasthttpadp

import (
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	natsadp "payment-system/internal/adapters/nats"
	"payment-system/pkg/logger"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const timeout = 10 * time.Second
const retryDeltaTimeout = 500 * time.Millisecond

type CallbackHandler struct {
	natsClient     natsadp.NatsClient
	natsSubject    string
	callbackSecret string
	logger         *logger.Logger
}

func NewCallbackHandler(
	natsClient natsadp.NatsClient,
	natsSubject, callbackSecret string,
	l *logger.Logger,
) *CallbackHandler {
	if natsClient == nil {
		panic("nats client is required")
	}
	if l == nil {
		l = logger.New()
	}
	return &CallbackHandler{
		natsClient:     natsClient,
		natsSubject:    natsSubject,
		callbackSecret: callbackSecret,
		logger:         l,
	}
}

func (ch *CallbackHandler) NotifyHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	signatureHeader := string(ctx.Request.Header.Peek("X-ROZETKAPAY-SIGNATURE"))

	ch.logger.Info("callback request received: method=%s uri=%s", ctx.Method(), ctx.URI().String())

	if !verifyRequestSignature(body, ch.callbackSecret, signatureHeader) {
		ch.logger.Error("invalid callback signature")
		writeResponse(ctx, fasthttp.StatusUnauthorized, `{"error":"invalid callback signature"}`)
		return
	}

	var response rozetkaApiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		ch.logger.Error("invalid callback json: %v", err)
		writeResponse(ctx, fasthttp.StatusBadRequest, `{"error":"invalid json"}`)
		return
	}

	paymentEvent, err := toPaymentEvent(response)
	if err != nil {
		ch.logger.Error("invalid callback payload: %v", err)
		writeResponse(ctx, fasthttp.StatusBadRequest, `{"error":"invalid callback payload"}`)
		return
	}

	eventData, err := json.Marshal(paymentEvent)
	if err != nil {
		ch.logger.Error("marshal payment event failed: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}

	publishCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := ch.publishWithRetry(publishCtx, ch.natsSubject, eventData); err != nil {
		ch.logger.Error("publish payment event failed: %v", err)
		writeResponse(ctx, fasthttp.StatusInternalServerError, `{"error":"internal server error"}`)
		return
	}

	ch.logger.Info(
		"callback event published: subject=%s type=%s customer=%s payment=%s status=%s",
		ch.natsSubject,
		paymentEvent.EventType,
		paymentEvent.CustomerID,
		paymentEvent.PaymentID,
		paymentEvent.Status,
	)
	writeResponse(ctx, fasthttp.StatusOK, `{"status":"ok"}`)
}

func (ch *CallbackHandler) publishWithRetry(
	ctx context.Context,
	subject string,
	payload []byte,
) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = ch.natsClient.Publish(ctx, subject, payload)
		if err == nil {
			return nil
		}
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return err
			case <-time.After(time.Duration(attempt+1) * retryDeltaTimeout):
			}
		}
	}
	return err
}

func verifyRequestSignature(body []byte, password, signature string) bool {
	if len(body) == 0 || strings.TrimSpace(password) == "" || strings.TrimSpace(signature) == "" {
		return false
	}

	encodedBody := base64.URLEncoding.EncodeToString(body)
	sum := sha1.Sum([]byte(password + encodedBody + password))
	expected := base64.URLEncoding.EncodeToString(sum[:])

	trimmedSignature := strings.TrimSpace(signature)
	trimmedExpected := strings.TrimSpace(expected)

	return subtle.ConstantTimeCompare([]byte(trimmedSignature), []byte(trimmedExpected)) == 1 ||
		subtle.ConstantTimeCompare(
			[]byte(trimmedSignature),
			[]byte(strings.TrimSuffix(trimmedExpected, "=")),
		) == 1
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
	if len(compositeID) != 3 || compositeID[0] == "" ||
		compositeID[1] == "" || compositeID[2] == "" {

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
