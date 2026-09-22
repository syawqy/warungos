package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/warungos/order-service/model"
)

// OrderRepository handles order persistence.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create inserts an order and its items in a single transaction.
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	if order.Status == "" {
		order.Status = model.StatusPending
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, user_id, branch_id, status, subtotal, tax, total_price, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		order.ID, order.UserID, order.BranchID, order.Status,
		order.Subtotal, order.Tax, order.TotalPrice, order.Notes,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, menu_id, menu_name, quantity, unit_price, total_price)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].MenuID,
			order.Items[i].MenuName, order.Items[i].Quantity,
			order.Items[i].UnitPrice, order.Items[i].TotalPrice,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// FindByID retrieves an order with its items using a JOIN (not N+1).
func (r *OrderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order

	// Single query with JOIN to fetch order + items
	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.user_id, o.branch_id, o.status, o.subtotal, o.tax, o.total_price,
		        o.notes, o.created_at, o.updated_at,
		        oi.id, oi.order_id, oi.menu_id, oi.menu_name, oi.quantity, oi.unit_price, oi.total_price
		 FROM orders o
		 LEFT JOIN order_items oi ON oi.order_id = o.id
		 WHERE o.id = $1
		 ORDER BY oi.id`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&order.ID, &order.UserID, &order.BranchID, &order.Status,
			&order.Subtotal, &order.Tax, &order.TotalPrice,
			&order.Notes, &order.CreatedAt, &order.UpdatedAt,
			&item.ID, &item.OrderID, &item.MenuID, &item.MenuName,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice,
		); err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}
		if first {
			first = false
		}
		order.Items = append(order.Items, item)
	}

	if first {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return &order, nil
}

// List retrieves orders with optional filtering and pagination.
func (r *OrderRepository) List(ctx context.Context, q model.OrderListQuery) ([]model.Order, int, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if q.BranchID != "" {
		where = append(where, fmt.Sprintf("o.branch_id = $%d", argIdx))
		args = append(args, q.BranchID)
		argIdx++
	}
	if q.UserID != "" {
		where = append(where, fmt.Sprintf("o.user_id = $%d", argIdx))
		args = append(args, q.UserID)
		argIdx++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", argIdx))
		args = append(args, q.Status)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders o %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	// Fetch page
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataArgs := make([]interface{}, len(args))
	copy(dataArgs, args)
	dataArgs = append(dataArgs, pageSize, offset)

	dataQuery := fmt.Sprintf(
		`SELECT o.id, o.user_id, o.branch_id, o.status, o.subtotal, o.tax, o.total_price,
		        o.notes, o.created_at, o.updated_at
		 FROM orders o %s
		 ORDER BY o.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.BranchID, &o.Status,
			&o.Subtotal, &o.Tax, &o.TotalPrice,
			&o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}

	// For each order, fetch its items
	for i := range orders {
		itemRows, err := r.pool.Query(ctx,
			`SELECT id, order_id, menu_id, menu_name, quantity, unit_price, total_price
			 FROM order_items WHERE order_id = $1`, orders[i].ID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("query order items: %w", err)
		}
		for itemRows.Next() {
			var item model.OrderItem
			if err := itemRows.Scan(
				&item.ID, &item.OrderID, &item.MenuID, &item.MenuName,
				&item.Quantity, &item.UnitPrice, &item.TotalPrice,
			); err != nil {
				itemRows.Close()
				return nil, 0, fmt.Errorf("scan order item: %w", err)
			}
			orders[i].Items = append(orders[i].Items, item)
		}
		itemRows.Close()
	}

	return orders, total, nil
}

// UpdateStatus changes the status of an order.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status, now, id,
	)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %s", id)
	}
	return nil
}

// CountByBranchAndDate returns the number of orders for a branch on a specific date.
func (r *OrderRepository) CountByBranchAndDate(ctx context.Context, branchID string, date time.Time) (int, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE branch_id = $1 AND created_at >= $2 AND created_at < $3`,
		branchID, start, end,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count orders by branch and date: %w", err)
	}
	return count, nil
}

// GetOrderByID is an alias for FindByID used by the service layer.
func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return r.FindByID(ctx, id)
}

// GetOrderByIDForUpdate retrieves an order for update within a transaction.
func (r *OrderRepository) GetOrderByIDForUpdate(ctx context.Context, tx pgx.Tx, id string) (*model.Order, error) {
	var order model.Order
	err := tx.QueryRow(ctx,
		`SELECT id, user_id, branch_id, status, subtotal, tax, total_price, notes, created_at, updated_at
		 FROM orders WHERE id = $1 FOR UPDATE`, id,
	).Scan(
		&order.ID, &order.UserID, &order.BranchID, &order.Status,
		&order.Subtotal, &order.Tax, &order.TotalPrice,
		&order.Notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get order for update: %w", err)
	}
	return &order, nil
}
