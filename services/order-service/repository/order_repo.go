package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/warungos/order-service/model"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d%04d", time.Now().Unix()%100000, rand.Intn(10000))
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.OrderNumber = generateOrderNumber()
	if order.Status == "" {
		order.Status = model.StatusPending
	}
	order.PaymentStatus = "unpaid"

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, order_number, user_id, branch_id, status, order_type, customer_name, subtotal, tax_amount, total_price, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		order.ID, order.OrderNumber, order.UserID, order.BranchID, order.Status,
		order.OrderType, order.CustomerName, order.Subtotal, order.TaxAmount, order.TotalPrice, order.Notes,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if order.Items[i].ID == "" {
			order.Items[i].ID = fmt.Sprintf("oi-%s-%d", order.ID[:8], i)
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, menu_item_id, menu_item_name, quantity, unit_price, total_price, special_instructions)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].MenuItemID,
			order.Items[i].MenuItemName, order.Items[i].Quantity,
			order.Items[i].UnitPrice, order.Items[i].TotalPrice, order.Items[i].Notes,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.order_number, o.user_id, o.branch_id, o.status, o.order_type, o.customer_name, o.subtotal, o.tax_amount, o.total_price,
		        o.notes, o.created_at, o.updated_at,
		        oi.id, oi.order_id, oi.menu_item_id, oi.menu_item_name, oi.quantity, oi.unit_price, oi.total_price
		 FROM orders o
		 LEFT JOIN order_items oi ON oi.order_id = o.id
		 WHERE o.id = $1 ORDER BY oi.id`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		var item model.OrderItem
		if first {
			if err := rows.Scan(
				&order.ID, &order.OrderNumber, &order.UserID, &order.BranchID, &order.Status,
				&order.OrderType, &order.CustomerName, &order.Subtotal, &order.TaxAmount, &order.TotalPrice,
				&order.Notes, &order.CreatedAt, &order.UpdatedAt,
				&item.ID, &item.OrderID, &item.MenuItemID, &item.MenuItemName,
				&item.Quantity, &item.UnitPrice, &item.TotalPrice,
			); err != nil {
				return nil, fmt.Errorf("scan order: %w", err)
			}
			first = false
		} else {
			if err := rows.Scan(
				&order.ID, &order.OrderNumber, &order.UserID, &order.BranchID, &order.Status,
				&order.OrderType, &order.CustomerName, &order.Subtotal, &order.TaxAmount, &order.TotalPrice,
				&order.Notes, &order.CreatedAt, &order.UpdatedAt,
				&item.ID, &item.OrderID, &item.MenuItemID, &item.MenuItemName,
				&item.Quantity, &item.UnitPrice, &item.TotalPrice,
			); err != nil {
				return nil, fmt.Errorf("scan item: %w", err)
			}
		}
		order.Items = append(order.Items, item)
	}
	if first {
		return nil, fmt.Errorf("order not found")
	}
	return &order, nil
}

func (r *OrderRepository) List(ctx context.Context, branchID, status string, page, limit int) ([]model.Order, int64, error) {
	offset := (page - 1) * limit

	countQ := `SELECT COUNT(*) FROM orders WHERE branch_id = $1`
	args := []interface{}{branchID}
	if status != "" {
		countQ += ` AND status = $2`
		args = append(args, status)
	}
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	q := `SELECT id, order_number, user_id, branch_id, status, order_type, customer_name, subtotal, tax_amount, total_price, notes, created_at, updated_at
	      FROM orders WHERE branch_id = $1`
	listArgs := []interface{}{branchID}
	argN := 2
	if status != "" {
		q += fmt.Sprintf(` AND status = $%d`, argN)
		listArgs = append(listArgs, status)
		argN++
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argN, argN+1)
	listArgs = append(listArgs, limit, offset)

	rows, err := r.pool.Query(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.UserID, &o.BranchID, &o.Status,
			&o.OrderType, &o.CustomerName, &o.Subtotal, &o.TaxAmount, &o.TotalPrice, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	tag, err := r.pool.Exec(ctx, `UPDATE orders SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

// Ensure pgx.Tx is used
var _ = pgx.Tx(nil)
