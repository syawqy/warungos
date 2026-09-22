package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/warungos/menu-service/model"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// stubRepo is a minimal in-memory implementation of repository.MenuRepository.
type stubRepo struct {
	items map[bson.ObjectID]*model.MenuItem
}

func newStubRepo() *stubRepo {
	return &stubRepo{items: make(map[bson.ObjectID]*model.MenuItem)}
}

func (s *stubRepo) Create(_ context.Context, item *model.MenuItem) error {
	item.ID = bson.NewObjectID()
	s.items[item.ID] = item
	return nil
}

func (s *stubRepo) GetByID(_ context.Context, id bson.ObjectID) (*model.MenuItem, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, bson.ErrDecodeInvalidValue
	}
	return item, nil
}

func (s *stubRepo) GetAll(_ context.Context, _, _, _ string) ([]model.MenuItem, error) {
	var items []model.MenuItem
	for _, v := range s.items {
		items = append(items, *v)
	}
	return items, nil
}

func (s *stubRepo) Update(_ context.Context, id bson.ObjectID, item *model.MenuItem) error {
	if _, ok := s.items[id]; !ok {
		return bson.ErrDecodeInvalidValue
	}
	item.ID = id
	s.items[id] = item
	return nil
}

func (s *stubRepo) Delete(_ context.Context, id bson.ObjectID) error {
	if _, ok := s.items[id]; !ok {
		return bson.ErrDecodeInvalidValue
	}
	delete(s.items, id)
	return nil
}

func (s *stubRepo) GetCategoryStats(_ context.Context, _ string) ([]model.CategoryStats, error) {
	return []model.CategoryStats{
		{Category: "Food", TotalItems: 2, AvgPrice: 25000, MinPrice: 15000, MaxPrice: 35000},
	}, nil
}

func TestCreateMenuItem(t *testing.T) {
	repo := newStubRepo()
	h := NewMenuHandler(repo)

	reqBody := CreateMenuItemRequest{
		Name:     "Nasi Goreng",
		Category: "Food",
		Price:    25000,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/menu", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMenuItem(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var item model.MenuItem
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if item.Name != "Nasi Goreng" {
		t.Errorf("expected name 'Nasi Goreng', got '%s'", item.Name)
	}
	if item.ID.IsZero() {
		t.Error("expected non-zero ID")
	}
}

func TestCreateMenuItemMissingName(t *testing.T) {
	repo := newStubRepo()
	h := NewMenuHandler(repo)

	reqBody := CreateMenuItemRequest{
		Category: "Food",
		Price:    25000,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/menu", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMenuItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateMenuItemMissingCategory(t *testing.T) {
	repo := newStubRepo()
	h := NewMenuHandler(repo)

	reqBody := CreateMenuItemRequest{
		Name:  "Nasi Goreng",
		Price: 25000,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/menu", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMenuItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestListMenuItems(t *testing.T) {
	repo := newStubRepo()
	repo.items[bson.NewObjectID()] = &model.MenuItem{Name: "Item 1", Category: "Food", Price: 10000}
	repo.items[bson.NewObjectID()] = &model.MenuItem{Name: "Item 2", Category: "Drink", Price: 5000}

	h := NewMenuHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	w := httptest.NewRecorder()

	h.ListMenuItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var items []model.MenuItem
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestGetMenuItem(t *testing.T) {
	repo := newStubRepo()
	id := bson.NewObjectID()
	repo.items[id] = &model.MenuItem{Name: "Ayam Bakar", Category: "Food", Price: 35000}

	h := NewMenuHandler(repo)
	r := chi.NewRouter()
	r.Get("/menu/{id}", h.GetMenuItem)

	req := httptest.NewRequest(http.MethodGet, "/menu/"+id.Hex(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var item model.MenuItem
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if item.Name != "Ayam Bakar" {
		t.Errorf("expected name 'Ayam Bakar', got '%s'", item.Name)
	}
}

func TestGetMenuItemNotFound(t *testing.T) {
	repo := newStubRepo()
	h := NewMenuHandler(repo)

	r := chi.NewRouter()
	r.Get("/menu/{id}", h.GetMenuItem)

	fakeID := bson.NewObjectID()
	req := httptest.NewRequest(http.MethodGet, "/menu/"+fakeID.Hex(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestGetCategoryStats(t *testing.T) {
	repo := newStubRepo()
	h := NewMenuHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/menu/stats", nil)
	w := httptest.NewRecorder()

	h.GetCategoryStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var stats []model.CategoryStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("expected 1 stat entry, got %d", len(stats))
	}
}

func TestMenuItemJSONRoundTrip(t *testing.T) {
	item := model.MenuItem{
		Name:        "Ayam Bakar",
		Description: "Grilled chicken with sambal",
		Category:    "Food",
		Price:       35000,
		IsAvailable: true,
		Variants: []model.Variant{
			{Name: "Small", Price: 25000, SKU: "AYAM-S", IsDefault: true},
			{Name: "Large", Price: 45000, SKU: "AYAM-L"},
		},
		Modifiers: []model.Modifier{
			{Name: "Extra Sambal", Price: 5000},
		},
		BranchIDs: []string{"branch-001"},
		Tags:      []string{"spicy"},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded model.MenuItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != item.Name {
		t.Errorf("name mismatch after roundtrip")
	}
	if decoded.Price != item.Price {
		t.Errorf("price mismatch after roundtrip")
	}
	if len(decoded.Variants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(decoded.Variants))
	}
}
