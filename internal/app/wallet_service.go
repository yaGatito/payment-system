package app

import (
	"context"
	"fmt"
	thirdparty "payment-system/internal/adapters/3rdparty"
	"payment-system/internal/domain"
	"payment-system/internal/ports"
	"payment-system/pkg/logger"

	"github.com/google/uuid"
)

type WalletInteractor interface {
	AddCard(ctx context.Context, customerID uuid.UUID) (domain.WalletCard, error)
	GetCards(ctx context.Context, customerID uuid.UUID) ([]domain.Card, error)
	GetPayments(
		ctx context.Context,
		customerID uuid.UUID,
		limit, offset int32,
	) ([]domain.Payment, error)
	ChargeSavedCard(ctx context.Context, payment domain.Payment) (domain.PaymentResult, error)
}

var _ WalletInteractor = (*WalletService)(nil)

type WalletService struct {
	repo          ports.WalletRepository
	rozetkaClient thirdparty.RozetkaClient
	logger        *logger.Logger
}

func NewWalletService(
	repo ports.WalletRepository,
	rozetkaClient thirdparty.RozetkaClient,
	l *logger.Logger,
) WalletInteractor {
	if l == nil {
		l = logger.New()
	}
	return &WalletService{
		repo:          repo,
		rozetkaClient: rozetkaClient,
		logger:        l,
	}
}

func (ws *WalletService) AddCard(
	ctx context.Context,
	customerID uuid.UUID,
) (domain.WalletCard, error) {
	if customerID == uuid.Nil {
		ws.logger.Error("invalid customer ID for add card")
		return domain.WalletCard{}, fmt.Errorf("invalid customer ID")
	}

	resp, err := ws.rozetkaClient.AddPaymentMethod(ctx, thirdparty.AddPaymentMethodRequest{
		CustomerID: customerID.String(),
	})
	if err != nil {
		ws.logger.Error("add card request failed: customer=%s err=%v", customerID, err)
		return domain.WalletCard{}, err
	}

	ws.logger.Info("new card flow created: customer=%s redirect=%s", customerID, resp.RedirectURL)
	return domain.WalletCard{
		CustomerID:  customerID,
		RedirectURL: resp.RedirectURL,
	}, nil
}

func (ws *WalletService) GetCards(
	ctx context.Context,
	customerID uuid.UUID,
) ([]domain.Card, error) {
	if customerID == uuid.Nil {
		ws.logger.Error("invalid customer ID for get cards")
		return []domain.Card{}, fmt.Errorf("invalid customer ID")
	}
	cards := ws.repo.GetCards(ctx, customerID)
	ws.logger.Info("cards fetched: customer=%s count=%d", customerID, len(cards))
	return cards, nil
}

func (ws *WalletService) GetPayments(
	ctx context.Context,
	customerID uuid.UUID,
	limit, offset int32,
) ([]domain.Payment, error) {
	if customerID == uuid.Nil {
		ws.logger.Error("invalid customer ID for get payments")
		return nil, fmt.Errorf("invalid customer ID")
	}
	if limit <= 0 {
		limit = 10
	}
	payments, err := ws.repo.GetPayments(ctx, customerID, limit, offset)
	if err != nil {
		ws.logger.Error("payments fetch failed: customer=%s err=%v", customerID, err)
		return nil, err
	}
	ws.logger.Info("payments fetched: customer=%s count=%d", customerID, len(payments))
	return payments, nil
}

func (ws *WalletService) ChargeSavedCard(
	ctx context.Context,
	payment domain.Payment,
) (domain.PaymentResult, error) {
	if payment.CustomerID == uuid.Nil {
		ws.logger.Error("invalid customer ID for charge saved card")
		return domain.PaymentResult{}, fmt.Errorf("invalid customer ID")
	}
	if payment.OrderID == uuid.Nil {
		ws.logger.Error("invalid order ID for charge saved card: customer=%s", payment.CustomerID)
		return domain.PaymentResult{}, fmt.Errorf("invalid order ID")
	}
	if payment.CardID == uuid.Nil {
		ws.logger.Error(
			"card id is required for charge saved card: customer=%s order=%s",
			payment.CustomerID,
			payment.OrderID,
		)
		return domain.PaymentResult{}, fmt.Errorf("card id is required")
	}

	card, err := ws.repo.GetCardByID(ctx, payment.CardID)
	if err != nil {
		ws.logger.Error(
			"failed to load card for payment: card=%s customer=%s order=%s err=%v",
			payment.CardID,
			payment.CustomerID,
			payment.OrderID,
			err,
		)
		return domain.PaymentResult{}, err
	}
	if card.CustomerID != payment.CustomerID {
		ws.logger.Error(
			"card ownership mismatch: customer=%s card=%s owner=%s",
			payment.CustomerID,
			payment.CardID,
			card.CustomerID,
		)
		return domain.PaymentResult{}, fmt.Errorf("card does not belong to customer")
	}

	resp, err := ws.rozetkaClient.CreatePaymentWithSavedCard(
		ctx,
		thirdparty.CreatePaymentWithTokenRequest{
			OrderID:       payment.OrderID.String(),
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			CustomerID:    payment.CustomerID.String(),
			CustomerToken: card.Token,
		},
	)
	if err != nil {
		ws.logger.Error(
			"create saved card payment failed: customer=%s order=%s card=%s err=%v",
			payment.CustomerID,
			payment.OrderID,
			payment.CardID,
			err,
		)
		return domain.PaymentResult{}, err
	}
	ws.logger.Info(
		"saved card payment created: customer=%s order=%s status=%s",
		payment.CustomerID,
		payment.OrderID,
		resp.Status,
	)

	pendingPayment := domain.Payment{
		CustomerID: payment.CustomerID,
		CardID:     payment.CardID,
		OrderID:    payment.OrderID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		Status:     resp.Status,
	}
	if err := ws.repo.AddPayment(ctx, pendingPayment); err != nil {
		ws.logger.Error(
			"failed to store pending payment: customer=%s order=%s err=%v",
			payment.CustomerID,
			payment.OrderID,
			err,
		)
		return domain.PaymentResult{}, err
	}

	return domain.PaymentResult{
		OrderID:     payment.OrderID,
		Status:      resp.Status,
		RedirectURL: resp.RedirectURL,
	}, nil
}
