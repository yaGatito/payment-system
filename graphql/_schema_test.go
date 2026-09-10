package graphql

import (
	"testing"

	gql "github.com/graphql-go/graphql"
)

func exec(t *testing.T, query string) *gql.Result {
	t.Helper()
	schema, err := NewSchema(NewStore())
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	result := gql.Do(gql.Params{Schema: schema, RequestString: query})
	if len(result.Errors) > 0 {
		t.Fatalf("graphql errors: %v", result.Errors)
	}
	return result
}

func TestAddCardToWallet(t *testing.T) {
	result := exec(t, `
		mutation {
			addCardToWallet {
				cardId
				redirectUrl
			}
		}
	`)
	data := result.Data.(map[string]interface{})["addCardToWallet"].(map[string]interface{})
	if data["cardId"] == "" || data["redirectUrl"] == "" {
		t.Fatalf("expected cardId and redirectUrl, got %#v", data)
	}
}

func TestGetCards(t *testing.T) {
	result := exec(t, `
		query {
			getCards {
				id
				type
				linkedAt
				last4
			}
		}
	`)
	cards := result.Data.(map[string]interface{})["getCards"].([]interface{})
	if len(cards) == 0 {
		t.Fatal("expected seeded card")
	}
	card := cards[0].(map[string]interface{})
	if card["id"] != "card_123" {
		t.Fatalf("expected seeded card_123, got %#v", card)
	}
}

func TestPayWithCard(t *testing.T) {
	result := exec(t, `
		mutation {
			payWithCard(
				cardId: "card_123"
				amount: 10000
				currency: "UAH"
				orderId: "order-123"
			) {
				paymentId
				status
			}
		}
	`)
	pay := result.Data.(map[string]interface{})["payWithCard"].(map[string]interface{})
	if pay["status"] != "completed" {
		t.Fatalf("expected completed, got %#v", pay)
	}
	if pay["paymentId"] == "" {
		t.Fatalf("expected paymentId, got %#v", pay)
	}
}

func TestPayWithCardUnknownCard(t *testing.T) {
	schema, err := NewSchema(NewStore())
	if err != nil {
		t.Fatal(err)
	}
	result := gql.Do(gql.Params{
		Schema: schema,
		RequestString: `
			mutation {
				payWithCard(
					cardId: "missing"
					amount: 100
					currency: "UAH"
					orderId: "order-1"
				) {
					paymentId
					status
				}
			}
		`,
	})
	if len(result.Errors) == 0 {
		t.Fatal("expected error for unknown card")
	}
}
