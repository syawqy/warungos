package model

// InventoryItem represents a stock item in a branch
type InventoryItem struct {
	ID          string  `json:"id"`
	BranchID    string  `json:"branch_id"`
	ItemName    string  `json:"item_name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinStock    float64 `json:"min_stock"`
	CostPerUnit float64 `json:"cost_per_unit"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateInventoryRequest is the payload for creating inventory items
type CreateInventoryRequest struct {
	BranchID    string  `json:"branch_id"`
	ItemName    string  `json:"item_name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinStock    float64 `json:"min_stock"`
	CostPerUnit float64 `json:"cost_per_unit"`
}

// UpdateInventoryRequest is the payload for updating inventory items
type UpdateInventoryRequest struct {
	ItemName    *string  `json:"item_name,omitempty"`
	Quantity    *float64 `json:"quantity,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	MinStock    *float64 `json:"min_stock,omitempty"`
	CostPerUnit *float64 `json:"cost_per_unit,omitempty"`
}

// ReserveRequest is the payload for reserving stock
type ReserveRequest struct {
	ItemID   string  `json:"item_id"`
	Quantity float64 `json:"quantity"`
}
