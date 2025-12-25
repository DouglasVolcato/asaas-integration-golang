package payments

import "time"

// Customer represents the payload required by Asaas to create a customer.
type Customer struct {
    Name           string  `json:"name"`
    Email          string  `json:"email,omitempty"`
    Phone          string  `json:"phone,omitempty"`
    MobilePhone    string  `json:"mobilePhone,omitempty"`
    CPF            string  `json:"cpfCnpj,omitempty"`
    Company        string  `json:"company,omitempty"`
    Address        string  `json:"address,omitempty"`
    AddressNumber  string  `json:"addressNumber,omitempty"`
    Complement     string  `json:"complement,omitempty"`
    Province       string  `json:"province,omitempty"`
    PostalCode     string  `json:"postalCode,omitempty"`
    ExternalRef    string  `json:"externalReference,omitempty"`
    Notification   bool    `json:"notificationDisabled,omitempty"`
    AdditionalInfo string  `json:"additionalEmails,omitempty"`
    MunicipalInscr string  `json:"municipalInscription,omitempty"`
    StateInscr     string  `json:"stateInscription,omitempty"`
    GroupName      string  `json:"groupName,omitempty"`
}

// CustomerResponse includes the ID returned by Asaas.
type CustomerResponse struct {
    ID string `json:"id"`
    Customer
}

// Payment represents a payment creation payload.
type Payment struct {
    CustomerID         string    `json:"customer"`
    BillingType        string    `json:"billingType"`
    Value              float64   `json:"value"`
    DueDate            string    `json:"dueDate,omitempty"`
    Description        string    `json:"description,omitempty"`
    InstallmentCount   int       `json:"installmentCount,omitempty"`
    InstallmentValue   float64   `json:"installmentValue,omitempty"`
    ExternalReference  string    `json:"externalReference,omitempty"`
    Card               *CardData `json:"creditCard,omitempty"`
    CardHolder         *Payer    `json:"creditCardHolderInfo,omitempty"`
    Callback           *Callback `json:"callback,omitempty"`
}

// PaymentResponse captures the payment ID from Asaas.
type PaymentResponse struct {
    ID          string  `json:"id"`
    Status      string  `json:"status"`
    InvoiceURL  string  `json:"invoiceUrl,omitempty"`
    InvoiceID   string  `json:"invoiceId,omitempty"`
    Description string  `json:"description,omitempty"`
    Value       float64 `json:"value,omitempty"`
    CustomerID  string  `json:"customer"`
}

// Subscription represents a recurring billing payload.
type Subscription struct {
    CustomerID        string  `json:"customer"`
    BillingType       string  `json:"billingType"`
    Value             float64 `json:"value"`
    NextDueDate       string  `json:"nextDueDate"`
    Cycle             string  `json:"cycle,omitempty"`
    Description       string  `json:"description,omitempty"`
    ExternalReference string  `json:"externalReference,omitempty"`
}

// SubscriptionResponse holds data returned from the API.
type SubscriptionResponse struct {
    ID     string `json:"id"`
    Status string `json:"status"`
    Subscription
}

// Invoice represents invoice creation request.
type Invoice struct {
    CustomerID        string  `json:"customer"`
    BillingType       string  `json:"billingType"`
    Value             float64 `json:"value"`
    DueDate           string  `json:"dueDate"`
    Description       string  `json:"description,omitempty"`
    ExternalReference string  `json:"externalReference,omitempty"`
}

// InvoiceResponse describes the invoice returned.
type InvoiceResponse struct {
    ID          string  `json:"id"`
    Status      string  `json:"status"`
    InvoiceURL  string  `json:"invoiceUrl"`
    CustomerID  string  `json:"customer"`
    Value       float64 `json:"value"`
    Description string  `json:"description"`
}

// CardData stores credit card information required by Asaas.
type CardData struct {
    HolderName     string `json:"holderName"`
    Number         string `json:"number"`
    ExpirationMonth string `json:"expiryMonth"`
    ExpirationYear  string `json:"expiryYear"`
    Cvv            string `json:"ccv"`
}

// Payer contains card holder details.
type Payer struct {
    Name           string `json:"name"`
    Email          string `json:"email"`
    CPF            string `json:"cpfCnpj"`
    PostalCode     string `json:"postalCode"`
    AddressNumber  string `json:"addressNumber"`
    Phone          string `json:"phone"`
}

// Callback contains webhook URLs.
type Callback struct {
    SuccessURL string `json:"successUrl,omitempty"`
    FailureURL string `json:"failureUrl,omitempty"`
}

// WebhookEvent represents Asaas webhook notifications.
type WebhookEvent struct {
    Event     string          `json:"event"`
    Date      time.Time       `json:"date"`
    Payment   *PaymentWebhook `json:"payment"`
    Customer  *CustomerHook   `json:"customer"`
    Invoice   *InvoiceHook    `json:"invoice"`
    Signature string          `json:"signature"`
}

// PaymentWebhook describes a payment notification.
type PaymentWebhook struct {
    ID         string  `json:"id"`
    Status     string  `json:"status"`
    Value      float64 `json:"value"`
    CustomerID string  `json:"customer"`
}

// CustomerHook represents webhook data for customers.
type CustomerHook struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// InvoiceHook represents webhook invoice data.
type InvoiceHook struct {
    ID         string  `json:"id"`
    Status     string  `json:"status"`
    CustomerID string  `json:"customer"`
    Value      float64 `json:"value"`
}

// CustomerRecord stores customer metadata in PostgreSQL.
type CustomerRecord struct {
    ID             string
    ExternalRef    string
    Name           string
    Email          string
    DocumentNumber string
    CreatedAt      time.Time
}

// PaymentRecord stores payment info locally.
type PaymentRecord struct {
    ID            string
    CustomerID    string
    ExternalRef   string
    Status        string
    Value         float64
    Description   string
    CreatedAt     time.Time
}

// SubscriptionRecord stores subscription metadata locally.
type SubscriptionRecord struct {
    ID            string
    CustomerID    string
    Status        string
    ExternalRef   string
    Value         float64
    Cycle         string
    CreatedAt     time.Time
}

// InvoiceRecord stores invoice metadata locally.
type InvoiceRecord struct {
    ID          string
    CustomerID  string
    Status      string
    Value       float64
    Description string
    CreatedAt   time.Time
}
