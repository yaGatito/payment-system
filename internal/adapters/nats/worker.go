package natsadp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"payment-system/internal/app"
	"payment-system/internal/domain"
	"payment-system/pkg/logger"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

const handleMessageTimeout = 10 * time.Second

type PaymentMessageHandler struct {
	worker app.PaymentWorkerInteractor
	seen   *sync.Map
	logger *logger.Logger
}

func NewPaymentMessageHandler(worker app.PaymentWorkerInteractor, l *logger.Logger) (*PaymentMessageHandler, error) {
	if worker == nil {
		return nil, fmt.Errorf("payment worker is required")
	}
	if l == nil {
		l = logger.New()
	}
	return &PaymentMessageHandler{
		worker: worker,
		seen:   &sync.Map{},
		logger: l,
	}, nil
}

func (h *PaymentMessageHandler) HandleMessage(msg jetstream.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), handleMessageTimeout)
	defer cancel()

	var event PaymentEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		h.logger.Error("failed to deserialize event data: %s", string(msg.Data()))
		msg.Ack()
		return
	}

	dedupKey := fmt.Sprintf("%s:%s:%s", event.CustomerID, event.TransactionID,  event.Status)
	h.logger.Info("event received: type=%s customer=%s payment=%s tx=%s status=%s", event.EventType, event.CustomerID, event.PaymentID, event.TransactionID, event.Status)
	if _, ok := h.seen.LoadOrStore(dedupKey, struct{}{}); ok {
		h.logger.Warn("duplicate payment event skipped: %s", dedupKey)
		msg.Ack()
		return
	}

	if err := h.handleEvent(ctx, event); err != nil {
		h.logger.Error("failed to process event %s: %v", dedupKey, err)
		msg.Ack()
		return
	}

	h.logger.Info("event processed successfully: type=%s customer=%s payment=%s status=%s", event.EventType, event.CustomerID, event.PaymentID, event.Status)
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
		err = h.worker.SaveCard(ctx, domain.Card{
			CustomerID: customerID,
			Token:      event.CcToken,
			Type:       event.PaymentSystem,
			Last4:      event.CardMask,
		})
		if err == nil {
			h.logger.Info("card saved to DB: customer=%s last4=%s", customerID, event.CardMask)
		}
		return err

	case domain.PaymentEventType:
		paymentID, err := uuid.Parse(event.PaymentID)
		if err != nil {
			return fmt.Errorf("invalid payment id: %w", err)
		}
		err = h.worker.UpdatePaymentStatus(ctx, paymentID, event.Status)
		if err == nil {
			h.logger.Info("payment status updated in DB: payment=%s status=%s", paymentID, event.Status)
		}
		return err

	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}
