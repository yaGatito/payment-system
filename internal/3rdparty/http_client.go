package rozetkaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	BaseURL     string
	Username    string
	Password    string
	CallbackURL string
	HTTPClient  *http.Client
}

func NewClient(baseURL, username, password, callbackURL string) *Client {
	return &Client{
		BaseURL:     baseURL,
		Username:    username,
		Password:    password,
		CallbackURL: callbackURL,
		HTTPClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// internal representation of the Rozetka API response (only used fields)
type apiResponse struct {
	ID             string `json:"id"`
	ExternalID     string `json:"external_id"`
	ActionRequired bool   `json:"action_required"`
	Action         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"action"`
	Details struct {
		Status     string `json:"status"`
		StatusCode string `json:"status_code"`
	} `json:"details"`
	PaymentMethod struct {
		Type    string `json:"type"`
		CcToken struct {
			Token     string `json:"token"`
			Mask      string `json:"mask"`
			ExpiresAt string `json:"expires_at"`
		} `json:"cc_token"`
	} `json:"payment_method"`
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (*apiResponse, error) {
	url := c.BaseURL + path
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	} else {
		return nil, fmt.Errorf("username and password are required for basic auth")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(data))
	}

	var ar apiResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &ar, nil
}

// typed request payloads
type paymentCustomer struct {
	ExternalID    string             `json:"external_id,omitempty"`
	PaymentMethod *paymentMethodBody `json:"payment_method,omitempty"`
}

type paymentMethodBody struct {
	Type    string           `json:"type"`
	CcToken *ccTokenFragment `json:"cc_token,omitempty"`
}

type ccTokenFragment struct {
	Token string `json:"token"`
}

type createPaymentRequest struct {
	ExternalID  string           `json:"external_id"`
	Amount      float64          `json:"amount"`
	Currency    string           `json:"currency"`
	Mode        string           `json:"mode"`
	CallbackURL string           `json:"callback_url,omitempty"`
	ResultURL   string           `json:"result_url,omitempty"`
	Customer    *paymentCustomer `json:"customer,omitempty"`
}

// AddPaymentMethod sends a hosted payment creation request and returns redirect URL
func (c *Client) AddPaymentMethod(ctx context.Context, req AddPaymentMethodRequest) (AddPaymentMethodResponse, error) {
	orderID, _ := uuid.NewRandom()

	payload := createPaymentRequest{
		ExternalID:  "add-card_" + req.CustomerID + "_" + orderID.String(),
		Amount:      1,
		Currency:    "UAH",
		Mode:        "hosted",
		CallbackURL: c.CallbackURL,
	}

	ar, err := c.post(ctx, "/api/payments/v1/new", payload)
	if err != nil {
		return AddPaymentMethodResponse{}, err
	}

	if ar.ActionRequired && ar.Action.Value != "" {
		return AddPaymentMethodResponse{RedirectURL: ar.Action.Value}, nil
	}
	return AddPaymentMethodResponse{RedirectURL: ar.Action.Value}, nil
}

// CreatePaymentWithSavedCard creates a payment using a saved card token
func (c *Client) CreatePaymentWithSavedCard(ctx context.Context, req CreatePaymentWithTokenRequest) (CreatePaymentWithTokenResponse, error) {
	payload := createPaymentRequest{
		ExternalID:  req.OrderID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Mode:        "direct",
		CallbackURL: c.CallbackURL,
		Customer: &paymentCustomer{
			ExternalID: req.CustomerID,
			PaymentMethod: &paymentMethodBody{
				Type: "cc_token",
				CcToken: &ccTokenFragment{
					Token: req.CustomerToken,
				},
			},
		},
	}

	ar, err := c.post(ctx, "/api/payments/v1/new", payload)
	if err != nil {
		return CreatePaymentWithTokenResponse{}, err
	}

	res := CreatePaymentWithTokenResponse{
		Status: ar.Details.Status,
	}
	if ar.ActionRequired {
		res.RedirectURL = ar.Action.Value
	}
	return res, nil
}
