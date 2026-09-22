package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/warungos/inventory-service/model"
)

// InventoryRepo handles PostgreSQL operations for inventory
type InventoryRepo struct {
	db *pgxpool.Pool
}

// NewInventoryRepo creates a new inventory repository
func NewInventoryRepo(db *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{db: db}
}

// Create inserts a new inventory item
func (r *InventoryRepo) Create(ctx context.Context, item *model.InventoryItem) error {
	query := `
		INSERT INTO inventory (id, branch_id, item_name, quantity, unit, min_stock, cost_per_unit, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, updated_at`
	return r.db.QueryRow(ctx, query,
		item.BranchID, item.ItemName, item.Quantity,
		item.Unit, item.MinStock, item.CostPerUnit,
	).Scan(&item.ID, &item.UpdatedAt)
}

// FindByID retrieves an inventory item by ID
func (r *InventoryRepo) FindByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	query := `
		SELECT id, branch_id, item_name, quantity, unit, min_stock, cost_per_unit, updated_at
		FROM inventory WHERE id = $1`
	item := &model.InventoryItem{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.BranchID, &item.ItemName, &item.Quantity,
		&item.Unit, &item.MinStock, &item.CostPerUnit, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find inventory by id: %w", err)
	}
	return item, nil
}

// Update modifies an existing inventory item
func (r *InventoryRepo) Update(ctx context.Context, id string, item *model.UpdateInventoryRequest) (*model.InventoryItem, error) {
	// Fetch existing first
	existing, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply partial updates
	if item.ItemName != nil {
		existing.ItemName = *item.ItemName
	}
	if item.Quantity != nil {
		existing.Quantity = *item.Quantity
	}
	if item.Unit != nil {
		existing.Unit = *item.Unit
	}
	if item.MinStock != nil {
		existing.MinStock = *item.MinStock
	}
	if item.CostPerUnit != nil {
		existing.CostPerUnit = *item.CostPerUnit
	}

	query := `
		UPDATE inventory
		SET item_name = $2, quantity = $3, unit = $4, min_stock = $5, cost_per_unit = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	err = r.db.QueryRow(ctx, query, id,
		existing.ItemName, existing.Quantity, existing.Unit,
		existing.MinStock, existing.CostPerUnit,
	).Scan(&existing.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update inventory: %w", err)
	}
	return existing, nil
}

// List retrieves inventory items for a branch with pagination
func (r *InventoryRepo) List(ctx context.Context, branchID string, page, limit int) ([]model.InventoryItem, int64, error) {
	offset := (page - 1) * limit

	// Count total
	var total int64
	countQuery := "SELECT COUNT(*) FROM inventory WHERE branch_id = $1"
	err := r.db.QueryRow(ctx, countQuery, branchID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count inventory: %w", err)
	}

	// Fetch items
	query := `
		SELECT id, branch_id, item_name, quantity, unit, min_stock, cost_per_unit, updated_at
		FROM inventory
		WHERE branch_id = $1
		ORDER BY item_name ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventory: %w", err)
	}
	defer rows.Close()

	var items []model.InventoryItem
	for rows.Next() {
		var item model.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.BranchID, &item.ItemName, &item.Quantity,
			&item.Unit, &item.MinStock, &item.CostPerUnit, &item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan inventory: %w", err)
		}
		items = append(items, item)
	}
	return items, total, nil
}

// GetLowStockItems finds items where quantity <= min_stock
func (r *InventoryRepo) GetLowStockItems(ctx context.Context, branchID string) ([]model.InventoryItem, error) {
	query := `
		SELECT id, branch_id, item_name, quantity, unit, min_stock, cost_per_unit, updated_at
		FROM inventory
		WHERE branch_id = $1 AND quantity <= min_stock
		ORDER BY (quantity / NULLIF(min_stock, 0)) ASC`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, fmt.Errorf("get low stock: %w", err)
	}
	defer rows.Close()

	var items []model.InventoryItem
	for rows.Next() {
		var item model.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.BranchID, &item.ItemName, &item.Quantity,
			&item.Unit, &item.MinStock, &item.CostPerUnit, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan low stock: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// ReserveStock atomically decrements stock with optimistic locking
// This prevents overselling by checking quantity >= requested amount
func (r *InventoryRepo) ReserveStock(ctx context.Context, itemID string, quantity float64) error {
	query := `
		UPDATE inventory
		SET quantity = quantity - $2, updated_at = NOW()
		WHERE id = $1 AND quantity >= $2`

	tag, err := r.db.Exec(ctx, query, itemID, quantity)
	if err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for item %s", itemID)
	}
	return nil
}

// FindByIDTx retrieves an inventory item within a transaction (for atomic operations)
func (r *InventoryRepo) FindByIDTx(ctx context.Context, tx pgx.Tx, id string) (*model.InventoryItem, error) {
	query := `
		SELECT id, branch_id, item_name, quantity, unit, min_stock, cost_per_unit, updated_at
		FROM inventory WHERE id = $1 FOR UPDATE`
	item := &model.InventoryItem{}
	err := tx.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.BranchID, &item.ItemName, &item.Quantity,
		&item.Unit, &item.MinStock, &item.CostPerUnit, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateQuantityTx updates quantity within a transaction
func (r *InventoryRepo) UpdateQuantityTx(ctx context.Context, tx pgx.Tx, id string, quantity float64) error {
	query := `UPDATE inventory SET quantity = $2, updated_at = NOW() WHERE id = $1`
	_, err := tx.Exec(ctx, query, id, quantity)
	return err
}

// BeginTx starts a new transaction
func (r *InventoryRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// Ensure we use time import
var _ = time.Now
