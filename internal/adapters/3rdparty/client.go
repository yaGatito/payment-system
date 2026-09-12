package rozetkaclient

import "context"

type RozetkaClient interface {
	AddPaymentMethod(
		ctx context.Context,
		req AddPaymentMethodRequest,
	) (AddPaymentMethodResponse, error)
	CreatePaymentWithSavedCard(
		ctx context.Context,
		req CreatePaymentWithTokenRequest,
	) (CreatePaymentWithTokenResponse, error)
}
