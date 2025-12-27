package payments

import (
	"context"
	"encoding/json"
)

// HandleWebhookPayload parses and dispatches webhook events.
func (s *Service) HandleWebhookPayload(ctx context.Context, payload []byte) error {
	var event NotificationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	return s.HandleWebhookNotification(ctx, event)
}
