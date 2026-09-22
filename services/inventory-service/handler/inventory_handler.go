package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/warungos/inventory-service/model"
	"github.com/warungos/inventory-service/repository"
	sharedModel "github.com/warungos/shared/model"
	"time"
)

type InventoryHandler struct {
	repo *repository.InventoryRepo
}

func NewInventoryHandler(repo *repository.InventoryRepo) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "branch_id is required"})
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
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sharedModel.APIResponse{
		Success:    true,
		Data:       items,
		Pagination: &sharedModel.Pagination{Page: page, PageSize: limit, Total: int(total)},
		Timestamp:  time.Now(),
	})
}

func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid request body"})
		return
	}
	item := &model.InventoryItem{
		BranchID:    req.BranchID,
		ItemName:    req.ItemName,
		ItemCode:    req.ItemCode,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinQuantity: req.MinQuantity,
		MaxQuantity: req.MaxQuantity,
		UnitCost:    req.UnitCost,
	}
	if err := h.repo.Create(r.Context(), item); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, sharedModel.SuccessResponse(item))
}

func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, sharedModel.SuccessResponse(item))
}

func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid body"})
		return
	}
	item, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sharedModel.SuccessResponse(item))
}

func (h *InventoryHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	if branchID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "branch_id is required"})
		return
	}
	items, err := h.repo.GetLowStockItems(r.Context(), branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sharedModel.SuccessResponse(items))
}

func (h *InventoryHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req model.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "invalid body"})
		return
	}
	if req.Quantity <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "quantity must be positive"})
		return
	}
	if err := h.repo.ReserveStock(r.Context(), req.ItemID, req.Quantity); err != nil {
		writeJSON(w, http.StatusConflict, map[string]interface{}{"error": err.Error()})
		return
	}
	item, _ := h.repo.FindByID(r.Context(), req.ItemID)
	writeJSON(w, http.StatusOK, sharedModel.SuccessResponse(item))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
