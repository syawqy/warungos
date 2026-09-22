package model

import "time"
// InventoryItem represents a stock item in a branch
type InventoryItem struct {
	ID          string  `json:"id"`
	BranchID    string  `json:"branch_id"`
	ItemName    string  `json:"item_name"`
	ItemCode    string  `json:"item_code"`
	Category    string  `json:"category"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinQuantity float64 `json:"min_quantity"`
	MaxQuantity float64 `json:"max_quantity"`
	UnitCost    int64   `json:"unit_cost"`
	IsActive    bool    `json:"is_active"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateInventoryRequest is the payload for creating inventory items
type CreateInventoryRequest struct {
	BranchID    string  `json:"branch_id"`
	ItemName    string  `json:"item_name"`
	ItemCode    string  `json:"item_code"`
	Category    string  `json:"category"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinQuantity float64 `json:"min_quantity"`
	MaxQuantity float64 `json:"max_quantity"`
	UnitCost    int64   `json:"unit_cost"`
}

// UpdateInventoryRequest is the payload for updating inventory items
type UpdateInventoryRequest struct {
	ItemName    *string  `json:"item_name,omitempty"`
	Quantity    *float64 `json:"quantity,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	MinQuantity *float64 `json:"min_quantity,omitempty"`
	UnitCost    *int64   `json:"unit_cost,omitempty"`
}

// ReserveRequest is the payload for reserving stock
type ReserveRequest struct {
	ItemID   string  `json:"item_id"`
	Quantity float64 `json:"quantity"`
}
