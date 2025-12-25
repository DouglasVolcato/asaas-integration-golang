package payments

import "time"

// CustomerRecord represents a customer stored in the local database.
type CustomerRecord struct {
	ID         string
	ExternalID string
	Name       string
	Email      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PaymentRecord represents a payment persisted locally.
type PaymentRecord struct {
	ID                    string
	ExternalID            string
	CustomerID            string
	BillingType           string
	Value                 float64
	DueDate               time.Time
	Status                string
	InvoiceURL            string
	TransactionReceiptURL string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// SubscriptionRecord represents a subscription persisted locally.
type SubscriptionRecord struct {
	ID          string
	ExternalID  string
	CustomerID  string
	Status      string
	Value       float64
	Cycle       string
	NextDueDate time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InvoiceRecord represents an invoice persisted locally.
type InvoiceRecord struct {
	ID          string
	ExternalID  string
	CustomerID  string
	Status      string
	Value       float64
	DueDate     time.Time
	PaymentLink string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
