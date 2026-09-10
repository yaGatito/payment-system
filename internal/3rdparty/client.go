package rozetkaclient

import "context"

type AddPaymentMethodRequest struct {
	CustomerID string `json:"customer_id"` // curls payment external id
}

type AddPaymentMethodResponse struct {
	RedirectURL string `json:"redirect_url"` // in callback service it will receive specific order id (customer id)
}

type CreatePaymentWithTokenRequest struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	CustomerID    string  `json:"customer_id"`
	CustomerToken string  `json:"customer_token"`
}

type CreatePaymentWithTokenResponse struct {
	Status      string `json:"status"`
	RedirectURL string `json:"redirect_url"`
}

type RozetkaClient interface {
	AddPaymentMethod(ctx context.Context, req AddPaymentMethodRequest) (AddPaymentMethodResponse, error)
	CreatePaymentWithSavedCard(ctx context.Context, req CreatePaymentWithTokenRequest) (CreatePaymentWithTokenResponse, error)
}
