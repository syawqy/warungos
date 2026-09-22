package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"warungos/payment-service/model"
)

const (
	maxRetries    = 5
	initBackoffMs = 500 * time.Millisecond
	maxBackoff    = 30 * time.Second
	backoffFactor = 2.0
)

// MidtransClient handles communication with the Midtrans Snap API.
type MidtransClient struct {
	serverKey  string
	baseURL    string
	httpClient *http.Client
}

// NewMidtransClient creates a new Midtrans API client.
func NewMidtransClient(serverKey, baseURL string) *MidtransClient {
	return &MidtransClient{
		serverKey: serverKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateTransactionRequest represents the payload for creating a Midtrans transaction.
type CreateTransactionRequest struct {
	TransactionDetails struct {
		OrderID     string  `json:"order_id"`
		GrossAmount float64 `json:"gross_amount"`
	} `json:"transaction_details"`
	CustomerDetails *CustomerDetails `json:"customer_details,omitempty"`
	EnabledPayments []string         `json:"enabled_payments,omitempty"`
	Expiry          *Expiry          `json:"expiry,omitempty"`
}

// CustomerDetails holds customer info for Midtrans.
type CustomerDetails struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// Expiry defines payment expiry settings.
type Expiry struct {
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	Unit      string `json:"unit"`
}

// CreateTransactionResponse is the response from Midtrans Snap API.
type CreateTransactionResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// TransactionStatusResponse represents the status response from Midtrans.
type TransactionStatusResponse struct {
	TransactionID     string `json:"transaction_id"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
	Bank              string `json:"bank,omitempty"`
	VA                string `json:"va_number,omitempty"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	ExpiryTime        string `json:"expiry_time,omitempty"`
}

// CreateTransaction creates a new payment transaction in Midtrans with retry and exponential backoff.
func (c *MidtransClient) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	var lastErr error
	backoff := initBackoffMs

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("midtrans: retry attempt %d/%d, sleeping %v", attempt, maxRetries, backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff = time.Duration(math.Min(float64(maxBackoff), float64(backoff)*backoffFactor))
		}

		resp, err := c.doRequest(ctx, http.MethodPost, "/snap/v1/transactions", req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var result CreateTransactionResponse
			if err := json.Unmarshal(body, &result); err != nil {
				lastErr = fmt.Errorf("failed to decode response: %w", err)
				continue
			}
			return &result, nil
		}

		lastErr = fmt.Errorf("midtrans API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("midtrans: all %d retries exhausted: %w", maxRetries+1, lastErr)
}

// GetStatus queries the transaction status from Midtrans with retry and exponential backoff.
func (c *MidtransClient) GetStatus(ctx context.Context, orderID string) (*TransactionStatusResponse, error) {
	var lastErr error
	backoff := initBackoffMs

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("midtrans status: retry attempt %d/%d, sleeping %v", attempt, maxRetries, backoff)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff = time.Duration(math.Min(float64(maxBackoff), float64(backoff)*backoffFactor))
		}

		resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v2/%s/status", orderID), nil)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result TransactionStatusResponse
			if err := json.Unmarshal(body, &result); err != nil {
				lastErr = fmt.Errorf("failed to decode response: %w", err)
				continue
			}
			return &result, nil
		}

		lastErr = fmt.Errorf("midtrans status API returned %d: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("midtrans status: all %d retries exhausted: %w", maxRetries+1, lastErr)
}

// MapNotificationToPayment maps a Midtrans notification to a payment status.
func MapNotificationToPayment(notif model.MidtransNotification) model.PaymentStatus {
	return MapTransactionStatus(notif.TransactionStatus)
}

// MapTransactionStatus maps a Midtrans transaction status string to a PaymentStatus.
func MapTransactionStatus(txStatus string) model.PaymentStatus {
	switch txStatus {
	case "capture", "settlement":
		return model.StatusPaid
	case "pending":
		return model.StatusPending
	case "deny", "cancel", "failure":
		return model.StatusFailed
	case "expire":
		return model.StatusExpired
	default:
		return model.StatusPending
	}
}

// doRequest executes an HTTP request with the Midtrans server key.
func (c *MidtransClient) doRequest(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.serverKey, "")

	return c.httpClient.Do(req)
}
