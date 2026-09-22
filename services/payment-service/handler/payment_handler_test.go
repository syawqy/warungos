package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"warungos/payment-service/integration"
	"warungos/payment-service/model"

	"github.com/go-chi/chi/v5"
)

// stubPaymentRepo is a minimal in-memory implementation of PaymentRepository.
type stubPaymentRepo struct {
	payments map[string]*model.Payment
}

func newStubPaymentRepo() *stubPaymentRepo {
	return &stubPaymentRepo{payments: make(map[string]*model.Payment)}
}

func (s *stubPaymentRepo) Create(_ context.Context, p *model.Payment) error {
	if p.ID == "" {
		p.ID = "pay_test_1"
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	s.payments[p.ID] = p
	return nil
}

func (s *stubPaymentRepo) GetByID(_ context.Context, id string) (*model.Payment, error) {
	p, ok := s.payments[id]
	if !ok {
		return nil, json.Unmarshal([]byte("null"), p)
	}
	return p, nil
}

func (s *stubPaymentRepo) GetByOrderID(_ context.Context, orderID string) (*model.Payment, error) {
	for _, p := range s.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, json.Unmarshal([]byte("null"), new(model.Payment))
}

func (s *stubPaymentRepo) GetByExternalID(_ context.Context, externalID string) (*model.Payment, error) {
	for _, p := range s.payments {
		if p.ExternalID == externalID {
			return p, nil
		}
	}
	return nil, json.Unmarshal([]byte("null"), new(model.Payment))
}

func (s *stubPaymentRepo) UpdateStatus(_ context.Context, id string, status model.PaymentStatus) error {
	if p, ok := s.payments[id]; ok {
		p.Status = status
		p.UpdatedAt = time.Now()
		if status == model.StatusPaid {
			now := time.Now()
			p.PaidAt = &now
		}
	}
	return nil
}

func (s *stubPaymentRepo) UpdateMidtransInfo(_ context.Context, id, token, url string, expiresAt *time.Time) error {
	if p, ok := s.payments[id]; ok {
		p.MidtransToken = token
		p.PaymentURL = url
		p.ExpiresAt = expiresAt
		p.UpdatedAt = time.Now()
	}
	return nil
}

func (s *stubPaymentRepo) GetPendingPayments(_ context.Context) ([]model.Payment, error) {
	var pending []model.Payment
	for _, p := range s.payments {
		if p.Status == model.StatusPending {
			pending = append(pending, *p)
		}
	}
	return pending, nil
}

func TestCreatePayment(t *testing.T) {
	repo := newStubPaymentRepo()

	// Mock Midtrans server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := integration.CreateTransactionResponse{
			Token:       "test-token-abc123",
			RedirectURL: "https://app.sandbox.midtrans.com/snap/vtweb/test-abc123",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	midtransClient := integration.NewMidtransClient("test-server-key", mockServer.URL)
	h := NewPaymentHandler(repo, midtransClient, nil)

	reqBody := model.CreatePaymentRequest{
		OrderID:  "ORD-001",
		Amount:   50000,
		Currency: "IDR",
		Method:   model.MethodBankTransfer,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreatePayment(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var payment model.Payment
	if err := json.NewDecoder(w.Body).Decode(&payment); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payment.OrderID != "ORD-001" {
		t.Errorf("expected order_id 'ORD-001', got '%s'", payment.OrderID)
	}
	if payment.Amount != 50000 {
		t.Errorf("expected amount 50000, got %f", payment.Amount)
	}
	if payment.MidtransToken != "test-token-abc123" {
		t.Errorf("expected midtrans token 'test-token-abc123', got '%s'", payment.MidtransToken)
	}
	if payment.Status != model.StatusPending {
		t.Errorf("expected status 'pending', got '%s'", payment.Status)
	}
}

func TestCreatePaymentMissingOrderID(t *testing.T) {
	repo := newStubPaymentRepo()
	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil)

	reqBody := model.CreatePaymentRequest{
		Amount: 50000,
		Method: model.MethodBankTransfer,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreatePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetPayment(t *testing.T) {
	repo := newStubPaymentRepo()
	repo.payments["pay_test_1"] = &model.Payment{
		ID:      "pay_test_1",
		OrderID: "ORD-002",
		Amount:  75000,
		Status:  model.StatusPending,
	}

	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil)

	r := chi.NewRouter()
	r.Get("/payments/{id}", h.GetPayment)

	req := httptest.NewRequest(http.MethodGet, "/payments/pay_test_1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payment model.Payment
	if err := json.NewDecoder(w.Body).Decode(&payment); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payment.ID != "pay_test_1" {
		t.Errorf("expected id 'pay_test_1', got '%s'", payment.ID)
	}
}

func TestGetPaymentNotFound(t *testing.T) {
	repo := newStubPaymentRepo()
	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil)

	r := chi.NewRouter()
	r.Get("/payments/{id}", h.GetPayment)

	req := httptest.NewRequest(http.MethodGet, "/payments/pay_nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestGetPaymentByOrderID(t *testing.T) {
	repo := newStubPaymentRepo()
	repo.payments["pay_test_1"] = &model.Payment{
		ID:      "pay_test_1",
		OrderID: "ORD-003",
		Amount:  100000,
		Status:  model.StatusPaid,
	}

	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil)

	r := chi.NewRouter()
	r.Get("/payments/order/{order_id}", h.GetPaymentByOrderID)

	req := httptest.NewRequest(http.MethodGet, "/payments/order/ORD-003", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payment model.Payment
	if err := json.NewDecoder(w.Body).Decode(&payment); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payment.OrderID != "ORD-003" {
		t.Errorf("expected order_id 'ORD-003', got '%s'", payment.OrderID)
	}
}

func TestHandleWebhook(t *testing.T) {
	repo := newStubPaymentRepo()
	repo.payments["pay_test_1"] = &model.Payment{
		ID:      "pay_test_1",
		OrderID: "ORD-004",
		Amount:  50000,
		Status:  model.StatusPending,
	}

	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil)

	notif := model.MidtransNotification{
		OrderID:           "ORD-004",
		TransactionStatus: "settlement",
		StatusCode:        "200",
	}
	body, _ := json.Marshal(notif)
	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify payment was updated
	payment := repo.payments["pay_test_1"]
	if payment.Status != model.StatusPaid {
		t.Errorf("expected status 'paid', got '%s'", payment.Status)
	}
}

func TestHandleWebhookWithoutRedis(t *testing.T) {
	repo := newStubPaymentRepo()
	repo.payments["pay_test_1"] = &model.Payment{
		ID:      "pay_test_1",
		OrderID: "ORD-005",
		Amount:  50000,
		Status:  model.StatusPending,
	}

	midtransClient := integration.NewMidtransClient("test-key", "http://localhost")
	h := NewPaymentHandler(repo, midtransClient, nil) // nil Redis

	notif := model.MidtransNotification{
		OrderID:           "ORD-005",
		TransactionStatus: "settlement",
		StatusCode:        "200",
	}
	body, _ := json.Marshal(notif)

	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	payment := repo.payments["pay_test_1"]
	if payment.Status != model.StatusPaid {
		t.Errorf("expected status 'paid', got '%s'", payment.Status)
	}
}

func TestPaymentModelJSON(t *testing.T) {
	now := time.Now()
	payment := model.Payment{
		ID:            "pay_123",
		OrderID:       "ORD-006",
		Amount:        100000,
		Currency:      "IDR",
		Method:        model.MethodQRIS,
		Status:        model.StatusPending,
		MidtransToken: "abc-token",
		PaymentURL:    "https://midtrans.com/pay/abc",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded model.Payment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != payment.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Amount != payment.Amount {
		t.Errorf("Amount mismatch")
	}
	if decoded.Method != model.MethodQRIS {
		t.Errorf("Method mismatch")
	}
}

func TestMapNotificationStatus(t *testing.T) {
	tests := []struct {
		txStatus string
		expected model.PaymentStatus
	}{
		{"settlement", model.StatusPaid},
		{"capture", model.StatusPaid},
		{"pending", model.StatusPending},
		{"deny", model.StatusFailed},
		{"cancel", model.StatusFailed},
		{"expire", model.StatusExpired},
		{"unknown_status", model.StatusPending},
	}

	for _, tt := range tests {
		result := integration.MapTransactionStatus(tt.txStatus)
		if result != tt.expected {
			t.Errorf("MapTransactionStatus(%q) = %q, want %q", tt.txStatus, result, tt.expected)
		}
	}
}
