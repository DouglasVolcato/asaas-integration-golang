package repository

import (
	"context"

	"asaas/src/payments/models"
)

// Storage defines the minimal operations required by the payment service.
type Storage interface {
	UpsertCustomer(ctx context.Context, customer models.Customer) error
	SavePayment(ctx context.Context, payment models.Payment) error
	UpdatePaymentStatus(ctx context.Context, externalID, status string) error
	SaveSubscription(ctx context.Context, subscription models.Subscription) error
	SaveInvoice(ctx context.Context, invoice models.Invoice) error
	UpdateInvoiceStatus(ctx context.Context, externalID, status string) error
}
