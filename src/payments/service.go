package payments

import (
	"context"
	"fmt"
	"time"

	"asaas/src/payments/models"
	"asaas/src/payments/repository"
)

// Service orchestrates API calls and local persistence without coupling to Asaas.
type Service struct {
	client *AsaasClient
	store  repository.Storage
}

// NewService builds a Service instance.
func NewService(client *AsaasClient, store repository.Storage) *Service {
	return &Service{client: client, store: store}
}

// CreateCustomer creates a customer in Asaas and persists the record locally.
func (s *Service) CreateCustomer(ctx context.Context, payload CustomerCreateRequest, localID string) (*CustomerResponse, error) {
	if payload.Name == "" {
		return nil, fmt.Errorf("customer name is required")
	}

	var response CustomerResponse
	if err := s.client.doRequest(ctx, httpMethodPost, "/v3/customers", payload, &response); err != nil {
		return nil, err
	}

	customer := models.Customer{
		ID:                localID,
		ExternalID:        response.ID,
		Name:              response.Name,
		Email:             response.Email,
		Document:          response.CpfCnpj,
		Phone:             response.Phone,
		MobilePhone:       response.MobilePhone,
		ExternalReference: response.ExternalReference,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.store.UpsertCustomer(ctx, customer); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreatePayment creates a payment in Asaas and stores a local record.
func (s *Service) CreatePayment(ctx context.Context, payload PaymentCreateRequest, localID, customerID string) (*PaymentResponse, error) {
	if payload.Customer == "" || payload.BillingType == "" || payload.Value == 0 || payload.DueDate == "" {
		return nil, fmt.Errorf("customer, billingType, value and dueDate are required")
	}

	var response PaymentResponse
	if err := s.client.doRequest(ctx, httpMethodPost, "/v3/payments", payload, &response); err != nil {
		return nil, err
	}

	dueDate, _ := time.Parse("2006-01-02", response.DueDate)

	payment := models.Payment{
		ID:                localID,
		CustomerID:        customerID,
		ExternalID:        response.ID,
		BillingType:       string(response.BillingType),
		Value:             response.Value,
		DueDate:           dueDate,
		Status:            response.Status,
		ExternalReference: response.ExternalReference,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.store.SavePayment(ctx, payment); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateSubscription creates a subscription in Asaas and stores it locally.
func (s *Service) CreateSubscription(ctx context.Context, payload SubscriptionCreateRequest, localID, customerID string) (*SubscriptionResponse, error) {
	if payload.Customer == "" || payload.BillingType == "" || payload.Value == 0 || payload.Cycle == "" || payload.NextDueDate == "" {
		return nil, fmt.Errorf("customer, billingType, value, cycle and nextDueDate are required")
	}

	var response SubscriptionResponse
	if err := s.client.doRequest(ctx, httpMethodPost, "/v3/subscriptions", payload, &response); err != nil {
		return nil, err
	}

	nextDueDate, _ := time.Parse("2006-01-02", response.NextDueDate)

	subscription := models.Subscription{
		ID:                localID,
		CustomerID:        customerID,
		ExternalID:        response.ID,
		BillingType:       string(response.BillingType),
		Value:             response.Value,
		Cycle:             string(response.Cycle),
		NextDueDate:       nextDueDate,
		Status:            response.Status,
		ExternalReference: response.ExternalReference,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.store.SaveSubscription(ctx, subscription); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateInvoice creates an invoice in Asaas and stores it locally.
func (s *Service) CreateInvoice(ctx context.Context, payload InvoiceCreateRequest, localID, customerID string) (*InvoiceResponse, error) {
	if payload.Customer == "" || payload.ServiceDescription == "" || payload.Value == 0 || payload.DueDate == "" {
		return nil, fmt.Errorf("customer, serviceDescription, value and dueDate are required")
	}

	var response InvoiceResponse
	if err := s.client.doRequest(ctx, httpMethodPost, "/v3/invoices", payload, &response); err != nil {
		return nil, err
	}

	dueDate, _ := time.Parse("2006-01-02", payload.DueDate)

	invoice := models.Invoice{
		ID:                 localID,
		CustomerID:         customerID,
		ExternalID:         response.ID,
		ServiceDescription: response.ServiceDescription,
		Value:              response.Value,
		Status:             response.Status,
		DueDate:            dueDate,
		ExternalReference:  response.ExternalReference,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.store.SaveInvoice(ctx, invoice); err != nil {
		return nil, err
	}

	return &response, nil
}

// HandlePaymentStatus allows manual status updates triggered by webhooks.
func (s *Service) HandlePaymentStatus(ctx context.Context, externalID, status string) error {
	if externalID == "" || status == "" {
		return fmt.Errorf("externalID and status are required")
	}
	return s.store.UpdatePaymentStatus(ctx, externalID, status)
}

// HandleInvoiceStatus allows manual invoice status updates triggered by webhooks.
func (s *Service) HandleInvoiceStatus(ctx context.Context, externalID, status string) error {
	if externalID == "" || status == "" {
		return fmt.Errorf("externalID and status are required")
	}
	return s.store.UpdateInvoiceStatus(ctx, externalID, status)
}
