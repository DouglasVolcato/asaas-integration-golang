package payments

import "context"

// EnsureSchema creates the tables required to store integration data when they do not exist.
func EnsureSchema(ctx context.Context, db Database) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS customers (
            id SERIAL PRIMARY KEY,
            external_id VARCHAR(255),
            external_reference VARCHAR(255),
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255),
            document VARCHAR(32),
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS payments (
            id SERIAL PRIMARY KEY,
            customer_id BIGINT REFERENCES customers(id),
            external_id VARCHAR(255),
            external_reference VARCHAR(255),
            billing_type VARCHAR(32) NOT NULL,
            value NUMERIC(12,2) NOT NULL,
            status VARCHAR(32) NOT NULL,
            due_date DATE,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
            id SERIAL PRIMARY KEY,
            customer_id BIGINT REFERENCES customers(id),
            external_id VARCHAR(255),
            external_reference VARCHAR(255),
            billing_type VARCHAR(32) NOT NULL,
            value NUMERIC(12,2) NOT NULL,
            cycle VARCHAR(32) NOT NULL,
            status VARCHAR(32) NOT NULL,
            next_due_date DATE,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS invoices (
            id SERIAL PRIMARY KEY,
            customer_id BIGINT REFERENCES customers(id),
            external_id VARCHAR(255),
            external_reference VARCHAR(255),
            number VARCHAR(64),
            status VARCHAR(32) NOT NULL,
            due_date DATE,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}
