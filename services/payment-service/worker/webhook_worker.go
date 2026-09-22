package worker

import (
	"context"
	"log"
	"time"

	"warungos/payment-service/integration"
	"warungos/payment-service/repository"
)

// WebhookWorker polls pending payments and checks their status with Midtrans.
type WebhookWorker struct {
	repo          repository.PaymentRepository
	midtransClient *integration.MidtransClient
	pollInterval  time.Duration
}

// NewWebhookWorker creates a new webhook polling worker.
func NewWebhookWorker(repo repository.PaymentRepository, client *integration.MidtransClient) *WebhookWorker {
	return &WebhookWorker{
		repo:          repo,
		midtransClient: client,
		pollInterval:  10 * time.Second,
	}
}

// Start begins the polling loop. It blocks until the context is cancelled.
func (w *WebhookWorker) Start(ctx context.Context) {
	log.Println("webhook worker: starting pending payment poller")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Run immediately on start
	w.pollPendingPayments(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("webhook worker: shutting down")
			return
		case <-ticker.C:
			w.pollPendingPayments(ctx)
		}
	}
}

// pollPendingPayments checks all pending payments against Midtrans and updates their status.
func (w *WebhookWorker) pollPendingPayments(ctx context.Context) {
	payments, err := w.repo.GetPendingPayments(ctx)
	if err != nil {
		log.Printf("webhook worker: failed to get pending payments: %v", err)
		return
	}

	if len(payments) == 0 {
		return
	}

	log.Printf("webhook worker: checking %d pending payments", len(payments))

	for _, p := range payments {
		if ctx.Err() != nil {
			return
		}

		externalID := p.ExternalID
		if externalID == "" {
			externalID = p.OrderID
		}

		status, err := w.midtransClient.GetStatus(ctx, externalID)
		if err != nil {
			log.Printf("webhook worker: failed to check status for payment %s: %v", p.ID, err)
			continue
		}

		newStatus := integration.MapTransactionStatus(status.TransactionStatus)
		if newStatus != p.Status {
			if err := w.repo.UpdateStatus(ctx, p.ID, newStatus); err != nil {
				log.Printf("webhook worker: failed to update status for payment %s: %v", p.ID, err)
				continue
			}
			log.Printf("webhook worker: payment %s updated from %s to %s", p.ID, p.Status, newStatus)
		}
	}
}
