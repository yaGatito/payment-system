package natsadp

type PaymentEvent struct {
	TransactionID string `json:"transaction_id"`
	CustomerID    string `json:"customer_id"`
	PaymentID     string `json:"payment_id"`
	EventType     string `json:"event_type"`
	Status        string `json:"status"`
	CcToken       string `json:"cc_token"`
	CardMask      string `json:"mask"`
	PaymentSystem string `json:"payment_system"`
}
