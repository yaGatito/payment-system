package app

import (
	"context"
	"payment-system/internal/domain"
	"payment-system/internal/ports"
	"payment-system/pkg/logger"

	"github.com/google/uuid"
)

type PaymentWorkerInteractor interface {
	SaveCard(ctx context.Context, card domain.Card) error
	SavePayment(ctx context.Context, payment domain.Payment) error
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error
}

type PaymentWorkerService struct {
	repo   ports.WalletRepository
	logger *logger.Logger
}

func NewPaymentWorkerService(repo ports.WalletRepository, l *logger.Logger) *PaymentWorkerService {
	if l == nil {
		l = logger.New()
	}
	return &PaymentWorkerService{
		repo:   repo,
		logger: l,
	}
}

func (s *PaymentWorkerService) SaveCard(ctx context.Context, card domain.Card) error {
	if err := s.repo.AddCard(ctx, card); err != nil {
		s.logger.Error(
			"save card failed: customer=%s last4=%s err=%v",
			card.CustomerID,
			card.Last4,
			err,
		)
		return err
	}
	s.logger.Info("card saved: customer=%s last4=%s", card.CustomerID, card.Last4)
	return nil
}

func (s *PaymentWorkerService) SavePayment(ctx context.Context, payment domain.Payment) error {
	if err := s.repo.AddPayment(ctx, payment); err != nil {
		s.logger.Error(
			"save payment failed: customer=%s order=%s err=%v",
			payment.CustomerID,
			payment.OrderID,
			err,
		)
		return err
	}
	s.logger.Info(
		"payment saved: customer=%s order=%s status=%s",
		payment.CustomerID,
		payment.OrderID,
		payment.Status,
	)
	return nil
}

func (s *PaymentWorkerService) UpdatePaymentStatus(
	ctx context.Context,
	orderID uuid.UUID,
	paymentStatus string,
) error {
	if err := s.repo.UpdatePaymentStatus(ctx, orderID, paymentStatus); err != nil {
		s.logger.Error(
			"update payment status failed: order=%s status=%s err=%v",
			orderID,
			paymentStatus,
			err,
		)
		return err
	}
	s.logger.Info("payment status updated: order=%s status=%s", orderID, paymentStatus)
	return nil
}
