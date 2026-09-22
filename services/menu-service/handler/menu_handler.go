package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/warungos/menu-service/model"
	"github.com/warungos/menu-service/repository"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MenuHandler handles HTTP requests for menu operations.
type MenuHandler struct {
	repo repository.MenuRepository
}

// NewMenuHandler creates a new MenuHandler with the given repository.
func NewMenuHandler(repo repository.MenuRepository) *MenuHandler {
	return &MenuHandler{repo: repo}
}

// Routes returns a chi router with all menu routes configured.
func (h *MenuHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateMenuItem)
	r.Get("/", h.ListMenuItems)
	r.Get("/stats", h.GetCategoryStats)
	r.Get("/{id}", h.GetMenuItem)
	r.Put("/{id}", h.UpdateMenuItem)
	r.Delete("/{id}", h.DeleteMenuItem)
	return r
}

// CreateMenuItemRequest is the JSON body for creating a menu item.
type CreateMenuItemRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Price       float64          `json:"price"`
	ImageURL    string           `json:"image_url"`
	IsAvailable *bool            `json:"is_available"`
	Variants    []model.Variant  `json:"variants"`
	Modifiers   []model.Modifier `json:"modifiers"`
	BranchIDs   []string         `json:"branch_ids"`
	Tags        []string         `json:"tags"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// CreateMenuItem handles POST /menu
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var req CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	if req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "price must be positive")
		return
	}

	item := &model.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		IsAvailable: true,
		Variants:    req.Variants,
		Modifiers:   req.Modifiers,
		BranchIDs:   req.BranchIDs,
		Tags:        req.Tags,
	}

	if req.IsAvailable != nil {
		item.IsAvailable = *req.IsAvailable
	}

	if err := h.repo.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create menu item")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// ListMenuItems handles GET /menu
func (h *MenuHandler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	items, err := h.repo.GetAll(r.Context(), branchID, category, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list menu items")
		return
	}

	if items == nil {
		items = []model.MenuItem{}
	}

	writeJSON(w, http.StatusOK, items)
}

// GetMenuItem handles GET /menu/{id}
func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "menu item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get menu item")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// UpdateMenuItem handles PUT /menu/{id}
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "menu item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get menu item")
		return
	}

	var req CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Price > 0 {
		existing.Price = req.Price
	}
	if req.ImageURL != "" {
		existing.ImageURL = req.ImageURL
	}
	if req.IsAvailable != nil {
		existing.IsAvailable = *req.IsAvailable
	}
	if req.Variants != nil {
		existing.Variants = req.Variants
	}
	if req.Modifiers != nil {
		existing.Modifiers = req.Modifiers
	}
	if req.BranchIDs != nil {
		existing.BranchIDs = req.BranchIDs
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}

	if err := h.repo.Update(r.Context(), id, existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update menu item")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// DeleteMenuItem handles DELETE /menu/{id}
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "menu item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete menu item")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "menu item deleted"})
}

// GetCategoryStats handles GET /menu/stats
func (h *MenuHandler) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	branchID := r.URL.Query().Get("branch_id")

	stats, err := h.repo.GetCategoryStats(r.Context(), branchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get category stats")
		return
	}

	if stats == nil {
		stats = []model.CategoryStats{}
	}

	writeJSON(w, http.StatusOK, stats)
}
