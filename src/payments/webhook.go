package payments

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

// WebhookHandler processes incoming webhook notifications.
type WebhookHandler struct {
    repo *Repository
}

// NewWebhookHandler creates a new handler.
func NewWebhookHandler(repo *Repository) *WebhookHandler {
    return &WebhookHandler{repo: repo}
}

// Handle parses the webhook and persists the referenced entities using the repository only.
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "unable to read body", http.StatusBadRequest)
        return
    }

    var event WebhookEvent
    if err := json.Unmarshal(payload, &event); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    if err := h.persistEvent(ctx, event); err != nil {
        http.Error(w, fmt.Sprintf("unable to persist event: %v", err), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) persistEvent(ctx context.Context, event WebhookEvent) error {
    switch event.Event {
    case "PAYMENT_CREATED", "PAYMENT_CONFIRMED", "PAYMENT_OVERDUE", "PAYMENT_UPDATED":
        if event.Payment == nil {
            return fmt.Errorf("missing payment payload")
        }
        rec := PaymentRecord{
            ID:         event.Payment.ID,
            CustomerID: event.Payment.CustomerID,
            Status:     event.Payment.Status,
            Value:      event.Payment.Value,
        }
        return h.repo.SavePayment(ctx, rec)
    case "CUSTOMER_CREATED", "CUSTOMER_UPDATED":
        if event.Customer == nil {
            return fmt.Errorf("missing customer payload")
        }
        rec := CustomerRecord{ID: event.Customer.ID, Name: event.Customer.Name}
        return h.repo.SaveCustomer(ctx, rec)
    case "INVOICE_CREATED", "INVOICE_UPDATED", "INVOICE_DELETED":
        if event.Invoice == nil {
            return fmt.Errorf("missing invoice payload")
        }
        rec := InvoiceRecord{
            ID:         event.Invoice.ID,
            CustomerID: event.Invoice.CustomerID,
            Status:     event.Invoice.Status,
            Value:      event.Invoice.Value,
        }
        return h.repo.SaveInvoice(ctx, rec)
    default:
        return nil
    }
}
