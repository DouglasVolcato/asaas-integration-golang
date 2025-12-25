package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WebhookHandler processes Asaas notification callbacks and delegates status updates to the service.
type WebhookHandler struct {
	service *Service
	secret  string
}

// NewWebhookHandler builds a new webhook handler with optional secret validation.
func NewWebhookHandler(service *Service, secret string) *WebhookHandler {
	return &WebhookHandler{service: service, secret: secret}
}

// ServeHTTP reads the webhook payload and updates local records.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		signature := r.Header.Get("X-Asaas-Signature")
		if signature != h.secret {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.handleEvent(ctx, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleEvent(ctx context.Context, event WebhookEvent) error {
	switch event.Event {
	case "PAYMENT_CONFIRMED", "PAYMENT_RECEIVED", "PAYMENT_OVERDUE", "PAYMENT_PENDING":
		if event.Payment == nil {
			return fmt.Errorf("payment payload is required for payment events")
		}
		return h.service.HandlePaymentStatus(ctx, event.Payment.ID, event.Payment.Status)
	case "INVOICE_AUTHORIZED", "INVOICE_CREATED", "INVOICE_CANCELED":
		if event.Invoice == nil {
			return fmt.Errorf("invoice payload is required for invoice events")
		}
		return h.service.HandleInvoiceStatus(ctx, event.Invoice.ID, event.Invoice.Status)
	default:
		return nil
	}
}
