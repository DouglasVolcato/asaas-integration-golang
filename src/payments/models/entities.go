package models

import "time"

// Customer represents a persisted customer record.
type Customer struct {
	ID                string
	ExternalID        string
	Name              string
	Email             string
	Document          string
	Phone             string
	MobilePhone       string
	PostalCode        string
	Address           string
	AddressNumber     string
	Complement        string
	Province          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExternalReference string
}

// Payment represents a payment record stored locally.
type Payment struct {
	ID                string
	CustomerID        string
	ExternalID        string
	BillingType       string
	Value             float64
	DueDate           time.Time
	Status            string
	Description       string
	ExternalReference string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Subscription represents a recurring billing stored locally.
type Subscription struct {
	ID                string
	CustomerID        string
	ExternalID        string
	BillingType       string
	Value             float64
	Cycle             string
	Description       string
	NextDueDate       time.Time
	Status            string
	ExternalReference string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Invoice represents a fiscal note stored locally.
type Invoice struct {
	ID                 string
	CustomerID         string
	ExternalID         string
	ServiceDescription string
	Value              float64
	Status             string
	DueDate            time.Time
	ExternalReference  string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
