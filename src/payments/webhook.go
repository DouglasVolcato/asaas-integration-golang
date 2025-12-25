package payments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// WebhookHandler exposes HTTP handling for Asaas webhooks.
type WebhookHandler struct {
	service *Service
}

// NewWebhookHandler returns a handler wired to the service layer.
func NewWebhookHandler(service *Service) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// ServeHTTP implements http.Handler and dispatches webhook events.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event NotificationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.service.HandleWebhookNotification(ctx, event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WithContext allows explicit context usage in other environments.
func (h *WebhookHandler) WithContext(ctx context.Context, payload []byte) error {
	var event NotificationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	return h.service.HandleWebhookNotification(ctx, event)
}
