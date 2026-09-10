package graphql

// Card is a linked payment card in the wallet.
type Card struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	LinkedAt int64  `json:"linkedAt"`
	Last4    string `json:"last4"`
}

// WalletCard is the result of starting a card-link flow.
type WalletCard struct {
	// CardID      string `json:"cardId"`
	RedirectURL string `json:"redirectUrl"`
}

// PaymentResult is the result of charging a saved card.
type PaymentResult struct {
	PaymentID string `json:"paymentId"`
	Status    string `json:"status"`
}
