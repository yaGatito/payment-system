package rozetkaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"payment-system/internal/domain"

	"github.com/google/uuid"
)

const rozetkaPaymentsPath = "/api/payments/v1/new"
const tokenPaymentType = "cc_token"
const retryDeltaTimeout = 750 * time.Millisecond

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

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
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
			lastErr = err
			if attempt < 2 {
				select {
				case <-ctx.Done():
					return nil, err
				case <-time.After(time.Duration(attempt+1) * retryDeltaTimeout):
				}
			}
			continue
		}

		data, _ := io.ReadAll(resp.Body)
		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("closing body error")
		}
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("bad status %d: %s", resp.StatusCode, string(data))
			if attempt < 2 {
				select {
				case <-ctx.Done():
					return nil, lastErr
				case <-time.After(time.Duration(attempt+1) * retryDeltaTimeout):
				}
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(data))
		}

		var ar apiResponse
		if err := json.Unmarshal(data, &ar); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
		return &ar, nil
	}
	return nil, lastErr
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
	Amount      int64            `json:"amount"`
	Currency    string           `json:"currency"`
	Mode        string           `json:"mode"`
	CallbackURL string           `json:"callback_url,omitempty"`
	ResultURL   string           `json:"result_url,omitempty"`
	Customer    *paymentCustomer `json:"customer,omitempty"`
}

// AddPaymentMethod sends a hosted payment creation request and returns redirect URL
func (c *Client) AddPaymentMethod(
	ctx context.Context,
	req AddPaymentMethodRequest,
) (AddPaymentMethodResponse, error) {
	orderID, _ := uuid.NewRandom()

	payload := createPaymentRequest{
		ExternalID: domain.BuildExternalID(
			domain.AddCardEventType,
			req.CustomerID,
			orderID.String(),
		),
		Amount:      1,
		Currency:    "UAH",
		Mode:        "hosted",
		CallbackURL: c.CallbackURL,
	}

	ar, err := c.post(ctx, rozetkaPaymentsPath, payload)
	if err != nil {
		return AddPaymentMethodResponse{}, err
	}

	if ar.ActionRequired && ar.Action.Value != "" {
		return AddPaymentMethodResponse{RedirectURL: ar.Action.Value}, nil
	}
	return AddPaymentMethodResponse{RedirectURL: ar.Action.Value}, nil
}

// CreatePaymentWithSavedCard creates a payment using a saved card token
func (c *Client) CreatePaymentWithSavedCard(
	ctx context.Context,
	req CreatePaymentWithTokenRequest,
) (CreatePaymentWithTokenResponse, error) {
	payload := createPaymentRequest{
		ExternalID:  domain.BuildExternalID(domain.PaymentEventType, req.CustomerID, req.OrderID),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Mode:        "direct",
		CallbackURL: c.CallbackURL,
		Customer: &paymentCustomer{
			ExternalID: req.CustomerID,
			PaymentMethod: &paymentMethodBody{
				Type: tokenPaymentType,
				CcToken: &ccTokenFragment{
					Token: req.CustomerToken,
				},
			},
		},
	}

	ar, err := c.post(ctx, rozetkaPaymentsPath, payload)
	if err != nil {
		return CreatePaymentWithTokenResponse{}, err
	}

	if !domain.ValidateStatus(ar.Details.Status) {
		return CreatePaymentWithTokenResponse{}, fmt.Errorf(
			"integration error: status not recognized",
		)
	}

	res := CreatePaymentWithTokenResponse{
		Status:      ar.Details.Status,
		RedirectURL: ar.Action.Value,
	}

	if ar.ActionRequired {
		res.RedirectURL = ar.Action.Value
	}
	return res, nil
}
