package service

import (
	"context"
	"fmt"
	"time"

	"github.com/warungos/order-service/model"
	"github.com/warungos/order-service/repository"
	sharedredis "github.com/warungos/shared/redis"
)

type OrderService struct {
	repo   *repository.OrderRepository
	pubsub *sharedredis.PubSub
}

func NewOrderService(repo *repository.OrderRepository, pubsub *sharedredis.PubSub) *OrderService {
	return &OrderService{repo: repo, pubsub: pubsub}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *model.CreateOrderRequest) (*model.Order, error) {
	order := &model.Order{
		ID:           generateUUID(),
		UserID:       userID,
		BranchID:     req.BranchID,
		OrderType:    req.OrderType,
		CustomerName: req.CustomerName,
		Notes:        req.Notes,
	}

	for _, item := range req.Items {
		order.Items = append(order.Items, model.OrderItem{
			ID:           generateUUID(),
			MenuItemID:   item.MenuItemID,
			MenuItemName: item.MenuItemID, // placeholder — real app would look up name
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			Notes:        item.Notes,
		})
	}

	if err := order.ValidateItems(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	order.CalculateTotal()

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	_ = s.pubsub.PublishOrderUpdate(ctx, sharedredis.OrderUpdate{
		OrderID:    order.ID,
		Status:     string(order.Status),
		BranchID:   order.BranchID,
		TotalPrice: order.TotalPrice,
		UpdatedAt:  time.Now(),
	})

	return order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, newStatus model.OrderStatus) (*model.Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	if !order.Status.CanTransitionTo(newStatus) {
		return nil, fmt.Errorf("invalid transition from %s to %s", order.Status, newStatus)
	}

	if err := s.repo.UpdateStatus(ctx, orderID, newStatus); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	order.Status = newStatus
	_ = s.pubsub.PublishOrderUpdate(ctx, sharedredis.OrderUpdate{
		OrderID:    orderID,
		Status:     string(newStatus),
		BranchID:   order.BranchID,
		TotalPrice: order.TotalPrice,
		UpdatedAt:  time.Now(),
	})

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context, branchID, status string, page, limit int) ([]model.Order, int64, error) {
	return s.repo.List(ctx, branchID, status, page, limit)
}

func generateUUID() string {
	b := make([]byte, 16)
	now := time.Now().UnixNano()
	for i := 0; i < 16; i++ {
		b[i] = byte((now >> (uint(i) * 4)) & 0xff)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
