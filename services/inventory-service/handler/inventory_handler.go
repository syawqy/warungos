package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/warungos/inventory-service/model"
	"github.com/warungos/inventory-service/repository"
	sharedModel "github.com/warungos/shared/model"
)

// InventoryHandler handles HTTP requests for inventory
type InventoryHandler struct {
	repo *repository.InventoryRepo
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(repo *repository.InventoryRepo) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

// List handles GET /api/v1/inventory
func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "branch_id is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := h.repo.List(r.Context(), branchID, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, items, &sharedModel.Pagination{
		Page:     page,
		PageSize: limit,
		Total:    int(total),
	})
}

// Create handles POST /api/v1/inventory
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item := &model.InventoryItem{
		BranchID:    req.BranchID,
		ItemName:    req.ItemName,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinStock:    req.MinStock,
		CostPerUnit: req.CostPerUnit,
	}

	if err := h.repo.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, item, nil)
}

// GetByID handles GET /api/v1/inventory/:id
func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "inventory item not found")
		return
	}
	writeSuccess(w, http.StatusOK, item, nil)
}

// Update handles PATCH /api/v1/inventory/:id
func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, item, nil)
}

// GetAlerts handles GET /api/v1/inventory/alerts
func (h *InventoryHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "branch_id is required")
		return
	}

	items, err := h.repo.GetLowStockItems(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, items, nil)
}

// Reserve handles POST /api/v1/inventory/reserve
func (h *InventoryHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req model.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	if err := h.repo.ReserveStock(r.Context(), req.ItemID, req.Quantity); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	item, _ := h.repo.FindByID(r.Context(), req.ItemID)
	writeSuccess(w, http.StatusOK, item, nil)
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}, meta *sharedModel.Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(sharedModel.APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(sharedModel.APIResponse{
		Success: false,
		Error:   message,
	})
}
