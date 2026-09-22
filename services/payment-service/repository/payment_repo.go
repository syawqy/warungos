package repository

import (
	"context"
	"fmt"
	"time"

	"warungos/payment-service/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRepository defines the interface for payment data access.
type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id string) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*model.Payment, error)
	UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error
	UpdateMidtransInfo(ctx context.Context, id, token, url string, expiresAt *time.Time) error
	GetPendingPayments(ctx context.Context) ([]model.Payment, error)
}

// PostgresPaymentRepository implements PaymentRepository using PostgreSQL.
type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPaymentRepository creates a new PostgreSQL-backed payment repository.
func NewPostgresPaymentRepository(pool *pgxpool.Pool) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{pool: pool}
}

// Create inserts a new payment record.
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now
	if payment.Currency == "" {
		payment.Currency = "IDR"
	}
	if payment.Status == "" {
		payment.Status = model.StatusPending
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO payments (id, order_id, external_id, amount, currency, method, status, midtrans_token, payment_url, va_number, provider_ref, description, paid_at, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		payment.ID, payment.OrderID, payment.ExternalID, payment.Amount, payment.Currency,
		payment.Method, payment.Status, payment.MidtransToken, payment.PaymentURL,
		payment.VANumber, payment.ProviderRef, payment.Description,
		payment.PaidAt, payment.ExpiresAt, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetByID retrieves a payment by its primary ID.
func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, external_id, amount, currency, method, status, midtrans_token, payment_url, va_number, provider_ref, description, paid_at, expires_at, created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.OrderID, &p.ExternalID, &p.Amount, &p.Currency,
		&p.Method, &p.Status, &p.MidtransToken, &p.PaymentURL,
		&p.VANumber, &p.ProviderRef, &p.Description,
		&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &p, nil
}

// GetByOrderID retrieves a payment by its associated order ID.
func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, external_id, amount, currency, method, status, midtrans_token, payment_url, va_number, provider_ref, description, paid_at, expires_at, created_at, updated_at
		 FROM payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`, orderID,
	).Scan(
		&p.ID, &p.OrderID, &p.ExternalID, &p.Amount, &p.Currency,
		&p.Method, &p.Status, &p.MidtransToken, &p.PaymentURL,
		&p.VANumber, &p.ProviderRef, &p.Description,
		&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by order: %w", err)
	}
	return &p, nil
}

// GetByExternalID retrieves a payment by its external (Midtrans) ID.
func (r *PostgresPaymentRepository) GetByExternalID(ctx context.Context, externalID string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, external_id, amount, currency, method, status, midtrans_token, payment_url, va_number, provider_ref, description, paid_at, expires_at, created_at, updated_at
		 FROM payments WHERE external_id = $1`, externalID,
	).Scan(
		&p.ID, &p.OrderID, &p.ExternalID, &p.Amount, &p.Currency,
		&p.Method, &p.Status, &p.MidtransToken, &p.PaymentURL,
		&p.VANumber, &p.ProviderRef, &p.Description,
		&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by external ID: %w", err)
	}
	return &p, nil
}

// UpdateStatus updates the status and updated_at of a payment.
func (r *PostgresPaymentRepository) UpdateStatus(ctx context.Context, id string, status model.PaymentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	return nil
}

// UpdateMidtransInfo updates the Midtrans token, payment URL, and expiry.
func (r *PostgresPaymentRepository) UpdateMidtransInfo(ctx context.Context, id, token, url string, expiresAt *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET midtrans_token = $1, payment_url = $2, expires_at = $3, updated_at = $4 WHERE id = $5`,
		token, url, expiresAt, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update midtrans info: %w", err)
	}
	return nil
}

// GetPendingPayments retrieves all payments with pending status.
func (r *PostgresPaymentRepository) GetPendingPayments(ctx context.Context) ([]model.Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, external_id, amount, currency, method, status, midtrans_token, payment_url, va_number, provider_ref, description, paid_at, expires_at, created_at, updated_at
		 FROM payments WHERE status = $1 ORDER BY created_at ASC`, model.StatusPending,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending payments: %w", err)
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment
		if err := rows.Scan(
			&p.ID, &p.OrderID, &p.ExternalID, &p.Amount, &p.Currency,
			&p.Method, &p.Status, &p.MidtransToken, &p.PaymentURL,
			&p.VANumber, &p.ProviderRef, &p.Description,
			&p.PaidAt, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, nil
}
