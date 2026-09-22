package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/warungos/inventory-service/model"
)

type InventoryRepo struct {
	db *pgxpool.Pool
}

func NewInventoryRepo(db *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{db: db}
}

const selectCols = `id, branch_id, item_name, item_code, category, quantity, unit, min_quantity, max_quantity, unit_cost, is_active, updated_at`

func scanItem(row pgx.Row) (*model.InventoryItem, error) {
	item := &model.InventoryItem{}
	err := row.Scan(
		&item.ID, &item.BranchID, &item.ItemName, &item.ItemCode, &item.Category,
		&item.Quantity, &item.Unit, &item.MinQuantity, &item.MaxQuantity,
		&item.UnitCost, &item.IsActive, &item.UpdatedAt,
	)
	return item, err
}

func (r *InventoryRepo) Create(ctx context.Context, item *model.InventoryItem) error {
	query := `INSERT INTO inventory (branch_id, item_name, item_code, category, quantity, unit, min_quantity, max_quantity, unit_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING ` + selectCols
	return r.db.QueryRow(ctx, query,
		item.BranchID, item.ItemName, item.ItemCode, item.Category,
		item.Quantity, item.Unit, item.MinQuantity, item.MaxQuantity, item.UnitCost,
	).Scan(
		&item.ID, &item.BranchID, &item.ItemName, &item.ItemCode, &item.Category,
		&item.Quantity, &item.Unit, &item.MinQuantity, &item.MaxQuantity,
		&item.UnitCost, &item.IsActive, &item.UpdatedAt,
	)
}

func (r *InventoryRepo) FindByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	item, err := scanItem(r.db.QueryRow(ctx, `SELECT `+selectCols+` FROM inventory WHERE id = $1`, id))
	if err != nil {
		return nil, fmt.Errorf("find inventory by id: %w", err)
	}
	return item, nil
}

func (r *InventoryRepo) Update(ctx context.Context, id string, req *model.UpdateInventoryRequest) (*model.InventoryItem, error) {
	existing, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.ItemName != nil {
		existing.ItemName = *req.ItemName
	}
	if req.Quantity != nil {
		existing.Quantity = *req.Quantity
	}
	if req.Unit != nil {
		existing.Unit = *req.Unit
	}
	if req.MinQuantity != nil {
		existing.MinQuantity = *req.MinQuantity
	}
	if req.UnitCost != nil {
		existing.UnitCost = *req.UnitCost
	}
	query := `UPDATE inventory SET item_name=$2, quantity=$3, unit=$4, min_quantity=$5, unit_cost=$6, updated_at=NOW() WHERE id=$1 RETURNING ` + selectCols
	item, err := scanItem(r.db.QueryRow(ctx, query, id,
		existing.ItemName, existing.Quantity, existing.Unit, existing.MinQuantity, existing.UnitCost,
	))
	if err != nil {
		return nil, fmt.Errorf("update inventory: %w", err)
	}
	return item, nil
}

func (r *InventoryRepo) List(ctx context.Context, branchID string, page, limit int) ([]model.InventoryItem, int64, error) {
	offset := (page - 1) * limit
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM inventory WHERE branch_id=$1 AND is_active=true`, branchID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count inventory: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+selectCols+` FROM inventory WHERE branch_id=$1 AND is_active=true ORDER BY item_name LIMIT $2 OFFSET $3`, branchID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventory: %w", err)
	}
	defer rows.Close()
	var items []model.InventoryItem
	for rows.Next() {
		var item model.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.BranchID, &item.ItemName, &item.ItemCode, &item.Category,
			&item.Quantity, &item.Unit, &item.MinQuantity, &item.MaxQuantity,
			&item.UnitCost, &item.IsActive, &item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan inventory: %w", err)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *InventoryRepo) GetLowStockItems(ctx context.Context, branchID string) ([]model.InventoryItem, error) {
	rows, err := r.db.Query(ctx, `SELECT `+selectCols+` FROM inventory WHERE branch_id=$1 AND is_active=true AND quantity <= min_quantity ORDER BY quantity`, branchID)
	if err != nil {
		return nil, fmt.Errorf("get low stock: %w", err)
	}
	defer rows.Close()
	var items []model.InventoryItem
	for rows.Next() {
		var item model.InventoryItem
		if err := rows.Scan(
			&item.ID, &item.BranchID, &item.ItemName, &item.ItemCode, &item.Category,
			&item.Quantity, &item.Unit, &item.MinQuantity, &item.MaxQuantity,
			&item.UnitCost, &item.IsActive, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan low stock: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InventoryRepo) ReserveStock(ctx context.Context, itemID string, quantity float64) error {
	tag, err := r.db.Exec(ctx, `UPDATE inventory SET quantity = quantity - $2, updated_at = NOW() WHERE id = $1 AND quantity >= $2`, itemID, quantity)
	if err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for item %s", itemID)
	}
	return nil
}

func (r *InventoryRepo) FindByIDTx(ctx context.Context, tx pgx.Tx, id string) (*model.InventoryItem, error) {
	item, err := scanItem(tx.QueryRow(ctx, `SELECT `+selectCols+` FROM inventory WHERE id=$1 FOR UPDATE`, id))
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *InventoryRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}
