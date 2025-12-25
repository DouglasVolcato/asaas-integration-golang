package payments

import (
    "context"
    "fmt"
    "time"
)

// Service coordinates calls between the Asaas API client and the repository.
type Service struct {
    client *Client
    repo   *Repository
}

// NewService builds the payment service.
func NewService(client *Client, repo *Repository) *Service {
    return &Service{client: client, repo: repo}
}

// CreateCustomer registers a customer in Asaas and stores the metadata locally.
func (s *Service) CreateCustomer(ctx context.Context, payload Customer) (*CustomerResponse, error) {
    var response CustomerResponse
    if err := s.client.doRequest(ctx, "POST", "/v3/customers", payload, &response); err != nil {
        return nil, err
    }

    rec := CustomerRecord{
        ID:             response.ID,
        ExternalRef:    payload.ExternalRef,
        Name:           payload.Name,
        Email:          payload.Email,
        DocumentNumber: payload.CPF,
        CreatedAt:      time.Now(),
    }
    if err := s.repo.SaveCustomer(ctx, rec); err != nil {
        return nil, fmt.Errorf("store customer: %w", err)
    }
    return &response, nil
}

// CreatePayment builds a payment in Asaas and stores it locally.
func (s *Service) CreatePayment(ctx context.Context, payload Payment) (*PaymentResponse, error) {
    var response PaymentResponse
    if err := s.client.doRequest(ctx, "POST", "/v3/payments", payload, &response); err != nil {
        return nil, err
    }

    rec := PaymentRecord{
        ID:          response.ID,
        CustomerID:  response.CustomerID,
        ExternalRef: payload.ExternalReference,
        Status:      response.Status,
        Value:       response.Value,
        Description: response.Description,
        CreatedAt:   time.Now(),
    }
    if err := s.repo.SavePayment(ctx, rec); err != nil {
        return nil, fmt.Errorf("store payment: %w", err)
    }
    return &response, nil
}

// CreateSubscription builds a recurring subscription and stores locally.
func (s *Service) CreateSubscription(ctx context.Context, payload Subscription) (*SubscriptionResponse, error) {
    var response SubscriptionResponse
    if err := s.client.doRequest(ctx, "POST", "/v3/subscriptions", payload, &response); err != nil {
        return nil, err
    }

    rec := SubscriptionRecord{
        ID:          response.ID,
        CustomerID:  response.CustomerID,
        Status:      response.Status,
        ExternalRef: payload.ExternalReference,
        Value:       payload.Value,
        Cycle:       payload.Cycle,
        CreatedAt:   time.Now(),
    }
    if err := s.repo.SaveSubscription(ctx, rec); err != nil {
        return nil, fmt.Errorf("store subscription: %w", err)
    }
    return &response, nil
}

// CreateInvoice issues an invoice through Asaas and stores it locally.
func (s *Service) CreateInvoice(ctx context.Context, payload Invoice) (*InvoiceResponse, error) {
    var response InvoiceResponse
    if err := s.client.doRequest(ctx, "POST", "/v3/invoices", payload, &response); err != nil {
        return nil, err
    }

    rec := InvoiceRecord{
        ID:          response.ID,
        CustomerID:  response.CustomerID,
        Status:      response.Status,
        Value:       response.Value,
        Description: response.Description,
        CreatedAt:   time.Now(),
    }
    if err := s.repo.SaveInvoice(ctx, rec); err != nil {
        return nil, fmt.Errorf("store invoice: %w", err)
    }
    return &response, nil
}

// GetCustomer retrieves a customer stored in PostgreSQL.
func (s *Service) GetCustomer(ctx context.Context, id string) (*CustomerRecord, error) {
    return s.repo.GetCustomer(ctx, id)
}

// GetPayment retrieves a payment stored in PostgreSQL.
func (s *Service) GetPayment(ctx context.Context, id string) (*PaymentRecord, error) {
    return s.repo.GetPayment(ctx, id)
}
