package handler

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"warungos/payment-service/integration"
	"warungos/payment-service/model"
	"warungos/payment-service/repository"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

const (
	webhookDedupPrefix = "payment:webhook:dedup:"
	webhookDedupTTL   = 24 * time.Hour
)

// PaymentHandler handles HTTP requests for payment operations.
type PaymentHandler struct {
	repo          repository.PaymentRepository
	midtrans      *integration.MidtransClient
	redis         *redis.Client
	webhookSecret string
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(repo repository.PaymentRepository, midtrans *integration.MidtransClient, rdb *redis.Client) *PaymentHandler {
	return &PaymentHandler{
		repo:          repo,
		midtrans:      midtrans,
		redis:         rdb,
		webhookSecret: os.Getenv("MIDTRANS_WEBHOOK_SECRET"),
	}
}

// Routes returns a chi router with all payment routes configured.
func (h *PaymentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreatePayment)
	r.Get("/{id}", h.GetPayment)
	r.Get("/order/{order_id}", h.GetPaymentByOrderID)
	r.Post("/webhook", h.HandleWebhook)
	r.Post("/webhook/midtrans", h.HandleMidtransWebhook)
	return r
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// CreatePayment handles POST /payments - creates a new payment with Midtrans integration.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OrderID == "" {
		writeError(w, http.StatusBadRequest, "order_id is required")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Method == "" {
		writeError(w, http.StatusBadRequest, "method is required")
		return
	}

	// Create local payment record
	payment := &model.Payment{
		ID:          generatePaymentID(),
		OrderID:     req.OrderID,
		ExternalID:  req.OrderID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Status:      model.StatusPending,
		Description: req.Description,
	}

	if payment.Currency == "" {
		payment.Currency = "IDR"
	}

	if err := h.repo.Create(r.Context(), payment); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment record")
		return
	}

	// Create Midtrans transaction with retry
	midtransReq := integration.CreateTransactionRequest{}
	midtransReq.TransactionDetails.OrderID = payment.OrderID
	midtransReq.TransactionDetails.GrossAmount = payment.Amount

	midtransResp, err := h.midtrans.CreateTransaction(r.Context(), midtransReq)
	if err != nil {
		log.Printf("payment handler: midtrans create failed for payment %s: %v", payment.ID, err)
		// Still return the payment but note the error
		writeJSON(w, http.StatusCreated, payment)
		return
	}

	// Update with Midtrans info
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := h.repo.UpdateMidtransInfo(r.Context(), payment.ID, midtransResp.Token, midtransResp.RedirectURL, &expiresAt); err != nil {
		log.Printf("payment handler: failed to update midtrans info for payment %s: %v", payment.ID, err)
	}

	payment.MidtransToken = midtransResp.Token
	payment.PaymentURL = midtransResp.RedirectURL
	payment.ExpiresAt = &expiresAt

	writeJSON(w, http.StatusCreated, payment)
}

// GetPayment handles GET /payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "payment ID is required")
		return
	}

	payment, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "payment not found")
		return
	}

	writeJSON(w, http.StatusOK, payment)
}

// GetPaymentByOrderID handles GET /payments/order/{order_id}
func (h *PaymentHandler) GetPaymentByOrderID(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "order_id is required")
		return
	}

	payment, err := h.repo.GetByOrderID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "payment not found for this order")
		return
	}

	writeJSON(w, http.StatusOK, payment)
}

// HandleWebhook handles POST /payments/webhook - generic webhook endpoint.
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var notif model.MidtransNotification
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	h.processNotification(w, r.Context(), notif)
}

// HandleMidtransWebhook handles POST /payments/webhook/midtrans - Midtrans-specific webhook.
func (h *PaymentHandler) HandleMidtransWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// Verify signature if webhook secret is set
	if h.webhookSecret != "" {
		orderID := r.URL.Query().Get("order_id")
		statusCode := r.URL.Query().Get("status_code")
		grossAmount := r.URL.Query().Get("gross_amount")

		expectedSig := r.URL.Query().Get("signature_key")
		sigInput := orderID + statusCode + grossAmount + h.webhookSecret
		hash := sha512.Sum512([]byte(sigInput))
		actualSig := hex.EncodeToString(hash[:])

		if expectedSig != actualSig {
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}

	var notif model.MidtransNotification
	if err := json.Unmarshal(body, &notif); err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification payload")
		return
	}

	h.processNotification(w, r.Context(), notif)
}

// processNotification handles the common webhook processing logic with idempotency via Redis.
func (h *PaymentHandler) processNotification(w http.ResponseWriter, ctx context.Context, notif model.MidtransNotification) {
	// Idempotency check via Redis
	dedupKey := fmt.Sprintf("%s%s:%s", webhookDedupPrefix, notif.OrderID, notif.TransactionStatus)

	if h.redis != nil {
		exists, err := h.redis.Exists(ctx, dedupKey).Result()
		if err == nil && exists > 0 {
			log.Printf("webhook: duplicate notification for order %s, status %s - skipping", notif.OrderID, notif.TransactionStatus)
			writeJSON(w, http.StatusOK, map[string]string{"status": "already_processed"})
			return
		}
	}

	// Map status
	newStatus := integration.MapNotificationToPayment(notif)

	// Find payment by order ID
	payment, err := h.repo.GetByOrderID(ctx, notif.OrderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "payment not found for order")
		return
	}

	// Update status
	if err := h.repo.UpdateStatus(ctx, payment.ID, newStatus); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update payment status")
		return
	}

	// Set dedup key in Redis
	if h.redis != nil {
		h.redis.Set(ctx, dedupKey, "1", webhookDedupTTL)
	}

	log.Printf("webhook: payment %s updated to %s (order: %s)", payment.ID, newStatus, notif.OrderID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func generatePaymentID() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}
