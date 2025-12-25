package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

// AsaasClient handles authenticated HTTP communication with the Asaas API.
type AsaasClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewAsaasClient creates an AsaasClient using the provided configuration.
func NewAsaasClient(cfg Config) *AsaasClient {
	return &AsaasClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    cfg.APIURL,
		token:      cfg.APIToken,
	}
}

// CustomerRequest represents the payload for creating a customer in Asaas.
type CustomerRequest struct {
	Name                 string `json:"name"`
	Email                string `json:"email,omitempty"`
	CPF                  string `json:"cpf,omitempty"`
	CNPJ                 string `json:"cnpj,omitempty"`
	Phone                string `json:"phone,omitempty"`
	MobilePhone          string `json:"mobilePhone,omitempty"`
	Address              string `json:"address,omitempty"`
	AddressNumber        string `json:"addressNumber,omitempty"`
	Complement           string `json:"complement,omitempty"`
	Province             string `json:"province,omitempty"`
	PostalCode           string `json:"postalCode,omitempty"`
	ExternalID           string `json:"externalReference,omitempty"`
	NotificationDisabled bool   `json:"notificationDisabled,omitempty"`
	AdditionalEmails     string `json:"additionalEmails,omitempty"`
}

// CustomerResponse is the subset of Asaas response used by this module.
type CustomerResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ExternalID string `json:"externalReference"`
}

// PaymentRequest represents the payload for creating a payment.
type PaymentRequest struct {
	Customer         string  `json:"customer"`
	BillingType      string  `json:"billingType"`
	Value            float64 `json:"value"`
	DueDate          string  `json:"dueDate"`
	Description      string  `json:"description,omitempty"`
	InstallmentCount int     `json:"installmentCount,omitempty"`
	ExternalID       string  `json:"externalReference,omitempty"`
	CallbackURL      string  `json:"callbackUrl,omitempty"`
}

// PaymentResponse represents the relevant payment details returned by Asaas.
type PaymentResponse struct {
	ID                    string  `json:"id"`
	Customer              string  `json:"customer"`
	BillingType           string  `json:"billingType"`
	Value                 float64 `json:"value"`
	Status                string  `json:"status"`
	ExternalID            string  `json:"externalReference"`
	InvoiceURL            string  `json:"invoiceUrl,omitempty"`
	TransactionReceiptURL string  `json:"transactionReceiptUrl,omitempty"`
}

// SubscriptionRequest represents creation of an Asaas subscription.
type SubscriptionRequest struct {
	Customer    string  `json:"customer"`
	BillingType string  `json:"billingType"`
	Value       float64 `json:"value"`
	NextDueDate string  `json:"nextDueDate"`
	Cycle       string  `json:"cycle"`
	ExternalID  string  `json:"externalReference,omitempty"`
	Description string  `json:"description,omitempty"`
	EndDate     string  `json:"endDate,omitempty"`
	MaxPayments int     `json:"maxPayments,omitempty"`
}

// SubscriptionResponse captures required subscription fields.
type SubscriptionResponse struct {
	ID         string  `json:"id"`
	Customer   string  `json:"customer"`
	Status     string  `json:"status"`
	Value      float64 `json:"value"`
	ExternalID string  `json:"externalReference"`
}

// InvoiceRequest represents the payload to create an invoice in Asaas.
type InvoiceRequest struct {
	Customer    string  `json:"customer"`
	Value       float64 `json:"value"`
	DueDate     string  `json:"dueDate"`
	Description string  `json:"description,omitempty"`
	ExternalID  string  `json:"externalReference,omitempty"`
	PaymentLink bool    `json:"paymentLink,omitempty"`
}

// InvoiceResponse captures invoice fields from Asaas.
type InvoiceResponse struct {
	ID          string  `json:"id"`
	Customer    string  `json:"customer"`
	Status      string  `json:"status"`
	Value       float64 `json:"value"`
	ExternalID  string  `json:"externalReference"`
	PaymentLink string  `json:"paymentLink"`
}

// NotificationEvent represents webhook payloads sent by Asaas.
type NotificationEvent struct {
	Event        string                `json:"event"`
	Payment      *PaymentResponse      `json:"payment,omitempty"`
	Invoice      *InvoiceResponse      `json:"invoice,omitempty"`
	Subscription *SubscriptionResponse `json:"subscription,omitempty"`
}

func (c *AsaasClient) doRequest(ctx context.Context, method, endpoint string, payload any, v any) error {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	base.Path = path.Join(base.Path, endpoint)

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, base.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("access_token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("asaas error %d: %s", resp.StatusCode, string(respBody))
	}

	if v == nil {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// CreateCustomer sends a request to create a customer.
func (c *AsaasClient) CreateCustomer(ctx context.Context, req CustomerRequest) (CustomerResponse, error) {
	var resp CustomerResponse
	err := c.doRequest(ctx, http.MethodPost, "customers", req, &resp)
	return resp, err
}

// GetCustomer retrieves a customer.
func (c *AsaasClient) GetCustomer(ctx context.Context, id string) (CustomerResponse, error) {
	var resp CustomerResponse
	endpoint := path.Join("customers", id)
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp)
	return resp, err
}

// CreatePayment creates a payment for a customer.
func (c *AsaasClient) CreatePayment(ctx context.Context, req PaymentRequest) (PaymentResponse, error) {
	var resp PaymentResponse
	err := c.doRequest(ctx, http.MethodPost, "payments", req, &resp)
	return resp, err
}

// GetPayment retrieves a payment by id.
func (c *AsaasClient) GetPayment(ctx context.Context, id string) (PaymentResponse, error) {
	var resp PaymentResponse
	endpoint := path.Join("payments", id)
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp)
	return resp, err
}

// CreateSubscription creates a recurring subscription.
func (c *AsaasClient) CreateSubscription(ctx context.Context, req SubscriptionRequest) (SubscriptionResponse, error) {
	var resp SubscriptionResponse
	err := c.doRequest(ctx, http.MethodPost, "subscriptions", req, &resp)
	return resp, err
}

// CancelSubscription cancels a subscription in Asaas.
func (c *AsaasClient) CancelSubscription(ctx context.Context, id string) error {
	endpoint := path.Join("subscriptions", id, "cancel")
	return c.doRequest(ctx, http.MethodPost, endpoint, nil, nil)
}

// CreateInvoice creates an invoice for a customer.
func (c *AsaasClient) CreateInvoice(ctx context.Context, req InvoiceRequest) (InvoiceResponse, error) {
	var resp InvoiceResponse
	err := c.doRequest(ctx, http.MethodPost, "invoices", req, &resp)
	return resp, err
}

// GetInvoice retrieves invoice details.
func (c *AsaasClient) GetInvoice(ctx context.Context, id string) (InvoiceResponse, error) {
	var resp InvoiceResponse
	endpoint := path.Join("invoices", id)
	err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &resp)
	return resp, err
}
