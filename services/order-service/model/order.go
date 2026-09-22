package model

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusPreparing OrderStatus = "preparing"
	StatusReady     OrderStatus = "ready"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
)

var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPreparing, StatusCancelled},
	StatusPreparing: {StatusReady, StatusCancelled},
	StatusReady:     {StatusCompleted},
	StatusCompleted: {},
	StatusCancelled: {},
}

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

type Order struct {
	ID           string      `json:"id"`
	OrderNumber  string      `json:"order_number"`
	UserID       string      `json:"user_id"`
	BranchID     string      `json:"branch_id"`
	Status       OrderStatus `json:"status"`
	OrderType    string      `json:"order_type"`
	CustomerName string      `json:"customer_name"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	PaymentStatus string    `json:"payment_status"`
	Items        []OrderItem `json:"items"`
	Subtotal     int64       `json:"subtotal"`
	TaxAmount    int64       `json:"tax_amount"`
	TotalPrice   int64       `json:"total_price"`
	Notes        string      `json:"notes"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID           string  `json:"id"`
	OrderID      string  `json:"order_id"`
	MenuItemID   string  `json:"menu_item_id"`
	MenuItemName string  `json:"menu_item_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    int64   `json:"unit_price"`
	TotalPrice   int64   `json:"total_price"`
	Notes        string  `json:"special_instructions,omitempty"`
}

type CreateOrderRequest struct {
	BranchID     string              `json:"branch_id"`
	CustomerName string              `json:"customer_name,omitempty"`
	OrderType    string              `json:"order_type,omitempty"`
	Items        []CreateOrderItemReq `json:"items"`
	Discount     int64               `json:"discount"`
	Notes        string              `json:"notes"`
}

type CreateOrderItemReq struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	Notes      string `json:"notes,omitempty"`
}

func (o *Order) CalculateTotal() {
	var subtotal int64
	for i := range o.Items {
		o.Items[i].TotalPrice = o.Items[i].UnitPrice * int64(o.Items[i].Quantity)
		subtotal += o.Items[i].TotalPrice
	}
	o.Subtotal = subtotal
	o.TaxAmount = subtotal / 10
	o.TotalPrice = subtotal + o.TaxAmount
}

func (o *Order) ValidateItems() error {
	if len(o.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	for _, item := range o.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity for item %s", item.MenuItemName)
		}
	}
	return nil
}

type UpdateStatusRequest struct {
	Status OrderStatus `json:"status"`
}

type OrderListQuery struct {
	BranchID string
	UserID   string
	Status   string
	Page     int
	PageSize int
}
