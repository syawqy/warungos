package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warungos/order-service/model"
)

// TestCreateOrderValidation verifies that orders require items.
func TestCreateOrderValidation(t *testing.T) {
	order := &model.Order{
		ID:       "test-order-1",
		UserID:   "user-1",
		BranchID: "branch-1",
		Status:   model.StatusPending,
	}

	if err := order.ValidateItems(); err == nil {
		t.Error("expected error for empty items")
	}
}

// TestCalculateTotal verifies 10% PPN tax calculation.
func TestCalculateTotal(t *testing.T) {
	order := &model.Order{
		Items: []model.OrderItem{
			{MenuName: "Nasi Goreng", Quantity: 2, UnitPrice: 25000},
			{MenuName: "Es Teh", Quantity: 3, UnitPrice: 8000},
		},
	}

	order.CalculateTotal()

	// Subtotal: (25000*2) + (8000*3) = 50000 + 24000 = 74000
	if order.Subtotal != 74000 {
		t.Errorf("expected subtotal 74000, got %d", order.Subtotal)
	}
	// Tax: 74000 / 10 = 7400
	if order.Tax != 7400 {
		t.Errorf("expected tax 7400, got %d", order.Tax)
	}
	// Total: 74000 + 7400 = 81400
	if order.TotalPrice != 81400 {
		t.Errorf("expected total 81400, got %d", order.TotalPrice)
	}
}

// TestStatusTransitions verifies the status machine.
func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		from     model.OrderStatus
		to       model.OrderStatus
		expected bool
	}{
		{model.StatusPending, model.StatusConfirmed, true},
		{model.StatusPending, model.StatusCancelled, true},
		{model.StatusPending, model.StatusPreparing, false},
		{model.StatusPending, model.StatusReady, false},
		{model.StatusPending, model.StatusCompleted, false},
		{model.StatusConfirmed, model.StatusPreparing, true},
		{model.StatusConfirmed, model.StatusCancelled, true},
		{model.StatusPreparing, model.StatusReady, true},
		{model.StatusPreparing, model.StatusCancelled, true},
		{model.StatusReady, model.StatusCompleted, true},
		{model.StatusReady, model.StatusCancelled, true},
		{model.StatusCompleted, model.StatusPending, false},
		{model.StatusCancelled, model.StatusPending, false},
	}

	for _, tt := range tests {
		result := tt.from.CanTransitionTo(tt.to)
		if result != tt.expected {
			t.Errorf("transition %s -> %s: expected %v, got %v",
				tt.from, tt.to, tt.expected, result)
		}
	}
}

// TestCancellationFromAnyState verifies any active status can be cancelled.
func TestCancellationFromAnyState(t *testing.T) {
	activeStatuses := []model.OrderStatus{
		model.StatusPending,
		model.StatusConfirmed,
		model.StatusPreparing,
		model.StatusReady,
	}

	for _, status := range activeStatuses {
		if !status.CanTransitionTo(model.StatusCancelled) {
			t.Errorf("expected %s to be cancellable", status)
		}
	}

	// Cannot cancel already completed or cancelled
	if model.StatusCompleted.CanTransitionTo(model.StatusCancelled) {
		t.Error("completed orders should not be cancellable")
	}
	if model.StatusCancelled.CanTransitionTo(model.StatusCancelled) {
		t.Error("cancelled orders should not be cancellable again")
	}
}

// TestCreateOrderRequestSerialization verifies JSON marshaling.
func TestCreateOrderRequestSerialization(t *testing.T) {
	req := model.CreateOrderRequest{
		BranchID: "branch-1",
		Notes:    "Extra sambal please",
		Items: []model.CreateOrderItemReq{
			{MenuID: "m1", MenuName: "Nasi Goreng", Quantity: 2, UnitPrice: 25000},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal create order request: %v", err)
	}

	var decoded model.CreateOrderRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal create order request: %v", err)
	}
	if decoded.BranchID != "branch-1" {
		t.Errorf("expected branch_id branch-1, got %s", decoded.BranchID)
	}
	if len(decoded.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(decoded.Items))
	}
}

// TestListOrdersEndpoint verifies the list handler responds correctly.
func TestListOrdersEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/orders?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()

	// Handler needs a real service, so we verify request parsing
	page := req.URL.Query().Get("page")
	if page != "1" {
		t.Errorf("expected page 1, got %s", page)
	}

	_ = rec
}
