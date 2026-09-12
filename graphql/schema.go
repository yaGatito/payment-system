package graphql

import (
	"context"
	"fmt"
	"payment-system/internal/app"
	"payment-system/internal/domain"

	"github.com/google/uuid"
	gql "github.com/graphql-go/graphql"
)

const (
	fieldCustomerID = "customerId"
	fieldCardID     = "cardId"
	fieldOrderID    = "orderId"
	fieldAmount     = "amount"
	fieldCurrency   = "currency"
	fieldLimit      = "limit"
	fieldOffset     = "offset"
)

type SchemaHandler struct {
	walletInteractor app.WalletInteractor
}

func NewSchemaHandler(walletInteractor app.WalletInteractor) *SchemaHandler {
	return &SchemaHandler{walletInteractor: walletInteractor}
}

func (h *SchemaHandler) GetCards(ctx context.Context, customerIDStr string) ([]*Card, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customerId: %w", err)
	}
	cards, err := h.walletInteractor.GetCards(ctx, customerID)
	if err != nil {
		return nil, err
	}
	graphqlCards := make([]*Card, len(cards))
	for i, card := range cards {
		graphqlCards[i] = toGraphQLCard(card)
	}
	return graphqlCards, nil
}

func (h *SchemaHandler) AddCardToWallet(
	ctx context.Context,
	customerIDStr string,
) (*WalletCard, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customerId: %w", err)
	}
	wallet, err := h.walletInteractor.AddCard(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return toGraphQLWalletCard(wallet), nil
}

func (h *SchemaHandler) PayWithCard(
	ctx context.Context,
	customerIDStr, cardIDStr, orderIDStr, currency string,
	amount int64,
) (*PaymentResult, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customerId: %w", err)
	}
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cardId: %w", err)
	}
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid orderId: %w", err)
	}
	payment := domain.Payment{
		CustomerID: customerID,
		CardID:     cardID,
		OrderID:    orderID,
		Amount:     amount,
		Currency:   currency,
	}
	payResult, err := h.walletInteractor.ChargeSavedCard(ctx, payment)
	if err != nil {
		return nil, err
	}
	return &PaymentResult{
		PaymentID:   orderID.String(),
		Status:      payResult.Status,
		RedirectURL: payResult.RedirectURL,
	}, nil
}

func (h *SchemaHandler) GetPayments(
	ctx context.Context,
	customerIDStr string,
	limit, offset int32,
) ([]*Payment, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid customerId: %w", err)
	}
	if limit <= 0 {
		limit = 10
	}
	items, err := h.walletInteractor.GetPayments(ctx, customerID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]*Payment, len(items))
	for i, item := range items {
		result[i] = &Payment{
			PaymentID: item.OrderID.String(),
			Amount:    item.Amount,
			Currency:  item.Currency,
			Status:    item.Status,
			CreatedAt: item.CreatedAt.Unix(),
		}
	}
	return result, nil
}

func (h *SchemaHandler) getCardsResolver(p gql.ResolveParams) (interface{}, error) {
	customerIDStr, ok := p.Args[fieldCustomerID].(string)
	if !ok || customerIDStr == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	return h.GetCards(context.Background(), customerIDStr)
}

func (h *SchemaHandler) addCardToWalletResolver(p gql.ResolveParams) (interface{}, error) {
	customerIDStr, ok := p.Args[fieldCustomerID].(string)
	if !ok || customerIDStr == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	return h.AddCardToWallet(context.Background(), customerIDStr)
}

func (h *SchemaHandler) payWithCardResolver(p gql.ResolveParams) (interface{}, error) {
	customerIDStr, ok := p.Args[fieldCustomerID].(string)
	if !ok || customerIDStr == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	cardIDStr, ok := p.Args[fieldCardID].(string)
	if !ok || cardIDStr == "" {
		return nil, fmt.Errorf("cardId is required")
	}
	orderIDStr, ok := p.Args[fieldOrderID].(string)
	if !ok || orderIDStr == "" {
		return nil, fmt.Errorf("orderId is required")
	}
	currency, ok := p.Args[fieldCurrency].(string)
	if !ok || currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	amount, ok := p.Args[fieldAmount].(int64)
	if !ok || amount == 0 {
		return nil, fmt.Errorf("amount is zero")
	}

	return h.PayWithCard(
		context.Background(),
		customerIDStr,
		cardIDStr,
		orderIDStr,
		currency,
		amount,
	)
}

func (h *SchemaHandler) getPaymentsResolver(p gql.ResolveParams) (interface{}, error) {
	customerIDStr, ok := p.Args[fieldCustomerID].(string)
	if !ok || customerIDStr == "" {
		return nil, fmt.Errorf("customerId is required")
	}
	limit := int32(10)
	if v, ok := p.Args[fieldLimit].(int32); ok && v > 0 {
		limit = v
	}
	offset := int32(0)
	if v, ok := p.Args[fieldOffset].(int32); ok && v >= 0 {
		offset = v
	}
	return h.GetPayments(context.Background(), customerIDStr, limit, offset)
}

var cardType = gql.NewObject(gql.ObjectConfig{
	Name: "Card",
	Fields: gql.Fields{
		"id":       &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"type":     &gql.Field{Type: gql.NewNonNull(gql.String)},
		"linkedAt": &gql.Field{Type: gql.NewNonNull(gql.Int)},
		"last4":    &gql.Field{Type: gql.NewNonNull(gql.String)},
	},
})

var walletCardType = gql.NewObject(gql.ObjectConfig{
	Name: "WalletCard",
	Fields: gql.Fields{
		"cardId":      &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"redirectUrl": &gql.Field{Type: gql.NewNonNull(gql.String)},
	},
})

var paymentResultType = gql.NewObject(gql.ObjectConfig{
	Name: "PaymentResult",
	Fields: gql.Fields{
		"paymentId":   &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"status":      &gql.Field{Type: gql.NewNonNull(gql.String)},
		"redirectUrl": &gql.Field{Type: gql.NewNonNull(gql.String)},
	},
})

var paymentType = gql.NewObject(gql.ObjectConfig{
	Name: "Payment",
	Fields: gql.Fields{
		"paymentId": &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"amount":    &gql.Field{Type: gql.NewNonNull(gql.Int)},
		"currency":  &gql.Field{Type: gql.NewNonNull(gql.String)},
		"status":    &gql.Field{Type: gql.NewNonNull(gql.String)},
		"createdAt": &gql.Field{Type: gql.NewNonNull(gql.Int)},
	},
})

func NewSchema(walletInteractor app.WalletInteractor) (gql.Schema, error) {
	if walletInteractor == nil {
		return gql.Schema{}, fmt.Errorf("walletInteractor is required")
	}

	handler := NewSchemaHandler(walletInteractor)

	query := gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"getCards": &gql.Field{
				Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(cardType))),
				Args: gql.FieldConfigArgument{
					fieldCustomerID: &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
				},
				Resolve: handler.getCardsResolver,
			},
			"getPayments": &gql.Field{
				Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(paymentType))),
				Args: gql.FieldConfigArgument{
					fieldCustomerID: &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
					fieldLimit:      &gql.ArgumentConfig{Type: gql.Int},
					fieldOffset:     &gql.ArgumentConfig{Type: gql.Int},
				},
				Resolve: handler.getPaymentsResolver,
			},
		},
	})

	mutation := gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"addCardToWallet": &gql.Field{
				Type: gql.NewNonNull(walletCardType),
				Args: gql.FieldConfigArgument{
					fieldCustomerID: &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
				},
				Resolve: handler.addCardToWalletResolver,
			},
			"payWithCard": &gql.Field{
				Type: gql.NewNonNull(paymentResultType),
				Args: gql.FieldConfigArgument{
					fieldCustomerID: &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
					fieldCardID:     &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
					fieldAmount:     &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
					fieldCurrency:   &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
					fieldOrderID:    &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
				},
				Resolve: handler.payWithCardResolver,
			},
		},
	})

	return gql.NewSchema(gql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
}

func toGraphQLCard(card domain.Card) *Card {
	return &Card{
		ID:       card.CardID.String(),
		Type:     card.Type,
		LinkedAt: card.LinkedAt.Unix(),
		Last4:    card.Last4,
	}
}

func toGraphQLWalletCard(card domain.WalletCard) *WalletCard {
	walletCard := &WalletCard{
		CardID:      card.CardID.String(),
		RedirectURL: card.RedirectURL,
	}
	if walletCard.CardID == (uuid.UUID{}).String() {
		walletCard.CardID = uuid.NewString()
	}
	return walletCard
}
