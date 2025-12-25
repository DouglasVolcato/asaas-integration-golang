package payments

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
)

// Repository handles PostgreSQL operations without coupling to payment provider.
type Repository struct {
    db *sql.DB
}

// NewRepository builds a repository with a SQL database connection.
func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

// InitSchema creates required tables if they do not exist.
func (r *Repository) InitSchema(ctx context.Context) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS customers (
            id TEXT PRIMARY KEY,
            external_reference TEXT,
            name TEXT NOT NULL,
            email TEXT,
            document_number TEXT,
            created_at TIMESTAMPTZ DEFAULT NOW()
        )`,
        `CREATE TABLE IF NOT EXISTS payments (
            id TEXT PRIMARY KEY,
            customer_id TEXT NOT NULL,
            external_reference TEXT,
            status TEXT,
            value NUMERIC,
            description TEXT,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        )`,
        `CREATE TABLE IF NOT EXISTS subscriptions (
            id TEXT PRIMARY KEY,
            customer_id TEXT NOT NULL,
            status TEXT,
            external_reference TEXT,
            value NUMERIC,
            cycle TEXT,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        )`,
        `CREATE TABLE IF NOT EXISTS invoices (
            id TEXT PRIMARY KEY,
            customer_id TEXT NOT NULL,
            status TEXT,
            value NUMERIC,
            description TEXT,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        )`,
    }

    for _, stmt := range stmts {
        if _, err := r.db.ExecContext(ctx, stmt); err != nil {
            return fmt.Errorf("create table: %w", err)
        }
    }

    return nil
}

// SaveCustomer persists a customer record locally.
func (r *Repository) SaveCustomer(ctx context.Context, rec CustomerRecord) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO customers (id, external_reference, name, email, document_number, created_at) VALUES ($1,$2,$3,$4,$5,COALESCE($6, NOW())) ON CONFLICT (id) DO NOTHING`,
        rec.ID, rec.ExternalRef, rec.Name, rec.Email, rec.DocumentNumber, rec.CreatedAt)
    return err
}

// SavePayment persists payment metadata.
func (r *Repository) SavePayment(ctx context.Context, rec PaymentRecord) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO payments (id, customer_id, external_reference, status, value, description, created_at) VALUES ($1,$2,$3,$4,$5,$6,COALESCE($7, NOW())) ON CONFLICT (id) DO NOTHING`,
        rec.ID, rec.CustomerID, rec.ExternalRef, rec.Status, rec.Value, rec.Description, rec.CreatedAt)
    return err
}

// SaveSubscription persists subscription metadata.
func (r *Repository) SaveSubscription(ctx context.Context, rec SubscriptionRecord) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO subscriptions (id, customer_id, status, external_reference, value, cycle, created_at) VALUES ($1,$2,$3,$4,$5,$6,COALESCE($7, NOW())) ON CONFLICT (id) DO NOTHING`,
        rec.ID, rec.CustomerID, rec.Status, rec.ExternalRef, rec.Value, rec.Cycle, rec.CreatedAt)
    return err
}

// SaveInvoice persists invoice metadata.
func (r *Repository) SaveInvoice(ctx context.Context, rec InvoiceRecord) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO invoices (id, customer_id, status, value, description, created_at) VALUES ($1,$2,$3,$4,$5,COALESCE($6, NOW())) ON CONFLICT (id) DO NOTHING`,
        rec.ID, rec.CustomerID, rec.Status, rec.Value, rec.Description, rec.CreatedAt)
    return err
}

// GetCustomer returns a stored customer.
func (r *Repository) GetCustomer(ctx context.Context, id string) (*CustomerRecord, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, external_reference, name, email, document_number, created_at FROM customers WHERE id=$1`, id)
    var rec CustomerRecord
    if err := row.Scan(&rec.ID, &rec.ExternalRef, &rec.Name, &rec.Email, &rec.DocumentNumber, &rec.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &rec, nil
}

// GetPayment returns a stored payment.
func (r *Repository) GetPayment(ctx context.Context, id string) (*PaymentRecord, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, customer_id, external_reference, status, value, description, created_at FROM payments WHERE id=$1`, id)
    var rec PaymentRecord
    if err := row.Scan(&rec.ID, &rec.CustomerID, &rec.ExternalRef, &rec.Status, &rec.Value, &rec.Description, &rec.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &rec, nil
}
