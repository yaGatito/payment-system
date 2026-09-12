package app

import (
	"context"
	"fmt"
	thirdparty "payment-system/internal/adapters/3rdparty"
	"payment-system/internal/domain"
	"payment-system/internal/ports"

	"github.com/google/uuid"
)

type WalletInteractor interface {
	AddCard(ctx context.Context, customerID uuid.UUID) (domain.WalletCard, error)
	GetCards(ctx context.Context, customerID uuid.UUID) ([]domain.Card, error)
	GetPayments(ctx context.Context, customerID uuid.UUID, limit, offset int32) ([]domain.Payment, error)
	ChargeSavedCard(ctx context.Context, payment domain.Payment) (domain.PaymentResult, error)
}

var _ WalletInteractor = (*WalletService)(nil)

type WalletService struct {
	db            ports.WalletRepository
	rozetkaClient thirdparty.RozetkaClient
}

func NewWalletService(db ports.WalletRepository, rozetkaClient thirdparty.RozetkaClient) WalletInteractor {
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

func (ws *WalletService) GetPayments(ctx context.Context, customerID uuid.UUID, limit, offset int32) ([]domain.Payment, error) {
	if customerID == uuid.Nil {
		return nil, fmt.Errorf("invalid customer ID")
	}
	if limit <= 0 {
		limit = 10
	}
	return ws.db.GetPayments(ctx, customerID, limit, offset)
}

func (ws *WalletService) ChargeSavedCard(ctx context.Context, payment domain.Payment) (domain.PaymentResult, error) {
	if payment.CustomerID == uuid.Nil {
		return domain.PaymentResult{}, fmt.Errorf("invalid customer ID")
	}
	if payment.OrderID == uuid.Nil {
		return domain.PaymentResult{}, fmt.Errorf("invalid order ID")
	}
	if payment.CardID == uuid.Nil {
		return domain.PaymentResult{}, fmt.Errorf("card id is required")
	}

	card, err := ws.db.GetCardByID(ctx, payment.CardID)
	if err != nil {
		return domain.PaymentResult{}, err
	}
	if card.CustomerID != payment.CustomerID {
		return domain.PaymentResult{}, fmt.Errorf("card does not belong to customer")
	}

	resp, err := ws.rozetkaClient.CreatePaymentWithSavedCard(ctx, thirdparty.CreatePaymentWithTokenRequest{
		OrderID:       payment.OrderID.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		CustomerID:    payment.CustomerID.String(),
		CustomerToken: card.Token,
	})
	if err != nil {
		// _ = ws.db.UpdatePaymentStatus(ctx, payment.OrderID, domain.FailurePaymentStatus)
		return domain.PaymentResult{}, err
	}
	// if resp.Status == domain.FailurePaymentStatus {
	// 	// TODO: log on failure payments
	// }

	pendingPayment := domain.Payment{
		CustomerID: payment.CustomerID,
		CardID:     payment.CardID,
		OrderID:    payment.OrderID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		Status:     resp.Status,
	}
	if err := ws.db.AddPayment(ctx, pendingPayment); err != nil {
		return domain.PaymentResult{}, err
	}

	return domain.PaymentResult{
		OrderID:     payment.OrderID,
		Status:      resp.Status,
		RedirectURL: resp.RedirectURL,
	}, nil
}
