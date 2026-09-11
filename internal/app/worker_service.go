package app

import (
	"context"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/domain"

	"github.com/google/uuid"
)

type PaymentWorkerInteractor interface {
	SaveCard(ctx context.Context, card domain.Card) error
	SavePayment(ctx context.Context, payment domain.Payment) error
	UpdatePayment(ctx context.Context, orderID uuid.UUID, payment domain.Payment) error
}

type PaymentWorkerService struct {
	db postgres.WalletRepository
}

func NewPaymentWorkerService(db postgres.WalletRepository) *PaymentWorkerService {
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
