package payments

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository defines storage operations required by the service layer.
type Repository interface {
	SaveCustomer(ctx context.Context, customer CustomerRecord) error
	UpdateCustomerExternalID(ctx context.Context, id, externalID string) error
	FindCustomerByExternalID(ctx context.Context, externalID string) (CustomerRecord, error)

	SavePayment(ctx context.Context, payment PaymentRecord) error
	UpdatePaymentStatus(ctx context.Context, id, status, invoiceURL, receiptURL string) error
	UpdatePaymentExternalID(ctx context.Context, id, externalID string) error
	FindPaymentByExternalID(ctx context.Context, externalID string) (PaymentRecord, error)

	SaveSubscription(ctx context.Context, subscription SubscriptionRecord) error
	UpdateSubscriptionStatus(ctx context.Context, id, status string) error
	UpdateSubscriptionExternalID(ctx context.Context, id, externalID string) error

	SaveInvoice(ctx context.Context, invoice InvoiceRecord) error
	UpdateInvoiceStatus(ctx context.Context, id, status string) error
	UpdateInvoiceExternalID(ctx context.Context, id, externalID string) error
}

// PostgresRepository persists data in a PostgreSQL database.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository builds a repository backed by PostgreSQL.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// EnsureSchema creates database tables when they do not exist.
func (r *PostgresRepository) EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS customers (
            id TEXT PRIMARY KEY,
            external_id TEXT,
            name TEXT NOT NULL,
            email TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS payments (
            id TEXT PRIMARY KEY,
            external_id TEXT,
            customer_id TEXT NOT NULL REFERENCES customers(id),
            billing_type TEXT NOT NULL,
            value NUMERIC NOT NULL,
            due_date TIMESTAMPTZ NOT NULL,
            status TEXT NOT NULL,
            invoice_url TEXT,
            transaction_receipt_url TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
            id TEXT PRIMARY KEY,
            external_id TEXT,
            customer_id TEXT NOT NULL REFERENCES customers(id),
            status TEXT NOT NULL,
            value NUMERIC NOT NULL,
            cycle TEXT NOT NULL,
            next_due_date TIMESTAMPTZ NOT NULL,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS invoices (
            id TEXT PRIMARY KEY,
            external_id TEXT,
            customer_id TEXT NOT NULL REFERENCES customers(id),
            status TEXT NOT NULL,
            value NUMERIC NOT NULL,
            effective_date TIMESTAMPTZ NOT NULL,
            payment_link TEXT,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
	}

	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("schema migration failed: %w", err)
		}
	}
	return nil
}

// SaveCustomer inserts a new customer.
func (r *PostgresRepository) SaveCustomer(ctx context.Context, customer CustomerRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO customers (id, external_id, name, email, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, customer.ID, customer.ExternalID, customer.Name, customer.Email, customer.CreatedAt, customer.UpdatedAt)
	return err
}

// UpdateCustomerExternalID updates the external identifier reference.
func (r *PostgresRepository) UpdateCustomerExternalID(ctx context.Context, id, externalID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE customers SET external_id=$1, updated_at=$2 WHERE id=$3`, externalID, time.Now().UTC(), id)
	return err
}

// FindCustomerByExternalID returns a customer record by external reference.
func (r *PostgresRepository) FindCustomerByExternalID(ctx context.Context, externalID string) (CustomerRecord, error) {
	var customer CustomerRecord
	row := r.db.QueryRowContext(ctx, `
        SELECT id, external_id, name, email, created_at, updated_at
        FROM customers
        WHERE external_id = $1
    `, externalID)
	if err := row.Scan(&customer.ID, &customer.ExternalID, &customer.Name, &customer.Email, &customer.CreatedAt, &customer.UpdatedAt); err != nil {
		return CustomerRecord{}, err
	}
	return customer, nil
}

// SavePayment inserts a new payment row.
func (r *PostgresRepository) SavePayment(ctx context.Context, payment PaymentRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO payments (id, external_id, customer_id, billing_type, value, due_date, status, invoice_url, transaction_receipt_url, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, payment.ID, payment.ExternalID, payment.CustomerID, payment.BillingType, payment.Value, payment.DueDate, payment.Status, payment.InvoiceURL, payment.TransactionReceiptURL, payment.CreatedAt, payment.UpdatedAt)
	return err
}

// UpdatePaymentStatus updates the status and links of a payment.
func (r *PostgresRepository) UpdatePaymentStatus(ctx context.Context, id, status, invoiceURL, receiptURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payments SET status=$1, invoice_url=$2, transaction_receipt_url=$3, updated_at=$4 WHERE id=$5`, status, invoiceURL, receiptURL, time.Now().UTC(), id)
	return err
}

// UpdatePaymentExternalID updates the external identifier reference for a payment.
func (r *PostgresRepository) UpdatePaymentExternalID(ctx context.Context, id, externalID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payments SET external_id=$1, updated_at=$2 WHERE id=$3`, externalID, time.Now().UTC(), id)
	return err
}

// FindPaymentByExternalID returns a payment record by external reference.
func (r *PostgresRepository) FindPaymentByExternalID(ctx context.Context, externalID string) (PaymentRecord, error) {
	var payment PaymentRecord
	row := r.db.QueryRowContext(ctx, `
        SELECT id, external_id, customer_id, billing_type, value, due_date, status, invoice_url, transaction_receipt_url, created_at, updated_at
        FROM payments
        WHERE external_id = $1
    `, externalID)
	if err := row.Scan(
		&payment.ID,
		&payment.ExternalID,
		&payment.CustomerID,
		&payment.BillingType,
		&payment.Value,
		&payment.DueDate,
		&payment.Status,
		&payment.InvoiceURL,
		&payment.TransactionReceiptURL,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return PaymentRecord{}, err
	}
	return payment, nil
}

// SaveSubscription inserts a subscription row.
func (r *PostgresRepository) SaveSubscription(ctx context.Context, subscription SubscriptionRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO subscriptions (id, external_id, customer_id, status, value, cycle, next_due_date, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `, subscription.ID, subscription.ExternalID, subscription.CustomerID, subscription.Status, subscription.Value, subscription.Cycle, subscription.NextDueDate, subscription.CreatedAt, subscription.UpdatedAt)
	return err
}

// UpdateSubscriptionStatus updates the subscription status locally.
func (r *PostgresRepository) UpdateSubscriptionStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET status=$1, updated_at=$2 WHERE id=$3`, status, time.Now().UTC(), id)
	return err
}

// UpdateSubscriptionExternalID updates the external identifier reference for a subscription.
func (r *PostgresRepository) UpdateSubscriptionExternalID(ctx context.Context, id, externalID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscriptions SET external_id=$1, updated_at=$2 WHERE id=$3`, externalID, time.Now().UTC(), id)
	return err
}

// SaveInvoice inserts an invoice row.
func (r *PostgresRepository) SaveInvoice(ctx context.Context, invoice InvoiceRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO invoices (id, external_id, customer_id, status, value, effective_date, payment_link, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `, invoice.ID, invoice.ExternalID, invoice.CustomerID, invoice.Status, invoice.Value, invoice.EffectiveDate, invoice.PaymentLink, invoice.CreatedAt, invoice.UpdatedAt)
	return err
}

// UpdateInvoiceStatus updates invoice status locally.
func (r *PostgresRepository) UpdateInvoiceStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE invoices SET status=$1, updated_at=$2 WHERE id=$3`, status, time.Now().UTC(), id)
	return err
}

// UpdateInvoiceExternalID updates the external identifier for an invoice.
func (r *PostgresRepository) UpdateInvoiceExternalID(ctx context.Context, id, externalID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE invoices SET external_id=$1, updated_at=$2 WHERE id=$3`, externalID, time.Now().UTC(), id)
	return err
}

// InMemoryRepository is a testing implementation that keeps data in memory.
type InMemoryRepository struct {
	customers     map[string]CustomerRecord
	payments      map[string]PaymentRecord
	subscriptions map[string]SubscriptionRecord
	invoices      map[string]InvoiceRecord
}

// NewInMemoryRepository creates an in-memory storage for tests.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		customers:     make(map[string]CustomerRecord),
		payments:      make(map[string]PaymentRecord),
		subscriptions: make(map[string]SubscriptionRecord),
		invoices:      make(map[string]InvoiceRecord),
	}
}

func (r *InMemoryRepository) SaveCustomer(_ context.Context, customer CustomerRecord) error {
	r.customers[customer.ID] = customer
	return nil
}

func (r *InMemoryRepository) UpdateCustomerExternalID(_ context.Context, id, externalID string) error {
	c, ok := r.customers[id]
	if !ok {
		return fmt.Errorf("customer %s not found", id)
	}
	c.ExternalID = externalID
	r.customers[id] = c
	return nil
}

func (r *InMemoryRepository) FindCustomerByExternalID(_ context.Context, externalID string) (CustomerRecord, error) {
	for _, customer := range r.customers {
		if customer.ExternalID == externalID {
			return customer, nil
		}
	}
	return CustomerRecord{}, fmt.Errorf("customer externalReference %s not found", externalID)
}

func (r *InMemoryRepository) SavePayment(_ context.Context, payment PaymentRecord) error {
	r.payments[payment.ID] = payment
	return nil
}

func (r *InMemoryRepository) UpdatePaymentStatus(_ context.Context, id, status, invoiceURL, receiptURL string) error {
	p, ok := r.payments[id]
	if !ok {
		return fmt.Errorf("payment %s not found", id)
	}
	p.Status = status
	p.InvoiceURL = invoiceURL
	p.TransactionReceiptURL = receiptURL
	r.payments[id] = p
	return nil
}

func (r *InMemoryRepository) UpdatePaymentExternalID(_ context.Context, id, externalID string) error {
	p, ok := r.payments[id]
	if !ok {
		return fmt.Errorf("payment %s not found", id)
	}
	p.ExternalID = externalID
	r.payments[id] = p
	return nil
}

func (r *InMemoryRepository) FindPaymentByExternalID(_ context.Context, externalID string) (PaymentRecord, error) {
	for _, payment := range r.payments {
		if payment.ExternalID == externalID {
			return payment, nil
		}
	}
	return PaymentRecord{}, fmt.Errorf("payment externalReference %s not found", externalID)
}

func (r *InMemoryRepository) SaveSubscription(_ context.Context, subscription SubscriptionRecord) error {
	r.subscriptions[subscription.ID] = subscription
	return nil
}

func (r *InMemoryRepository) UpdateSubscriptionStatus(_ context.Context, id, status string) error {
	s, ok := r.subscriptions[id]
	if !ok {
		return fmt.Errorf("subscription %s not found", id)
	}
	s.Status = status
	r.subscriptions[id] = s
	return nil
}

func (r *InMemoryRepository) UpdateSubscriptionExternalID(_ context.Context, id, externalID string) error {
	s, ok := r.subscriptions[id]
	if !ok {
		return fmt.Errorf("subscription %s not found", id)
	}
	s.ExternalID = externalID
	r.subscriptions[id] = s
	return nil
}

func (r *InMemoryRepository) SaveInvoice(_ context.Context, invoice InvoiceRecord) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InMemoryRepository) UpdateInvoiceStatus(_ context.Context, id, status string) error {
	inv, ok := r.invoices[id]
	if !ok {
		return fmt.Errorf("invoice %s not found", id)
	}
	inv.Status = status
	r.invoices[id] = inv
	return nil
}

func (r *InMemoryRepository) UpdateInvoiceExternalID(_ context.Context, id, externalID string) error {
	inv, ok := r.invoices[id]
	if !ok {
		return fmt.Errorf("invoice %s not found", id)
	}
	inv.ExternalID = externalID
	r.invoices[id] = inv
	return nil
}
