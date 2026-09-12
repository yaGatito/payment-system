package ports

import (
	"context"
	"payment-system/internal/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	AddCard(ctx context.Context, card domain.Card) error
	GetCardByID(ctx context.Context, cardID uuid.UUID) (domain.Card, error)
	GetCards(ctx context.Context, ownerID uuid.UUID) []domain.Card
	RemoveCard(ctx context.Context, cardID uuid.UUID) error

	AddPayment(ctx context.Context, payment domain.Payment) error
	UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, paymentStatus string) error
	GetPayments(
		ctx context.Context,
		ownerID uuid.UUID,
		limit, offset int32,
	) ([]domain.Payment, error)
}
