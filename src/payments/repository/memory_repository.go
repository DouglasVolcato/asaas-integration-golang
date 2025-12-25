package repository

import (
	"context"
	"sync"
	"time"

	"asaas/src/payments/models"
)

// MemoryRepository offers an in-memory implementation useful for tests and local runs.
type MemoryRepository struct {
	mu            sync.RWMutex
	customers     map[string]models.Customer
	payments      map[string]models.Payment
	subscriptions map[string]models.Subscription
	invoices      map[string]models.Invoice
}

// NewMemoryRepository builds a new in-memory store.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		customers:     make(map[string]models.Customer),
		payments:      make(map[string]models.Payment),
		subscriptions: make(map[string]models.Subscription),
		invoices:      make(map[string]models.Invoice),
	}
}

// UpsertCustomer stores customer data by external ID to avoid duplicates.
func (r *MemoryRepository) UpsertCustomer(ctx context.Context, customer models.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	customer.UpdatedAt = time.Now()
	if customer.CreatedAt.IsZero() {
		customer.CreatedAt = time.Now()
	}
	r.customers[customer.ExternalID] = customer
	return nil
}

// SavePayment saves or overwrites a payment by external ID.
func (r *MemoryRepository) SavePayment(ctx context.Context, payment models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	payment.UpdatedAt = time.Now()
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
	}
	r.payments[payment.ExternalID] = payment
	return nil
}

// UpdatePaymentStatus updates an existing payment status.
func (r *MemoryRepository) UpdatePaymentStatus(ctx context.Context, externalID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if payment, ok := r.payments[externalID]; ok {
		payment.Status = status
		payment.UpdatedAt = time.Now()
		r.payments[externalID] = payment
	}
	return nil
}

// SaveSubscription stores subscription data.
func (r *MemoryRepository) SaveSubscription(ctx context.Context, subscription models.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subscription.UpdatedAt = time.Now()
	if subscription.CreatedAt.IsZero() {
		subscription.CreatedAt = time.Now()
	}
	r.subscriptions[subscription.ExternalID] = subscription
	return nil
}

// SaveInvoice stores invoice information.
func (r *MemoryRepository) SaveInvoice(ctx context.Context, invoice models.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	invoice.UpdatedAt = time.Now()
	if invoice.CreatedAt.IsZero() {
		invoice.CreatedAt = time.Now()
	}
	r.invoices[invoice.ExternalID] = invoice
	return nil
}

// UpdateInvoiceStatus updates invoice status.
func (r *MemoryRepository) UpdateInvoiceStatus(ctx context.Context, externalID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if invoice, ok := r.invoices[externalID]; ok {
		invoice.Status = status
		invoice.UpdatedAt = time.Now()
		r.invoices[externalID] = invoice
	}
	return nil
}

// PaymentsSnapshot returns a copy of payments map for safe read access.
func (r *MemoryRepository) PaymentsSnapshot() map[string]models.Payment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]models.Payment, len(r.payments))
	for k, v := range r.payments {
		out[k] = v
	}
	return out
}

// InvoicesSnapshot returns a copy of invoices map for safe read access.
func (r *MemoryRepository) InvoicesSnapshot() map[string]models.Invoice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]models.Invoice, len(r.invoices))
	for k, v := range r.invoices {
		out[k] = v
	}
	return out
}
