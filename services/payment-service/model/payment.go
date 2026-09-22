package model

import (
	"time"
)

// PaymentStatus represents the lifecycle state of a payment.
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusPaid      PaymentStatus = "paid"
	StatusFailed    PaymentStatus = "failed"
	StatusExpired   PaymentStatus = "expired"
	StatusCancelled PaymentStatus = "cancelled"
)

// PaymentMethod represents the type of payment method used.
type PaymentMethod string

const (
	MethodBankTransfer PaymentMethod = "bank_transfer"
	MethodCreditCard   PaymentMethod = "credit_card"
	MethodEWallet      PaymentMethod = "ewallet"
	MethodVA           PaymentMethod = "virtual_account"
	MethodQRIS         PaymentMethod = "qris"
)

// Payment is the core domain object for payments.
type Payment struct {
	ID            string        `json:"id" db:"id"`
	OrderID       string        `json:"order_id" db:"order_id"`
	ExternalID    string        `json:"external_id" db:"external_id"`
	Amount        float64       `json:"amount" db:"amount"`
	Currency      string        `json:"currency" db:"currency"`
	Method        PaymentMethod `json:"method" db:"method"`
	Status        PaymentStatus `json:"status" db:"status"`
	MidtransToken string        `json:"midtrans_token" db:"midtrans_token"`
	PaymentURL    string        `json:"payment_url" db:"payment_url"`
	VANumber      string        `json:"va_number" db:"va_number"`
	ProviderRef   string        `json:"provider_ref" db:"provider_ref"`
	Description   string        `json:"description" db:"description"`
	PaidAt        *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// MidtransNotification represents a webhook notification from Midtrans.
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	Acquirer          string `json:"acquirer,omitempty"`
	Bank              string `json:"bank,omitempty"`
	VANumbers         []struct {
		Bank     string `json:"bank"`
		VANumber string `json:"va_number"`
	} `json:"va_numbers,omitempty"`
	ExpiryTime string `json:"expiry_time,omitempty"`
}

// CreatePaymentRequest is the JSON body for creating a payment.
type CreatePaymentRequest struct {
	OrderID     string        `json:"order_id" validate:"required"`
	Amount      float64       `json:"amount" validate:"required"`
	Currency    string        `json:"currency"`
	Method      PaymentMethod `json:"method" validate:"required"`
	Description string        `json:"description"`
}
