package natsadp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"payment-system/internal/app"
	"payment-system/internal/domain"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

type PaymentMessageHandler struct {
	worker app.PaymentWorkerInteractor
	seen   *sync.Map
}

func NewPaymentMessageHandler(worker app.PaymentWorkerInteractor) *PaymentMessageHandler {
	return &PaymentMessageHandler{
		worker: worker,
		seen:   &sync.Map{},
	}
}

func (h *PaymentMessageHandler) HandleMessage(ctx context.Context, msg jetstream.Msg) {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		log.Printf("failed to deserialize event data: %s", string(msg.Data()))
		msg.Ack()
		return
	}

	key := fmt.Sprintf("%s:%s:%s:%s", event.CustomerID, event.PaymentID, event.TransactionID, event.Status)
	if _, ok := h.seen.LoadOrStore(key, struct{}{}); ok {
		log.Printf("duplicate payment event skipped: %s", key)
		msg.Ack()
		return
	}

	if err := h.handleEvent(ctx, event); err != nil {
		log.Printf("failed to process event %s: %v", key, err)
		msg.Ack()
		return
	}

	msg.Ack()
}

func (h *PaymentMessageHandler) handleEvent(ctx context.Context, event PaymentEvent) error {
	switch event.EventType {
	case domain.AddCardEventType:
		if event.Status != "success" {
			return nil
		}
		customerID, err := uuid.Parse(event.CustomerID)
		if err != nil {
			return fmt.Errorf("invalid customer id: %w", err)
		}
		return h.worker.SaveCard(ctx, domain.Card{
			CustomerID: customerID,
			Token:      event.CcToken,
			Type:       event.PaymentSystem,
			Last4:      event.CardMask,
		})

	case domain.PaymentEventType:
		paymentID, err := uuid.Parse(event.PaymentID)
		if err != nil {
			return fmt.Errorf("invalid payment id: %w", err)
		}
		return h.worker.UpdatePaymentStatus(ctx, paymentID, event.Status)

	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}
