package payments

import "time"

// CustomerRequest maps to Asaas customer creation payload.
type CustomerRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email,omitempty"`
	MobilePhone       string `json:"mobilePhone,omitempty"`
	CPF               string `json:"cpf,omitempty"`
	CNPJ              string `json:"cnpj,omitempty"`
	PostalCode        string `json:"postalCode,omitempty"`
	Address           string `json:"address,omitempty"`
	AddressNumber     string `json:"addressNumber,omitempty"`
	Complement        string `json:"complement,omitempty"`
	Province          string `json:"province,omitempty"`
	ExternalReference string `json:"externalReference"`
}

// CustomerResponse mirrors relevant fields returned by Asaas customer endpoints.
type CustomerResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	CPF               string    `json:"cpf"`
	CNPJ              string    `json:"cnpj"`
	ExternalReference string    `json:"externalReference"`
	CreatedAt         time.Time `json:"dateCreated"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// PaymentRequest represents a charge to be created in Asaas.
type PaymentRequest struct {
	CustomerID        string  `json:"customer"`
	BillingType       string  `json:"billingType"`
	Value             float64 `json:"value"`
	DueDate           string  `json:"dueDate"`
	Description       string  `json:"description,omitempty"`
	ExternalReference string  `json:"externalReference"`
}

// PaymentResponse captures fields returned by Asaas payment endpoints.
type PaymentResponse struct {
	ID                string  `json:"id"`
	CustomerID        string  `json:"customer"`
	BillingType       string  `json:"billingType"`
	Value             float64 `json:"value"`
	ExternalReference string  `json:"externalReference"`
	Status            string  `json:"status"`
	InvoiceURL        string  `json:"invoiceUrl"`
}

// SubscriptionRequest reflects Asaas subscription payload expectations.
type SubscriptionRequest struct {
	CustomerID        string  `json:"customer"`
	BillingType       string  `json:"billingType"`
	Value             float64 `json:"value"`
	NextDueDate       string  `json:"nextDueDate"`
	Cycle             string  `json:"cycle"`
	Description       string  `json:"description,omitempty"`
	ExternalReference string  `json:"externalReference"`
}

// SubscriptionResponse captures a subset of Asaas subscription response fields.
type SubscriptionResponse struct {
	ID                string  `json:"id"`
	CustomerID        string  `json:"customer"`
	BillingType       string  `json:"billingType"`
	Value             float64 `json:"value"`
	Status            string  `json:"status"`
	ExternalReference string  `json:"externalReference"`
}

// InvoiceRequest maps to Asaas invoice endpoint payload.
type InvoiceRequest struct {
	CustomerID        string  `json:"customer"`
	Value             float64 `json:"value"`
	Description       string  `json:"description"`
	DueDate           string  `json:"dueDate"`
	ExternalReference string  `json:"externalReference"`
}

// InvoiceResponse captures relevant invoice fields returned by Asaas.
type InvoiceResponse struct {
	ID                string `json:"id"`
	CustomerID        string `json:"customer"`
	Number            string `json:"number"`
	Status            string `json:"status"`
	ExternalReference string `json:"externalReference"`
	SecureURL         string `json:"secureUrl"`
}

// CustomerRecord represents customer data stored locally.
type CustomerRecord struct {
	ID         int64
	ExternalID string
	Name       string
	Email      string
	Document   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PaymentRecord represents a locally stored payment.
type PaymentRecord struct {
	ID                int64
	CustomerID        int64
	ExternalID        string
	ExternalReference string
	BillingType       string
	Value             float64
	Status            string
	DueDate           time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// SubscriptionRecord represents a subscription stored locally.
type SubscriptionRecord struct {
	ID                int64
	CustomerID        int64
	ExternalID        string
	ExternalReference string
	BillingType       string
	Value             float64
	Cycle             string
	Status            string
	NextDueDate       time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// InvoiceRecord represents an invoice stored locally.
type InvoiceRecord struct {
	ID                int64
	CustomerID        int64
	ExternalID        string
	ExternalReference string
	Number            string
	Status            string
	DueDate           time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// WebhookEvent represents incoming notifications from Asaas webhooks.
type WebhookEvent struct {
	Event        string                `json:"event"`
	DateCreated  time.Time             `json:"dateCreated"`
	Attempt      int                   `json:"attempt"`
	Payment      *PaymentResponse      `json:"payment"`
	Subscription *SubscriptionResponse `json:"subscription"`
	Invoice      *InvoiceResponse      `json:"invoice"`
}
