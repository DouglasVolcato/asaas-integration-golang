package payments_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"asaas/src/payments"
	"asaas/src/payments/repository"
)

type apiResponse struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name,omitempty"`
	Email              string  `json:"email,omitempty"`
	CpfCnpj            string  `json:"cpfCnpj,omitempty"`
	MobilePhone        string  `json:"mobilePhone,omitempty"`
	Phone              string  `json:"phone,omitempty"`
	Customer           string  `json:"customer,omitempty"`
	BillingType        string  `json:"billingType,omitempty"`
	Value              float64 `json:"value,omitempty"`
	DueDate            string  `json:"dueDate,omitempty"`
	Status             string  `json:"status,omitempty"`
	ExternalReference  string  `json:"externalReference,omitempty"`
	Cycle              string  `json:"cycle,omitempty"`
	NextDueDate        string  `json:"nextDueDate,omitempty"`
	ServiceDescription string  `json:"serviceDescription,omitempty"`
}

func TestServiceIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v3/customers":
			json.NewEncoder(w).Encode(apiResponse{ID: "cus_123", Name: "Jane Doe", Email: "jane@example.com", CpfCnpj: "12345678901", Phone: "1133224455", MobilePhone: "1199998888", ExternalReference: "local-customer"})
		case "/v3/payments":
			json.NewEncoder(w).Encode(apiResponse{ID: "pay_123", Customer: "cus_123", BillingType: "PIX", Value: 150.5, DueDate: "2024-11-10", Status: "PENDING", ExternalReference: "local-payment"})
		case "/v3/subscriptions":
			json.NewEncoder(w).Encode(apiResponse{ID: "sub_123", Customer: "cus_123", BillingType: "BOLETO", Value: 200, Cycle: "MONTHLY", NextDueDate: "2024-11-15", Status: "ACTIVE", ExternalReference: "local-subscription"})
		case "/v3/invoices":
			json.NewEncoder(w).Encode(apiResponse{ID: "inv_123", Customer: "cus_123", ServiceDescription: "Consulting", Value: 300, Status: "PENDING", ExternalReference: "local-invoice"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := payments.Config{APIBaseURL: server.URL, APIToken: "token"}
	client := payments.NewAsaasClient(cfg, server.Client())
	repo := repository.NewMemoryRepository()
	service := payments.NewService(client, repo)

	ctx := context.Background()

	customer, err := service.CreateCustomer(ctx, payments.CustomerCreateRequest{Name: "Jane Doe", Email: "jane@example.com", CpfCnpj: "12345678901"}, "customer-local")
	if err != nil {
		t.Fatalf("CreateCustomer failed: %v", err)
	}
	if customer.ID != "cus_123" {
		t.Fatalf("unexpected customer id: %s", customer.ID)
	}

	payment, err := service.CreatePayment(ctx, payments.PaymentCreateRequest{Customer: customer.ID, BillingType: payments.BillingTypePix, Value: 150.5, DueDate: "2024-11-10", ExternalReference: "local-payment"}, "payment-local", "customer-local")
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}

	subscription, err := service.CreateSubscription(ctx, payments.SubscriptionCreateRequest{Customer: customer.ID, BillingType: payments.BillingTypeBoleto, Value: 200, Cycle: payments.SubscriptionCycleMonthly, NextDueDate: "2024-11-15", ExternalReference: "local-subscription"}, "subscription-local", "customer-local")
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	invoice, err := service.CreateInvoice(ctx, payments.InvoiceCreateRequest{Customer: customer.ID, ServiceDescription: "Consulting", Value: 300, DueDate: "2024-11-30", ExternalReference: "local-invoice"}, "invoice-local", "customer-local")
	if err != nil {
		t.Fatalf("CreateInvoice failed: %v", err)
	}

	if payment.ID == "" || subscription.ID == "" || invoice.ID == "" {
		t.Fatalf("expected ids to be filled")
	}

	webhook := payments.NewWebhookHandler(service, "secret")
	event := payments.WebhookEvent{Event: "PAYMENT_CONFIRMED", Payment: &payments.PaymentResponse{ID: payment.ID, Status: "RECEIVED"}}
	payload, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-Asaas-Signature", "secret")
	rr := httptest.NewRecorder()
	webhook.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("webhook handler returned %d", rr.Code)
	}

	invoiceEvent := payments.WebhookEvent{Event: "INVOICE_AUTHORIZED", Invoice: &payments.InvoiceResponse{ID: invoice.ID, Status: "AUTHORIZED"}}
	invoicePayload, _ := json.Marshal(invoiceEvent)
	reqInvoice := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(invoicePayload))
	reqInvoice.Header.Set("X-Asaas-Signature", "secret")
	rrInvoice := httptest.NewRecorder()
	webhook.ServeHTTP(rrInvoice, reqInvoice)
	if rrInvoice.Code != http.StatusOK {
		t.Fatalf("invoice webhook handler returned %d", rrInvoice.Code)
	}

	// Validate repository state
	if repo.PaymentsSnapshot()[payment.ID].Status != "RECEIVED" {
		t.Fatalf("expected payment status to be updated")
	}
	if repo.InvoicesSnapshot()[invoice.ID].Status != "AUTHORIZED" {
		t.Fatalf("expected invoice status to be updated")
	}
}
