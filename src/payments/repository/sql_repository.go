package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"asaas/src/payments/models"
)

// SQLRepository implements Storage using PostgreSQL tables.
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a new SQL-backed repository.
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// UpsertCustomer stores or updates a customer record.
func (r *SQLRepository) UpsertCustomer(ctx context.Context, customer models.Customer) error {
	query := `
        INSERT INTO customers (
            id, external_id, name, email, document, phone, mobile_phone, postal_code, address, address_number,
            complement, province, external_reference, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        ) ON CONFLICT (id) DO UPDATE SET
            external_id = EXCLUDED.external_id,
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            document = EXCLUDED.document,
            phone = EXCLUDED.phone,
            mobile_phone = EXCLUDED.mobile_phone,
            postal_code = EXCLUDED.postal_code,
            address = EXCLUDED.address,
            address_number = EXCLUDED.address_number,
            complement = EXCLUDED.complement,
            province = EXCLUDED.province,
            external_reference = EXCLUDED.external_reference,
            updated_at = EXCLUDED.updated_at
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		customer.ID,
		customer.ExternalID,
		customer.Name,
		customer.Email,
		customer.Document,
		customer.Phone,
		customer.MobilePhone,
		customer.PostalCode,
		customer.Address,
		customer.AddressNumber,
		customer.Complement,
		customer.Province,
		customer.ExternalReference,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("upsert customer: %w", err)
	}

	return nil
}

// SavePayment persists a payment row.
func (r *SQLRepository) SavePayment(ctx context.Context, payment models.Payment) error {
	query := `
        INSERT INTO payments (
            id, customer_id, external_id, billing_type, value, due_date, status, description, external_reference,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.CustomerID,
		payment.ExternalID,
		payment.BillingType,
		payment.Value,
		payment.DueDate,
		payment.Status,
		payment.Description,
		payment.ExternalReference,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("save payment: %w", err)
	}

	return nil
}

// UpdatePaymentStatus updates the status for a payment identified by its external ID.
func (r *SQLRepository) UpdatePaymentStatus(ctx context.Context, externalID, status string) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE external_id = $3`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), externalID)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	return nil
}

// SaveSubscription persists subscription information.
func (r *SQLRepository) SaveSubscription(ctx context.Context, subscription models.Subscription) error {
	query := `
        INSERT INTO subscriptions (
            id, customer_id, external_id, billing_type, value, cycle, description, next_due_date, status,
            external_reference, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		subscription.ID,
		subscription.CustomerID,
		subscription.ExternalID,
		subscription.BillingType,
		subscription.Value,
		subscription.Cycle,
		subscription.Description,
		subscription.NextDueDate,
		subscription.Status,
		subscription.ExternalReference,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("save subscription: %w", err)
	}

	return nil
}

// SaveInvoice persists invoice information.
func (r *SQLRepository) SaveInvoice(ctx context.Context, invoice models.Invoice) error {
	query := `
        INSERT INTO invoices (
            id, customer_id, external_id, service_description, value, status, due_date, external_reference,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		invoice.ID,
		invoice.CustomerID,
		invoice.ExternalID,
		invoice.ServiceDescription,
		invoice.Value,
		invoice.Status,
		invoice.DueDate,
		invoice.ExternalReference,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	return nil
}

// UpdateInvoiceStatus updates an invoice status.
func (r *SQLRepository) UpdateInvoiceStatus(ctx context.Context, externalID, status string) error {
	query := `UPDATE invoices SET status = $1, updated_at = $2 WHERE external_id = $3`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), externalID)
	if err != nil {
		return fmt.Errorf("update invoice status: %w", err)
	}

	return nil
}
