package graphql

import (
	"context"
	"fmt"
	"payment-system/internal/app"
	"payment-system/internal/domain"

	"github.com/google/uuid"
	gql "github.com/graphql-go/graphql"
)

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
		// "cardId":      &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"redirectUrl": &gql.Field{Type: gql.NewNonNull(gql.String)},
	},
})

var paymentResultType = gql.NewObject(gql.ObjectConfig{
	Name: "PaymentResult",
	Fields: gql.Fields{
		"paymentId": &gql.Field{Type: gql.NewNonNull(gql.ID)},
		"status":    &gql.Field{Type: gql.NewNonNull(gql.String)},
	},
})

// NewSchema builds the wallet GraphQL schema backed by store.
func NewSchema(walletInteractor app.WalletInteractor) (gql.Schema, error) {
	if walletInteractor == nil {
		return gql.Schema{}, fmt.Errorf("walletInteractor is required")
	}

	query := gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"getCards": &gql.Field{
				Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(cardType))),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					customerID, _ := p.Args["customerId"].(string)
					return walletInteractor.GetCards(context.Background(), uuid.MustParse(customerID))
				},
			},
		},
	})

	mutation := gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"addCardToWallet": &gql.Field{
				Type: gql.NewNonNull(walletCardType),
				Args: gql.FieldConfigArgument{
					"customerId": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					customerID, _ := p.Args["customerId"].(string)
					wallet, err := walletInteractor.AddCard(context.Background(), uuid.MustParse(customerID))
					if err != nil {
						return nil, err
					}
					return toGraphQLWalletCard(wallet), nil
				},
			},
			"payWithCard": &gql.Field{
				Type: gql.NewNonNull(paymentResultType),
				Args: gql.FieldConfigArgument{
					"cardId":   &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
					"amount":   &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
					"currency": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
					"orderId":  &gql.ArgumentConfig{Type: gql.NewNonNull(gql.ID)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					cardID, _ := p.Args["cardId"].(string)
					amount, _ := p.Args["amount"].(float64)
					currency, _ := p.Args["currency"].(string)
					orderID, _ := p.Args["orderId"].(string)
					return nil, walletInteractor.ChargeSavedCard(context.Background(), domain.Payment{
						CardID:   uuid.MustParse(cardID),
						OrderID:  uuid.MustParse(orderID),
						Amount:   amount,
						Currency: currency,
					})
				},
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
	return &WalletCard{
		RedirectURL: card.RedirectURL,
	}
}
