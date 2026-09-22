package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// OrderUpdate represents a real-time order status update published via PubSub.
type OrderUpdate struct {
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	Status     string    `json:"status"`
	BranchID   string    `json:"branch_id"`
	TotalPrice int64     `json:"total_price"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PubSub wraps a Redis client for order-related pub/sub messaging.
type PubSub struct {
	client *goredis.Client
}

// NewPubSub creates a new PubSub instance.
func NewPubSub(client *goredis.Client) *PubSub {
	return &PubSub{client: client}
}

const orderUpdatesChannel = "warungos:order_updates"

// PublishOrderUpdate publishes an order status update to subscribers.
func (ps *PubSub) PublishOrderUpdate(ctx context.Context, update OrderUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("marshal order update: %w", err)
	}
	return ps.client.Publish(ctx, orderUpdatesChannel, data).Err()
}

// SubscribeOrderUpdates subscribes to order status updates and returns a channel.
// The caller must close the returned channel's context to stop receiving.
func (ps *PubSub) SubscribeOrderUpdates(ctx context.Context) <-chan OrderUpdate {
	ch := make(chan OrderUpdate, 64)
	sub := ps.client.Subscribe(ctx, orderUpdatesChannel)

	go func() {
		defer close(ch)
		defer sub.Close()

		msgCh := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				var update OrderUpdate
				if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
					continue
				}
				select {
				case ch <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}
