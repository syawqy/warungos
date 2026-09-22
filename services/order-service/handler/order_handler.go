package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/warungos/order-service/model"
	"github.com/warungos/order-service/service"
	"github.com/warungos/shared/middleware"
)

// OrderHandler holds dependencies for order endpoints.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.BranchID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, &req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order id is required"})
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// UpdateOrderStatus handles PATCH /orders/{id}/status.
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order id is required"})
		return
	}

	var req model.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateOrderStatus(r.Context(), id, req.Status); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

// ListOrders handles GET /orders with query parameters.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := model.OrderListQuery{
		BranchID: r.URL.Query().Get("branch_id"),
		UserID:   r.URL.Query().Get("user_id"),
		Status:   r.URL.Query().Get("status"),
	}

	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		q.Page = p
	} else {
		q.Page = 1
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil {
		q.PageSize = ps
	} else {
		q.PageSize = 20
	}

	orders, total, err := h.svc.ListOrders(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
		"pagination": map[string]interface{}{
			"page":      q.Page,
			"page_size": q.PageSize,
			"total":     total,
		},
	})
}

// CountByBranchAndDate handles GET /orders/count?branch_id=xxx&date=2026-01-01.
func (h *OrderHandler) CountByBranchAndDate(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
		return
	}

	dateStr := r.URL.Query().Get("date")
	date := time.Now()
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
	}

	count, err := h.svc.CountByBranchAndDate(r.Context(), branchID, date)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"branch_id": branchID,
		"date":      date.Format("2006-01-02"),
		"count":     count,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
