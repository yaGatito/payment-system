package app

import (
	"context"
	"payment-system/internal/domain"
	"payment-system/internal/ports"

	"github.com/google/uuid"
)

type PaymentWorkerInteractor interface {
	SaveCard(ctx context.Context, card domain.Card) error
	SavePayment(ctx context.Context, payment domain.Payment) error
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error
}

type PaymentWorkerService struct {
	db ports.WalletRepository
}

func NewPaymentWorkerService(db ports.WalletRepository) *PaymentWorkerService {
	return &PaymentWorkerService{
		db: db,
	}
}

func (s *PaymentWorkerService) SaveCard(ctx context.Context, card domain.Card) error {
	return s.db.AddCard(ctx, card)
}

func (s *PaymentWorkerService) SavePayment(ctx context.Context, payment domain.Payment) error {
	return s.db.AddPayment(ctx, payment)
}

func (s *PaymentWorkerService) UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error {
	return s.db.UpdatePaymentStatus(ctx, orderID, paymentStatus)
}
