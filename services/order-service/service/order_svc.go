package service

import (
	"context"
	"fmt"
	"time"

	"github.com/warungos/order-service/model"
	"github.com/warungos/order-service/repository"
	sharedredis "github.com/warungos/shared/redis"
)

// OrderService handles order business logic.
type OrderService struct {
	repo   *repository.OrderRepository
	pubsub *sharedredis.PubSub
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo *repository.OrderRepository, pubsub *sharedredis.PubSub) *OrderService {
	return &OrderService{repo: repo, pubsub: pubsub}
}

// CreateOrder validates, calculates totals, persists, and publishes an event.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *model.CreateOrderRequest) (*model.Order, error) {
	order := &model.Order{
		ID:       generateUUID(),
		UserID:   userID,
		BranchID: req.BranchID,
		Notes:    req.Notes,
	}

	for _, item := range req.Items {
		order.Items = append(order.Items, model.OrderItem{
			ID:        generateUUID(),
			MenuID:    item.MenuID,
			MenuName:  item.MenuName,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	if err := order.ValidateItems(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	order.CalculateTotal()

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Publish creation event
	_ = s.pubsub.PublishOrderUpdate(ctx, sharedredis.OrderUpdate{
		OrderID:    order.ID,
		UserID:     order.UserID,
		Status:     string(order.Status),
		BranchID:   order.BranchID,
		TotalPrice: order.TotalPrice,
		UpdatedAt:  time.Now(),
	})

	return order, nil
}

// UpdateOrderStatus validates the transition, persists, and publishes.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, newStatus model.OrderStatus) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	if !order.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid transition from %s to %s", order.Status, newStatus)
	}

	if err := s.repo.UpdateStatus(ctx, orderID, newStatus); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Publish status change
	_ = s.pubsub.PublishOrderUpdate(ctx, sharedredis.OrderUpdate{
		OrderID:    orderID,
		UserID:     order.UserID,
		Status:     string(newStatus),
		BranchID:   order.BranchID,
		TotalPrice: order.TotalPrice,
		UpdatedAt:  time.Now(),
	})

	return nil
}

// GetOrder retrieves an order by ID.
func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

// ListOrders retrieves orders with filters and pagination.
func (s *OrderService) ListOrders(ctx context.Context, q model.OrderListQuery) ([]model.Order, int, error) {
	return s.repo.List(ctx, q)
}

// CountByBranchAndDate returns the daily order count for a branch.
func (s *OrderService) CountByBranchAndDate(ctx context.Context, branchID string, date time.Time) (int, error) {
	return s.repo.CountByBranchAndDate(ctx, branchID, date)
}

// generateUUID produces a simple UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	now := time.Now().UnixNano()
	for i := 0; i < 16; i++ {
		b[i] = byte((now >> (uint(i) * 4)) & 0xff)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
