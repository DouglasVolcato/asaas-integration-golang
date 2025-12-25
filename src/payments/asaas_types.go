package payments

import "time"

// CustomerCreateRequest matches the fields needed to create a customer in Asaas.
type CustomerCreateRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email,omitempty"`
	CpfCnpj           string `json:"cpfCnpj,omitempty"`
	MobilePhone       string `json:"mobilePhone,omitempty"`
	Phone             string `json:"phone,omitempty"`
	PostalCode        string `json:"postalCode,omitempty"`
	Address           string `json:"address,omitempty"`
	AddressNumber     string `json:"addressNumber,omitempty"`
	Complement        string `json:"complement,omitempty"`
	Province          string `json:"province,omitempty"`
	ExternalReference string `json:"externalReference,omitempty"`
}

// CustomerResponse represents the response returned by the Asaas API for customers.
type CustomerResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	CpfCnpj           string `json:"cpfCnpj"`
	MobilePhone       string `json:"mobilePhone"`
	Phone             string `json:"phone"`
	ExternalReference string `json:"externalReference"`
}

// BillingType enumerates supported billing methods in Asaas.
type BillingType string

const (
	BillingTypeBoleto     BillingType = "BOLETO"
	BillingTypeCreditCard BillingType = "CREDIT_CARD"
	BillingTypePix        BillingType = "PIX"
)

// PaymentCreateRequest models the payload for creating a payment.
type PaymentCreateRequest struct {
	Customer          string      `json:"customer"`
	BillingType       BillingType `json:"billingType"`
	Value             float64     `json:"value"`
	DueDate           string      `json:"dueDate"`
	Description       string      `json:"description,omitempty"`
	ExternalReference string      `json:"externalReference,omitempty"`
	InstallmentCount  int         `json:"installmentCount,omitempty"`
}

// PaymentResponse captures common fields returned from a payment creation.
type PaymentResponse struct {
	ID                string      `json:"id"`
	Customer          string      `json:"customer"`
	BillingType       BillingType `json:"billingType"`
	Value             float64     `json:"value"`
	DueDate           string      `json:"dueDate"`
	Status            string      `json:"status"`
	ExternalReference string      `json:"externalReference"`
}

// SubscriptionCycle expresses recurrence intervals supported by Asaas.
type SubscriptionCycle string

const (
	SubscriptionCycleWeekly  SubscriptionCycle = "WEEKLY"
	SubscriptionCycleMonthly SubscriptionCycle = "MONTHLY"
	SubscriptionCycleYearly  SubscriptionCycle = "YEARLY"
)

// SubscriptionCreateRequest models a subscription payload.
type SubscriptionCreateRequest struct {
	Customer          string            `json:"customer"`
	BillingType       BillingType       `json:"billingType"`
	Value             float64           `json:"value"`
	Cycle             SubscriptionCycle `json:"cycle"`
	Description       string            `json:"description,omitempty"`
	NextDueDate       string            `json:"nextDueDate"`
	ExternalReference string            `json:"externalReference,omitempty"`
}

// SubscriptionResponse represents the expected Asaas response for subscriptions.
type SubscriptionResponse struct {
	ID                string            `json:"id"`
	Customer          string            `json:"customer"`
	BillingType       BillingType       `json:"billingType"`
	Value             float64           `json:"value"`
	Cycle             SubscriptionCycle `json:"cycle"`
	NextDueDate       string            `json:"nextDueDate"`
	Status            string            `json:"status"`
	ExternalReference string            `json:"externalReference"`
}

// InvoiceCreateRequest describes the payload used for fiscal note generation.
type InvoiceCreateRequest struct {
	Customer           string  `json:"customer"`
	ServiceDescription string  `json:"serviceDescription"`
	Value              float64 `json:"value"`
	DueDate            string  `json:"dueDate"`
	ExternalReference  string  `json:"externalReference,omitempty"`
}

// InvoiceResponse represents the Asaas response for invoices.
type InvoiceResponse struct {
	ID                 string  `json:"id"`
	Customer           string  `json:"customer"`
	ServiceDescription string  `json:"serviceDescription"`
	Value              float64 `json:"value"`
	Status             string  `json:"status"`
	ExternalReference  string  `json:"externalReference"`
}

// WebhookEvent matches notifications sent by Asaas.
type WebhookEvent struct {
	Event   string           `json:"event"`
	Date    time.Time        `json:"date"`
	Payment *PaymentResponse `json:"payment,omitempty"`
	Invoice *InvoiceResponse `json:"invoice,omitempty"`
}
