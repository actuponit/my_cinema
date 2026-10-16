package repositories

import (
	"bytes"
	"cinema-server/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type VerifyCheckoutClientInterface interface {
	CreateDeposit(idempotencyKey, customerID, returnURL string) (domain.Deposit, error)
	GetDeposit(depositID string) (domain.Deposit, error)
}

type VerifyCheckoutClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewVerifyCheckoutClient(baseURL, apiKey string) *VerifyCheckoutClient {
	return &VerifyCheckoutClient{baseURL: baseURL, apiKey: apiKey, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

type verifyEnvelope struct {
	Data *struct {
		ID                 string     `json:"id"`
		Status             string     `json:"status"`
		Amount             *string    `json:"amount"`
		MerchantCustomerID string     `json:"merchant_customer_id"`
		CheckoutURL        string     `json:"checkout_url"`
		ExpiresAt          *time.Time `json:"expires_at"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Meta struct {
		RequestID string `json:"requestId"`
	} `json:"meta"`
}

// CreateDeposit always creates a variable-amount deposit: the payer enters the amount at checkout.
// The create body is strict; merchant_order_id/metadata are not accepted.
func (c *VerifyCheckoutClient) CreateDeposit(idempotencyKey, customerID, returnURL string) (domain.Deposit, error) {
	body, err := json.Marshal(map[string]string{
		"merchant_customer_id": customerID,
		"currency":             "ETB",
		"return_url":           returnURL,
	})
	if err != nil {
		return domain.Deposit{}, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/deposits", bytes.NewReader(body))
	if err != nil {
		return domain.Deposit{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)
	return c.do(req)
}

func (c *VerifyCheckoutClient) GetDeposit(depositID string) (domain.Deposit, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/v1/deposits/"+url.PathEscape(depositID), nil)
	if err != nil {
		return domain.Deposit{}, err
	}
	return c.do(req)
}

func (c *VerifyCheckoutClient) do(req *http.Request) (domain.Deposit, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("VerifyCheckout-Version", "2026-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.Deposit{}, fmt.Errorf("verify.et %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return domain.Deposit{}, fmt.Errorf("verify.et read response: %w", err)
	}
	var env verifyEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return domain.Deposit{}, fmt.Errorf("verify.et HTTP %d: undecodable body: %w", resp.StatusCode, err)
	}
	if env.Error != nil {
		return domain.Deposit{}, fmt.Errorf("verify.et %s: %s (request %s)", env.Error.Code, env.Error.Message, env.Meta.RequestID)
	}
	if resp.StatusCode >= 300 || env.Data == nil {
		return domain.Deposit{}, fmt.Errorf("verify.et HTTP %d without data (request %s)", resp.StatusCode, env.Meta.RequestID)
	}

	d := env.Data
	return domain.Deposit{
		ID: d.ID, Status: d.Status, Amount: deref(d.Amount), CustomerID: d.MerchantCustomerID,
		CheckoutURL: d.CheckoutURL, ExpiresAt: d.ExpiresAt,
	}, nil
}
