package model

import (
	"fmt"
	"time"
)

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusPreparing OrderStatus = "preparing"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

// validTransitions defines which status changes are allowed.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPreparing, StatusCancelled},
	StatusPreparing: {StatusReady, StatusCancelled},
	StatusReady:     {StatusCompleted, StatusCancelled},
	StatusCompleted: {},
	StatusCancelled: {},
}

// CanTransitionTo checks whether a transition from current to next is valid.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

// Order represents a customer order.
type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	BranchID    string      `json:"branch_id"`
	Status      OrderStatus `json:"status"`
	Items       []OrderItem `json:"items"`
	Subtotal    int64       `json:"subtotal"`
	Tax         int64       `json:"tax"`
	TotalPrice  int64       `json:"total_price"`
	Notes       string      `json:"notes"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem represents a single line item in an order.
type OrderItem struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	MenuID     string  `json:"menu_id"`
	MenuName   string  `json:"menu_name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  int64   `json:"unit_price"`
	TotalPrice int64   `json:"total_price"`
}

// CreateOrderRequest is the payload for creating a new order.
type CreateOrderRequest struct {
	BranchID string                 `json:"branch_id"`
	Notes    string                 `json:"notes"`
	Items    []CreateOrderItemReq   `json:"items"`
}

// CreateOrderItemReq is a single item in the create order request.
type CreateOrderItemReq struct {
	MenuID    string `json:"menu_id"`
	MenuName  string `json:"menu_name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

// CalculateTotal computes subtotal, 10% PPN tax, and grand total.
func (o *Order) CalculateTotal() {
	var subtotal int64
	for i := range o.Items {
		o.Items[i].TotalPrice = o.Items[i].UnitPrice * int64(o.Items[i].Quantity)
		subtotal += o.Items[i].TotalPrice
	}
	o.Subtotal = subtotal
	o.Tax = subtotal / 10 // 10% PPN
	o.TotalPrice = subtotal + o.Tax
}

// ValidateItems ensures the order has at least one item with positive quantity.
func (o *Order) ValidateItems() error {
	if len(o.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	for _, item := range o.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("item %s has invalid quantity: %d", item.MenuName, item.Quantity)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("item %s has negative unit price", item.MenuName)
		}
	}
	return nil
}

// UpdateStatusRequest is the payload for PATCH /orders/{id}/status.
type UpdateStatusRequest struct {
	Status OrderStatus `json:"status"`
}

// OrderListQuery holds query parameters for listing orders.
type OrderListQuery struct {
	BranchID string
	UserID   string
	Status   string
	Page     int
	PageSize int
}
