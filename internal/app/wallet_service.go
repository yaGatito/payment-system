package app

import (
	"context"
	"fmt"
	thirdparty "payment-system/internal/adapters/3rdparty"
	"payment-system/internal/adapters/postgres"
	"payment-system/internal/domain"

	"github.com/google/uuid"
)

type WalletInteractor interface {
	AddCard(ctx context.Context, customerID uuid.UUID) (domain.WalletCard, error)
	GetCards(ctx context.Context, customerID uuid.UUID) ([]domain.Card, error)
	ChargeSavedCard(ctx context.Context, payment domain.Payment) error
}

var _ WalletInteractor = (*WalletService)(nil)

type WalletService struct {
	db            postgres.WalletRepository
	rozetkaClient thirdparty.RozetkaClient
}

func NewWalletService(db postgres.WalletRepository, rozetkaClient thirdparty.RozetkaClient) WalletInteractor {
	return &WalletService{
		db:            db,
		rozetkaClient: rozetkaClient,
	}
}

func (ws *WalletService) AddCard(ctx context.Context, customerID uuid.UUID) (domain.WalletCard, error) {
	if customerID == uuid.Nil {
		return domain.WalletCard{}, fmt.Errorf("invalid customer ID")
	}

	resp, err := ws.rozetkaClient.AddPaymentMethod(ctx, thirdparty.AddPaymentMethodRequest{
		CustomerID: customerID.String(),
	})
	if err != nil {
		return domain.WalletCard{}, err
	}

	return domain.WalletCard{
		CustomerID:  customerID,
		RedirectURL: resp.RedirectURL,
	}, nil
}

func (ws *WalletService) GetCards(ctx context.Context, customerID uuid.UUID) ([]domain.Card, error) {
	if customerID == uuid.Nil {
		return []domain.Card{}, fmt.Errorf("invalid customer ID")
	}
	return ws.db.GetCards(ctx, customerID), nil
}

func (ws *WalletService) ChargeSavedCard(ctx context.Context, payment domain.Payment) error {
	card, err := ws.db.GetCardByID(ctx, payment.CardID)
	if err != nil {
		return err
	}

	resp, err := ws.rozetkaClient.CreatePaymentWithSavedCard(ctx, thirdparty.CreatePaymentWithTokenRequest{
		OrderID:       payment.OrderID.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		CustomerID:    payment.CustomerID.String(),
		CustomerToken: card.Token,
	})

	if resp.Status == "failed" {
		return fmt.Errorf("payment failed")
	}

	return nil
}
