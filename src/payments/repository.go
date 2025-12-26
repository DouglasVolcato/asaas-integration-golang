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
	UpdateCustomerExternalReference(ctx context.Context, id, externalReference string) error
	FindCustomerByExternalReference(ctx context.Context, externalReference string) (CustomerRecord, error)

	SavePayment(ctx context.Context, payment PaymentRecord) error
	UpdatePaymentStatus(ctx context.Context, externalReference, status, invoiceURL, receiptURL string) error
	UpdatePaymentExternalReference(ctx context.Context, id, externalReference string) error
	FindPaymentByExternalReference(ctx context.Context, externalReference string) (PaymentRecord, error)

	SaveSubscription(ctx context.Context, subscription SubscriptionRecord) error
	UpdateSubscriptionStatus(ctx context.Context, externalReference, status string) error
	UpdateSubscriptionExternalReference(ctx context.Context, id, externalReference string) error

	SaveInvoice(ctx context.Context, invoice InvoiceRecord) error
	UpdateInvoiceStatus(ctx context.Context, externalReference, status string) error
	UpdateInvoiceExternalReference(ctx context.Context, id, externalReference string) error
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
		`CREATE TABLE IF NOT EXISTS payment_customers (
            id UUID PRIMARY KEY,
            external_reference TEXT DEFAULT '',
            name TEXT NOT NULL,
            email TEXT DEFAULT '',
            cpfCnpj TEXT DEFAULT '',
            phone TEXT DEFAULT '',
            mobile_phone TEXT DEFAULT '',
            address TEXT DEFAULT '',
            address_number TEXT DEFAULT '',
            complement TEXT DEFAULT '',
            province TEXT DEFAULT '',
            postal_code TEXT DEFAULT '',
            notification_disabled BOOLEAN NOT NULL DEFAULT FALSE,
            additional_emails TEXT DEFAULT '',
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS payment_payments (
            id UUID PRIMARY KEY,
            external_reference TEXT DEFAULT '',
            customer_id UUID NOT NULL REFERENCES payment_customers(id),
            customer_external_reference TEXT DEFAULT '',
            billing_type TEXT NOT NULL,
            value NUMERIC NOT NULL,
            due_date TIMESTAMPTZ NOT NULL,
            description TEXT DEFAULT '',
            installment_count INTEGER NOT NULL DEFAULT 0,
            callback_success_url TEXT DEFAULT '',
            callback_auto_redirect BOOLEAN NOT NULL DEFAULT FALSE,
            status TEXT DEFAULT '',
            invoice_url TEXT DEFAULT '',
            transaction_receipt_url TEXT DEFAULT '',
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS payment_subscriptions (
            id UUID PRIMARY KEY,
            external_reference TEXT DEFAULT '',
            customer_id UUID NOT NULL REFERENCES payment_customers(id),
            customer_external_reference TEXT DEFAULT '',
            billing_type TEXT NOT NULL,
            status TEXT DEFAULT '',
            value NUMERIC NOT NULL,
            cycle TEXT NOT NULL,
            next_due_date TIMESTAMPTZ NOT NULL,
            description TEXT DEFAULT '',
            end_date TIMESTAMPTZ,
            max_payments INTEGER NOT NULL DEFAULT 0,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS payment_invoices (
            id UUID PRIMARY KEY,
            external_reference TEXT DEFAULT '',
            payment_id UUID NOT NULL REFERENCES payment_payments(id),
            payment_external_reference TEXT DEFAULT '',
            service_description TEXT NOT NULL,
            observations TEXT NOT NULL,
            value NUMERIC NOT NULL,
            deductions NUMERIC NOT NULL DEFAULT 0,
            effective_date TIMESTAMPTZ NOT NULL,
            municipal_service_id TEXT DEFAULT '',
            municipal_service_code TEXT DEFAULT '',
            municipal_service_name TEXT NOT NULL,
            update_payment BOOLEAN NOT NULL DEFAULT FALSE,
            taxes_retain_iss BOOLEAN NOT NULL DEFAULT FALSE,
            taxes_cofins NUMERIC NOT NULL DEFAULT 0,
            taxes_csll NUMERIC NOT NULL DEFAULT 0,
            taxes_inss NUMERIC NOT NULL DEFAULT 0,
            taxes_ir NUMERIC NOT NULL DEFAULT 0,
            taxes_pis NUMERIC NOT NULL DEFAULT 0,
            taxes_iss NUMERIC NOT NULL DEFAULT 0,
            status TEXT DEFAULT '',
            payment_link TEXT DEFAULT '',
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
        INSERT INTO payment_customers (
            id,
            external_reference,
            name,
            email,
            cpfCnpj,
            phone,
            mobile_phone,
            address,
            address_number,
            complement,
            province,
            postal_code,
            notification_disabled,
            additional_emails,
            created_at,
            updated_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
    `,
		customer.ID,
		customer.ExternalReference,
		customer.Name,
		customer.Email,
		customer.CpfCnpj,
		customer.Phone,
		customer.MobilePhone,
		customer.Address,
		customer.AddressNumber,
		customer.Complement,
		customer.Province,
		customer.PostalCode,
		customer.NotificationDisabled,
		customer.AdditionalEmails,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	return err
}

// UpdateCustomerExternalReference updates the external reference.
func (r *PostgresRepository) UpdateCustomerExternalReference(ctx context.Context, id, externalReference string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_customers SET external_reference=$1, updated_at=$2 WHERE id=$3`, externalReference, time.Now().UTC(), id)
	return err
}

// FindCustomerByExternalReference returns a customer record by external reference.
func (r *PostgresRepository) FindCustomerByExternalReference(ctx context.Context, externalReference string) (CustomerRecord, error) {
	var customer CustomerRecord
	row := r.db.QueryRowContext(ctx, `
        SELECT
            id,
            external_reference,
            name,
            email,
            cpfCnpj,
            phone,
            mobile_phone,
            address,
            address_number,
            complement,
            province,
            postal_code,
            notification_disabled,
            additional_emails,
            created_at,
            updated_at
        FROM payment_customers
        WHERE external_reference = $1
    `, externalReference)
	if err := row.Scan(
		&customer.ID,
		&customer.ExternalReference,
		&customer.Name,
		&customer.Email,
		&customer.CpfCnpj,
		&customer.Phone,
		&customer.MobilePhone,
		&customer.Address,
		&customer.AddressNumber,
		&customer.Complement,
		&customer.Province,
		&customer.PostalCode,
		&customer.NotificationDisabled,
		&customer.AdditionalEmails,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		return CustomerRecord{}, err
	}
	return customer, nil
}

// SavePayment inserts a new payment row.
func (r *PostgresRepository) SavePayment(ctx context.Context, payment PaymentRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO payment_payments (
            id,
            external_reference,
            customer_id,
            customer_external_reference,
            billing_type,
            value,
            due_date,
            description,
            installment_count,
            callback_success_url,
            callback_auto_redirect,
            status,
            invoice_url,
            transaction_receipt_url,
            created_at,
            updated_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
    `,
		payment.ID,
		payment.ExternalReference,
		payment.CustomerID,
		payment.CustomerExternalReference,
		payment.BillingType,
		payment.Value,
		payment.DueDate,
		payment.Description,
		payment.InstallmentCount,
		payment.CallbackSuccessURL,
		payment.CallbackAutoRedirect,
		payment.Status,
		payment.InvoiceURL,
		payment.TransactionReceiptURL,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

// UpdatePaymentStatus updates the status and links of a payment.
func (r *PostgresRepository) UpdatePaymentStatus(ctx context.Context, externalReference, status, invoiceURL, receiptURL string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE payment_payments SET status=$1, invoice_url=$2, transaction_receipt_url=$3, updated_at=$4 WHERE external_reference=$5`,
		status,
		invoiceURL,
		receiptURL,
		time.Now().UTC(),
		externalReference,
	)
	return err
}

// UpdatePaymentExternalReference updates the external reference for a payment.
func (r *PostgresRepository) UpdatePaymentExternalReference(ctx context.Context, id, externalReference string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_payments SET external_reference=$1, updated_at=$2 WHERE id=$3`, externalReference, time.Now().UTC(), id)
	return err
}

// FindPaymentByExternalReference returns a payment record by external reference.
func (r *PostgresRepository) FindPaymentByExternalReference(ctx context.Context, externalReference string) (PaymentRecord, error) {
	var payment PaymentRecord
	row := r.db.QueryRowContext(ctx, `
        SELECT
            id,
            external_reference,
            customer_id,
            customer_external_reference,
            billing_type,
            value,
            due_date,
            description,
            installment_count,
            callback_success_url,
            callback_auto_redirect,
            status,
            invoice_url,
            transaction_receipt_url,
            created_at,
            updated_at
        FROM payment_payments
        WHERE external_reference = $1
    `, externalReference)
	if err := row.Scan(
		&payment.ID,
		&payment.ExternalReference,
		&payment.CustomerID,
		&payment.CustomerExternalReference,
		&payment.BillingType,
		&payment.Value,
		&payment.DueDate,
		&payment.Description,
		&payment.InstallmentCount,
		&payment.CallbackSuccessURL,
		&payment.CallbackAutoRedirect,
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
        INSERT INTO payment_subscriptions (
            id,
            external_reference,
            customer_id,
            customer_external_reference,
            billing_type,
            status,
            value,
            cycle,
            next_due_date,
            description,
            end_date,
            max_payments,
            created_at,
            updated_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
    `,
		subscription.ID,
		subscription.ExternalReference,
		subscription.CustomerID,
		subscription.CustomerExternalReference,
		subscription.BillingType,
		subscription.Status,
		subscription.Value,
		subscription.Cycle,
		subscription.NextDueDate,
		subscription.Description,
		subscription.EndDate,
		subscription.MaxPayments,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)
	return err
}

// UpdateSubscriptionStatus updates the subscription status locally.
func (r *PostgresRepository) UpdateSubscriptionStatus(ctx context.Context, externalReference, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_subscriptions SET status=$1, updated_at=$2 WHERE external_reference=$3`, status, time.Now().UTC(), externalReference)
	return err
}

// UpdateSubscriptionExternalReference updates the external reference for a subscription.
func (r *PostgresRepository) UpdateSubscriptionExternalReference(ctx context.Context, id, externalReference string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_subscriptions SET external_reference=$1, updated_at=$2 WHERE id=$3`, externalReference, time.Now().UTC(), id)
	return err
}

// SaveInvoice inserts an invoice row.
func (r *PostgresRepository) SaveInvoice(ctx context.Context, invoice InvoiceRecord) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO payment_invoices (
            id,
            external_reference,
            payment_id,
            payment_external_reference,
            service_description,
            observations,
            value,
            deductions,
            effective_date,
            municipal_service_id,
            municipal_service_code,
            municipal_service_name,
            update_payment,
            taxes_retain_iss,
            taxes_cofins,
            taxes_csll,
            taxes_inss,
            taxes_ir,
            taxes_pis,
            taxes_iss,
            status,
            payment_link,
            created_at,
            updated_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
    `,
		invoice.ID,
		invoice.ExternalReference,
		invoice.PaymentID,
		invoice.PaymentExternalReference,
		invoice.ServiceDescription,
		invoice.Observations,
		invoice.Value,
		invoice.Deductions,
		invoice.EffectiveDate,
		invoice.MunicipalServiceID,
		invoice.MunicipalServiceCode,
		invoice.MunicipalServiceName,
		invoice.UpdatePayment,
		invoice.TaxesRetainISS,
		invoice.TaxesCofins,
		invoice.TaxesCsll,
		invoice.TaxesINSS,
		invoice.TaxesIR,
		invoice.TaxesPIS,
		invoice.TaxesISS,
		invoice.Status,
		invoice.PaymentLink,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)
	return err
}

// UpdateInvoiceStatus updates invoice status locally.
func (r *PostgresRepository) UpdateInvoiceStatus(ctx context.Context, externalReference, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_invoices SET status=$1, updated_at=$2 WHERE external_reference=$3`, status, time.Now().UTC(), externalReference)
	return err
}

// UpdateInvoiceExternalReference updates the external reference for an invoice.
func (r *PostgresRepository) UpdateInvoiceExternalReference(ctx context.Context, id, externalReference string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_invoices SET external_reference=$1, updated_at=$2 WHERE id=$3`, externalReference, time.Now().UTC(), id)
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

func (r *InMemoryRepository) UpdateCustomerExternalReference(_ context.Context, id, externalReference string) error {
	c, ok := r.customers[id]
	if !ok {
		return fmt.Errorf("customer %s not found", id)
	}
	c.ExternalReference = externalReference
	r.customers[id] = c
	return nil
}

func (r *InMemoryRepository) FindCustomerByExternalReference(_ context.Context, externalReference string) (CustomerRecord, error) {
	for _, customer := range r.customers {
		if customer.ExternalReference == externalReference {
			return customer, nil
		}
	}
	return CustomerRecord{}, fmt.Errorf("customer externalReference %s not found", externalReference)
}

func (r *InMemoryRepository) SavePayment(_ context.Context, payment PaymentRecord) error {
	r.payments[payment.ID] = payment
	return nil
}

func (r *InMemoryRepository) UpdatePaymentStatus(_ context.Context, externalReference, status, invoiceURL, receiptURL string) error {
	for id, payment := range r.payments {
		if payment.ExternalReference == externalReference {
			payment.Status = status
			payment.InvoiceURL = invoiceURL
			payment.TransactionReceiptURL = receiptURL
			r.payments[id] = payment
			return nil
		}
	}
	return fmt.Errorf("payment externalReference %s not found", externalReference)
}

func (r *InMemoryRepository) UpdatePaymentExternalReference(_ context.Context, id, externalReference string) error {
	p, ok := r.payments[id]
	if !ok {
		return fmt.Errorf("payment %s not found", id)
	}
	p.ExternalReference = externalReference
	r.payments[id] = p
	return nil
}

func (r *InMemoryRepository) FindPaymentByExternalReference(_ context.Context, externalReference string) (PaymentRecord, error) {
	for _, payment := range r.payments {
		if payment.ExternalReference == externalReference {
			return payment, nil
		}
	}
	return PaymentRecord{}, fmt.Errorf("payment externalReference %s not found", externalReference)
}

func (r *InMemoryRepository) SaveSubscription(_ context.Context, subscription SubscriptionRecord) error {
	r.subscriptions[subscription.ID] = subscription
	return nil
}

func (r *InMemoryRepository) UpdateSubscriptionStatus(_ context.Context, externalReference, status string) error {
	for id, sub := range r.subscriptions {
		if sub.ExternalReference == externalReference {
			sub.Status = status
			r.subscriptions[id] = sub
			return nil
		}
	}
	return fmt.Errorf("subscription externalReference %s not found", externalReference)
}

func (r *InMemoryRepository) UpdateSubscriptionExternalReference(_ context.Context, id, externalReference string) error {
	s, ok := r.subscriptions[id]
	if !ok {
		return fmt.Errorf("subscription %s not found", id)
	}
	s.ExternalReference = externalReference
	r.subscriptions[id] = s
	return nil
}

func (r *InMemoryRepository) SaveInvoice(_ context.Context, invoice InvoiceRecord) error {
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InMemoryRepository) UpdateInvoiceStatus(_ context.Context, externalReference, status string) error {
	for id, inv := range r.invoices {
		if inv.ExternalReference == externalReference {
			inv.Status = status
			r.invoices[id] = inv
			return nil
		}
	}
	return fmt.Errorf("invoice externalReference %s not found", externalReference)
}

func (r *InMemoryRepository) UpdateInvoiceExternalReference(_ context.Context, id, externalReference string) error {
	inv, ok := r.invoices[id]
	if !ok {
		return fmt.Errorf("invoice %s not found", id)
	}
	inv.ExternalReference = externalReference
	r.invoices[id] = inv
	return nil
}
