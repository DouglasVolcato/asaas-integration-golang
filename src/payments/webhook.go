package payments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// WebhookHandler processes Asaas webhook notifications and persists status updates.
func (c *Client) WebhookHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read webhook", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	switch event.Event {
	case "PAYMENT_RECEIVED", "PAYMENT_OVERDUE", "PAYMENT_CONFIRMED":
		if event.Payment != nil {
			_ = c.UpdatePaymentStatus(ctx, event.Payment.ID, event.Payment.Status)
		}
	case "SUBSCRIPTION_CREATED", "SUBSCRIPTION_UPDATED":
		if event.Subscription != nil {
			_ = c.UpdateSubscriptionStatus(ctx, event.Subscription.ID, event.Subscription.Status)
		}
	case "INVOICE_CREATED", "INVOICE_SETTLED", "INVOICE_CANCELED":
		if event.Invoice != nil {
			_ = c.UpdateInvoiceStatus(ctx, event.Invoice.ID, event.Invoice.Status)
		}
	}

	w.WriteHeader(http.StatusOK)
}
