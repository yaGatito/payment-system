package graphql

import (
	"context"
	"errors"
	"testing"
	"time"

	"payment-system/internal/domain"

	"github.com/google/uuid"
	gql "github.com/graphql-go/graphql"
)

type testWalletInteractor struct {
	cards    []domain.Card
	payments []domain.Payment
}

func (s *testWalletInteractor) AddCard(ctx context.Context, customerID uuid.UUID) (domain.WalletCard, error) {
	cardID := uuid.New()
	card := domain.Card{
		CustomerID: customerID,
		CardID:     cardID,
		Type:       "visa",
		Last4:      "4242",
		LinkedAt:   time.Now(),
	}
	s.cards = append(s.cards, card)
	return domain.WalletCard{
		CustomerID:  customerID,
		CardID:      cardID,
		RedirectURL: "https://example.com/redirect",
	}, nil
}

func (s *testWalletInteractor) GetCards(ctx context.Context, customerID uuid.UUID) ([]domain.Card, error) {
	if customerID == uuid.Nil {
		return nil, errors.New("invalid customer ID")
	}
	if len(s.cards) == 0 {
		seed := domain.Card{
			CustomerID: customerID,
			CardID:     uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			Token:      "token_123",
			Type:       "visa",
			Last4:      "1234",
			LinkedAt:   time.Now(),
		}
		s.cards = []domain.Card{seed}
	}
	return s.cards, nil
}

func (s *testWalletInteractor) ChargeSavedCard(ctx context.Context, payment domain.Payment) (domain.PaymentResult, error) {
	for _, card := range s.cards {
		if card.CardID == payment.CardID {
			if card.CustomerID != payment.CustomerID {
				return domain.PaymentResult{}, errors.New("card does not belong to customer")
			}
			if payment.CardID == uuid.Nil {
				return domain.PaymentResult{}, errors.New("missing card")
			}
			return domain.PaymentResult{
				OrderID: payment.OrderID,
			}, nil
		}
	}
	return domain.PaymentResult{}, errors.New("unknown card")
}

func (s *testWalletInteractor) GetPayments(ctx context.Context, customerID uuid.UUID, limit, offset int32) ([]domain.Payment, error) {
	if customerID == uuid.Nil {
		return nil, errors.New("invalid customer ID")
	}
	if len(s.payments) == 0 {
		seed := domain.Payment{
			CustomerID: customerID,
			CardID:     uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			OrderID:    uuid.MustParse("22222222-2222-4222-8222-222222222222"),
			Amount:     10000,
			Currency:   "UAH",
			Status:     "success",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		s.payments = []domain.Payment{seed}
	}
	return s.payments, nil
}

func newTestSchema(t *testing.T) gql.Schema {
	t.Helper()
	ownerID := uuid.MustParse("123e4567-e89b-42d3-a456-426614174000")
	cardID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	wallet := &testWalletInteractor{cards: []domain.Card{{
		CustomerID: ownerID,
		CardID:     cardID,
		Token:      "token_123",
		Type:       "visa",
		Last4:      "1234",
		LinkedAt:   time.Now(),
	}}}
	schema, err := NewSchema(wallet)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return schema
}

func TestGetCards(t *testing.T) {
	customerID := uuid.MustParse("123e4567-e89b-42d3-a456-426614174000").String()
	result := gql.Do(gql.Params{
		Schema: newTestSchema(t),
		RequestString: `
			query($customerId: ID!) {
				getCards(customerId: $customerId) {
					id
					type
					linkedAt
					last4
				}
			}`,
		VariableValues: map[string]interface{}{"customerId": customerID},
	})
	if len(result.Errors) > 0 {
		t.Fatalf("graphql errors: %v", result.Errors)
	}
	cards := result.Data.(map[string]interface{})["getCards"].([]interface{})
	if len(cards) == 0 {
		t.Fatal("expected seeded card")
	}
}

func TestPayWithCard(t *testing.T) {
	customerID := uuid.MustParse("123e4567-e89b-42d3-a456-426614174000").String()
	cardID := uuid.MustParse("11111111-1111-4111-8111-111111111111").String()
	result := gql.Do(gql.Params{
		Schema: newTestSchema(t),
		RequestString: `
			mutation($customerId: ID!, $cardId: ID!, $amount: Int!, $currency: String!, $orderId: ID!) {
				payWithCard(customerId: $customerId, cardId: $cardId, amount: $amount, currency: $currency, orderId: $orderId) {
					paymentId
					status
				}
			}`,
		VariableValues: map[string]interface{}{
			"customerId": customerID,
			"cardId":     cardID,
			"amount":     10000,
			"currency":   "UAH",
			"orderId":    "22222222-2222-4222-8222-222222222222",
		},
	})
	if len(result.Errors) > 0 {
		t.Fatalf("graphql errors: %v", result.Errors)
	}
	pay := result.Data.(map[string]interface{})["payWithCard"].(map[string]interface{})
	if pay["status"] != "pending" {
		t.Fatalf("expected pending, got %#v", pay)
	}
	if pay["paymentId"] == "" {
		t.Fatalf("expected paymentId, got %#v", pay)
	}
}

func TestGetPayments(t *testing.T) {
	customerID := uuid.MustParse("123e4567-e89b-42d3-a456-426614174000").String()
	result := gql.Do(gql.Params{
		Schema: newTestSchema(t),
		RequestString: `
			query($customerId: ID!) {
				getPayments(customerId: $customerId, limit: 10, offset: 0) {
					paymentId
					amount
					currency
					status
				}
			}`,
		VariableValues: map[string]interface{}{"customerId": customerID},
	})
	if len(result.Errors) > 0 {
		t.Fatalf("graphql errors: %v", result.Errors)
	}
	payments := result.Data.(map[string]interface{})["getPayments"].([]interface{})
	if len(payments) == 0 {
		t.Fatal("expected seeded payment")
	}
}

func TestPayWithCardUnknownCard(t *testing.T) {
	result := gql.Do(gql.Params{
		Schema: newTestSchema(t),
		RequestString: `
			mutation($customerId: ID!, $cardId: ID!, $amount: Int!, $currency: String!, $orderId: ID!) {
				payWithCard(customerId: $customerId, cardId: $cardId, amount: $amount, currency: $currency, orderId: $orderId) {
					paymentId
					status
				}
			}`,
		VariableValues: map[string]interface{}{
			"customerId": "22222222-2222-4222-8222-222222222222",
			"cardId":     "00000000-0000-0000-0000-000000000000",
			"amount":     100,
			"currency":   "UAH",
			"orderId":    "33333333-3333-4333-8333-333333333333",
		},
	})
	if len(result.Errors) == 0 {
		t.Fatal("expected error for unknown card")
	}
}

func TestPayWithCardCustomerMismatch(t *testing.T) {
	cardID := uuid.MustParse("11111111-1111-4111-8111-111111111111").String()
	result := gql.Do(gql.Params{
		Schema: newTestSchema(t),
		RequestString: `
			mutation($customerId: ID!, $cardId: ID!, $amount: Int!, $currency: String!, $orderId: ID!) {
				payWithCard(customerId: $customerId, cardId: $cardId, amount: $amount, currency: $currency, orderId: $orderId) {
					paymentId
					status
				}
			}`,
		VariableValues: map[string]interface{}{
			"customerId": "22222222-2222-4222-8222-222222222222",
			"cardId":     cardID,
			"amount":     100,
			"currency":   "UAH",
			"orderId":    "44444444-4444-4444-8444-444444444444",
		},
	})
	if len(result.Errors) == 0 {
		t.Fatal("expected error when card belongs to another customer")
	}
}
