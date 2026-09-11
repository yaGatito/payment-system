package fasthttpadp

type rozetkaApiResponse struct {
	ID                string `json:"id"`
	ExternalID        string `json:"external_id"`
	UnifiedExternalID string `json:"unified_external_id"`
	IsSuccess         bool   `json:"is_success"`
	Details           struct {
		TransactionID string `json:"transaction_id"`
		Amount        string `json:"amount"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
		StatusCode    string `json:"status_code"`
		StatusDesc    string `json:"status_description"`
		StatusDescEN  string `json:"status_description_en"`
		StatusDescUK  string `json:"status_description_uk"`
		CreatedAt     string `json:"created_at"`
		RecipientIban string `json:"recipient_iban"`
	} `json:"details"`
	ActionRequired bool `json:"action_required"`
	Action         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"action"`
	PaymentMethod struct {
		Type    string `json:"type"`
		CcToken struct {
			Token         string  `json:"token"`
			Mask          string  `json:"mask"`
			ExpiresAt     string  `json:"expires_at"`
			BankShortName *string `json:"bank_short_name"`
			PaymentSystem string  `json:"payment_system"`
			SavedCard     bool    `json:"saved_card"`
			BinCountry    string  `json:"bin_country"`
		} `json:"cc_token"`
	} `json:"payment_method"`
	Customer struct {
		ExternalID string `json:"external_id"`
	} `json:"customer"`
}
