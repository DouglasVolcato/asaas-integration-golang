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
	now := time.Now().UTC()
	local := CustomerRecord{
		ID:        generateID(),
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.SaveCustomer(ctx, local); err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to save local customer: %w", err)
	}

	remote, err := s.client.CreateCustomer(ctx, req)
	if err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to create asaas customer: %w", err)
	}
	if err := s.repo.UpdateCustomerExternalID(ctx, local.ID, remote.ID); err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to update customer external id: %w", err)
	}
	local.ExternalID = remote.ID

	return local, remote, nil
}

// CreatePayment persists the payment locally and in Asaas.
func (s *Service) CreatePayment(ctx context.Context, req PaymentRequest, customerID string) (PaymentRecord, PaymentResponse, error) {
	now := time.Now().UTC()
	local := PaymentRecord{
		ID:          generateID(),
		CustomerID:  customerID,
		BillingType: req.BillingType,
		Value:       req.Value,
		DueDate:     parseDate(req.DueDate),
		Status:      "PENDING",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.SavePayment(ctx, local); err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to save local payment: %w", err)
	}

	remote, err := s.client.CreatePayment(ctx, req)
	if err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to create asaas payment: %w", err)
	}

	if err := s.repo.UpdatePaymentExternalID(ctx, local.ID, remote.ID); err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to update payment external id: %w", err)
	}
	if err := s.repo.UpdatePaymentStatus(ctx, local.ID, remote.Status, remote.InvoiceURL, remote.TransactionReceiptURL); err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to sync payment status: %w", err)
	}
	local.ExternalID = remote.ID
	local.Status = remote.Status
	local.InvoiceURL = remote.InvoiceURL
	local.TransactionReceiptURL = remote.TransactionReceiptURL

	return local, remote, nil
}

// CreateSubscription persists the subscription locally and remotely.
func (s *Service) CreateSubscription(ctx context.Context, req SubscriptionRequest, customerID string) (SubscriptionRecord, SubscriptionResponse, error) {
	now := time.Now().UTC()
	local := SubscriptionRecord{
		ID:          generateID(),
		CustomerID:  customerID,
		Status:      "PENDING",
		Value:       req.Value,
		Cycle:       req.Cycle,
		NextDueDate: parseDate(req.NextDueDate),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.SaveSubscription(ctx, local); err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to save local subscription: %w", err)
	}

	remote, err := s.client.CreateSubscription(ctx, req)
	if err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to create asaas subscription: %w", err)
	}

	if err := s.repo.UpdateSubscriptionExternalID(ctx, local.ID, remote.ID); err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to update subscription external id: %w", err)
	}
	if err := s.repo.UpdateSubscriptionStatus(ctx, local.ID, remote.Status); err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to sync subscription status: %w", err)
	}
	local.ExternalID = remote.ID
	local.Status = remote.Status

	return local, remote, nil
}

// CreateInvoice persists the invoice locally and in Asaas.
func (s *Service) CreateInvoice(ctx context.Context, req InvoiceRequest, customerID string) (InvoiceRecord, InvoiceResponse, error) {
	now := time.Now().UTC()
	local := InvoiceRecord{
		ID:         generateID(),
		CustomerID: customerID,
		Status:     "PENDING",
		Value:      req.Value,
		DueDate:    parseDate(req.DueDate),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.SaveInvoice(ctx, local); err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to save local invoice: %w", err)
	}

	remote, err := s.client.CreateInvoice(ctx, req)
	if err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to create asaas invoice: %w", err)
	}

	if err := s.repo.UpdateInvoiceExternalID(ctx, local.ID, remote.ID); err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to update invoice external id: %w", err)
	}
	if err := s.repo.UpdateInvoiceStatus(ctx, local.ID, remote.Status); err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to sync invoice status: %w", err)
	}
	local.ExternalID = remote.ID
	local.Status = remote.Status
	local.PaymentLink = remote.PaymentLink

	return local, remote, nil
}

// HandleWebhookNotification updates local records based on webhook events.
func (s *Service) HandleWebhookNotification(ctx context.Context, event NotificationEvent) error {
	switch event.Event {
	case "PAYMENT_CREATED", "PAYMENT_RECEIVED", "PAYMENT_CONFIRMED", "PAYMENT_OVERDUE":
		if event.Payment == nil {
			return fmt.Errorf("payment payload missing")
		}
		return s.repo.UpdatePaymentStatus(ctx, event.Payment.ExternalID, event.Payment.Status, event.Payment.InvoiceURL, event.Payment.TransactionReceiptURL)
	case "SUBSCRIPTION_CREATED", "SUBSCRIPTION_UPDATED":
		if event.Subscription == nil {
			return fmt.Errorf("subscription payload missing")
		}
		return s.repo.UpdateSubscriptionStatus(ctx, event.Subscription.ExternalID, event.Subscription.Status)
	case "INVOICE_CREATED", "INVOICE_UPDATED", "INVOICE_OVERDUE":
		if event.Invoice == nil {
			return fmt.Errorf("invoice payload missing")
		}
		return s.repo.UpdateInvoiceStatus(ctx, event.Invoice.ExternalID, event.Invoice.Status)
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
