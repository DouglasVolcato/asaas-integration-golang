package payments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Service orchestrates local persistence and remote Asaas calls.
type Service struct {
	repo   Repository
	client *AsaasClient
}

// NewService creates a payment service.
func NewService(repo Repository, client *AsaasClient) *Service {
	return &Service{repo: repo, client: client}
}

// RegisterCustomer stores a local customer and creates it in Asaas.
func (s *Service) RegisterCustomer(ctx context.Context, req CustomerRequest) (CustomerRecord, CustomerResponse, error) {
	remote, err := s.client.CreateCustomer(ctx, req)
	if err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to create asaas customer: %w", err)
	}

	now := time.Now().UTC()
	local := CustomerRecord{
		ID:         remote.ID,
		ExternalID: remote.ExternalID,
		Name:       remote.Name,
		Email:      remote.Email,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.SaveCustomer(ctx, local); err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to save local customer: %w", err)
	}

	return local, remote, nil
}

// CreatePayment persists the payment locally and in Asaas.
func (s *Service) CreatePayment(ctx context.Context, req PaymentRequest) (PaymentRecord, PaymentResponse, error) {
	remote, err := s.client.CreatePayment(ctx, req)
	if err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to create asaas payment: %w", err)
	}

	now := time.Now().UTC()
	local := PaymentRecord{
		ID:                    remote.ID,
		ExternalID:            remote.ExternalID,
		CustomerID:            req.Customer,
		BillingType:           remote.BillingType,
		Value:                 remote.Value,
		DueDate:               parseDate(req.DueDate),
		Status:                remote.Status,
		InvoiceURL:            remote.InvoiceURL,
		TransactionReceiptURL: remote.TransactionReceiptURL,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.SavePayment(ctx, local); err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to save local payment: %w", err)
	}

	return local, remote, nil
}

// CreateSubscription persists the subscription locally and remotely.
func (s *Service) CreateSubscription(ctx context.Context, req SubscriptionRequest) (SubscriptionRecord, SubscriptionResponse, error) {
	remote, err := s.client.CreateSubscription(ctx, req)
	if err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to create asaas subscription: %w", err)
	}

	now := time.Now().UTC()
	local := SubscriptionRecord{
		ID:          remote.ID,
		ExternalID:  remote.ExternalID,
		CustomerID:  req.Customer,
		Status:      remote.Status,
		Value:       remote.Value,
		Cycle:       req.Cycle,
		NextDueDate: parseDate(req.NextDueDate),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.SaveSubscription(ctx, local); err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to save local subscription: %w", err)
	}

	return local, remote, nil
}

// CreateInvoice persists the invoice locally and in Asaas.
func (s *Service) CreateInvoice(ctx context.Context, req InvoiceRequest) (InvoiceRecord, InvoiceResponse, error) {
	remote, err := s.client.CreateInvoice(ctx, req)
	if err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to create asaas invoice: %w", err)
	}

	now := time.Now().UTC()
	local := InvoiceRecord{
		ID:            remote.ID,
		ExternalID:    remote.ExternalID,
		CustomerID:    remote.Customer,
		Status:        remote.Status,
		Value:         remote.Value,
		EffectiveDate: parseDate(req.EffectiveDate),
		PaymentLink:   remote.PaymentLink,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.SaveInvoice(ctx, local); err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to save local invoice: %w", err)
	}

	return local, remote, nil
}

// HandleWebhookNotification updates local records based on webhook events.
func (s *Service) HandleWebhookNotification(ctx context.Context, event NotificationEvent) error {
	switch event.Event {
	case "PAYMENT_CREATED", "PAYMENT_RECEIVED", "PAYMENT_CONFIRMED", "PAYMENT_OVERDUE":
		if event.Payment == nil {
			return fmt.Errorf("payment payload missing")
		}
		return s.repo.UpdatePaymentStatus(ctx, event.Payment.ID, event.Payment.Status, event.Payment.InvoiceURL, event.Payment.TransactionReceiptURL)
	case "SUBSCRIPTION_CREATED", "SUBSCRIPTION_UPDATED":
		if event.Subscription == nil {
			return fmt.Errorf("subscription payload missing")
		}
		return s.repo.UpdateSubscriptionStatus(ctx, event.Subscription.ID, event.Subscription.Status)
	case "INVOICE_CREATED", "INVOICE_UPDATED", "INVOICE_OVERDUE":
		if event.Invoice == nil {
			return fmt.Errorf("invoice payload missing")
		}
		return s.repo.UpdateInvoiceStatus(ctx, event.Invoice.ID, event.Invoice.Status)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Event)
	}
}

func parseDate(value string) time.Time {
	// Asaas uses yyyy-mm-dd format; parsing errors return zero time for caller validation.
	t, _ := time.Parse("2006-01-02", value)
	return t
}

// ParseDateForTests exposes parseDate for integration tests without changing production API.
func ParseDateForTests(value string) time.Time {
	return parseDate(value)
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
