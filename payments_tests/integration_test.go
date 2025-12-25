package payments_tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"asaas/src/payments"
)

type testBackend struct {
	server *httptest.Server
}

func newTestBackend(t *testing.T) (*testBackend, payments.Config) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if http.MethodPost != r.Method {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		var req payments.CustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode customer: %v", err)
		}
		_ = r.Body.Close()
		resp := payments.CustomerResponse{ID: "cus_123", Name: req.Name, Email: req.Email, ExternalID: "cus_123"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		if http.MethodPost != r.Method {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		var req payments.PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode payment: %v", err)
		}
		_ = r.Body.Close()
		resp := payments.PaymentResponse{ID: "pay_123", Customer: req.Customer, BillingType: req.BillingType, Value: req.Value, Status: "CONFIRMED", ExternalID: "pay_123"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if http.MethodPost != r.Method {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		var req payments.SubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode subscription: %v", err)
		}
		_ = r.Body.Close()
		resp := payments.SubscriptionResponse{ID: "sub_123", Customer: req.Customer, Status: "ACTIVE", Value: req.Value, ExternalID: "sub_123"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/invoices", func(w http.ResponseWriter, r *http.Request) {
		if http.MethodPost != r.Method {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		var req payments.InvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode invoice: %v", err)
		}
		_ = r.Body.Close()
		resp := payments.InvoiceResponse{ID: "inv_123", Customer: req.Customer, Status: "PENDING", Value: req.Value, ExternalID: "inv_123", PaymentLink: "http://example.com"}
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	cfg := payments.Config{APIURL: server.URL, APIToken: "token"}

	t.Cleanup(server.Close)
	return &testBackend{server: server}, cfg
}

func TestFullFlow(t *testing.T) {
	ctx := context.Background()
	backend, cfg := newTestBackend(t)
	_ = backend

	client := payments.NewAsaasClient(cfg)
	repo := payments.NewInMemoryRepository()
	service := payments.NewService(repo, client)

	customerLocal, customerRemote, err := service.RegisterCustomer(ctx, payments.CustomerRequest{Name: "Test User", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("register customer: %v", err)
	}
	if customerRemote.ID != "cus_123" {
		t.Fatalf("unexpected customer id: %s", customerRemote.ID)
	}
	if customerLocal.ExternalID != "cus_123" {
		t.Fatalf("unexpected local external id: %s", customerLocal.ExternalID)
	}

	payReq := payments.PaymentRequest{Customer: customerRemote.ID, BillingType: "BOLETO", Value: 10.5, DueDate: "2024-12-01"}
	paymentLocal, paymentRemote, err := service.CreatePayment(ctx, payReq, customerLocal.ID)
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if paymentRemote.ID != "pay_123" {
		t.Fatalf("unexpected payment id: %s", paymentRemote.ID)
	}
	if paymentLocal.Status != "CONFIRMED" {
		t.Fatalf("unexpected payment status: %s", paymentLocal.Status)
	}

	subReq := payments.SubscriptionRequest{Customer: customerRemote.ID, BillingType: "CREDIT_CARD", Value: 100, NextDueDate: "2024-12-01", Cycle: "MONTHLY"}
	subLocal, subRemote, err := service.CreateSubscription(ctx, subReq, customerLocal.ID)
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}
	if subRemote.ID != "sub_123" {
		t.Fatalf("unexpected subscription id: %s", subRemote.ID)
	}
	if subLocal.Status != "ACTIVE" {
		t.Fatalf("unexpected subscription status: %s", subLocal.Status)
	}

	invReq := payments.InvoiceRequest{Customer: customerRemote.ID, Value: 15.75, DueDate: "2024-12-10", Description: "Test"}
	invLocal, invRemote, err := service.CreateInvoice(ctx, invReq, customerLocal.ID)
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if invRemote.ID != "inv_123" {
		t.Fatalf("unexpected invoice id: %s", invRemote.ID)
	}
	if invLocal.Status != "PENDING" {
		t.Fatalf("unexpected invoice status: %s", invLocal.Status)
	}

	webhook := payments.NewWebhookHandler(service)
	event := payments.NotificationEvent{Event: "PAYMENT_RECEIVED", Payment: &payments.PaymentResponse{ExternalID: paymentLocal.ID, Status: "RECEIVED"}}
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal webhook: %v", err)
	}
	if err := webhook.WithContext(ctx, body); err != nil {
		t.Fatalf("handle webhook: %v", err)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("ASAAS_API_URL", "http://example.com")
	os.Setenv("ASAAS_API_TOKEN", "token")
	cfg, err := payments.LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.APIURL != "http://example.com" || cfg.APIToken != "token" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestInvalidWebhook(t *testing.T) {
	repo := payments.NewInMemoryRepository()
	client := payments.NewAsaasClient(payments.Config{APIURL: "http://localhost", APIToken: "token"})
	service := payments.NewService(repo, client)
	handler := payments.NewWebhookHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/webhook", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rr.Code)
	}
}

func TestParseDate(t *testing.T) {
	parsed := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	result := payments.ParseDateForTests("2024-12-01")
	if !result.Equal(parsed) {
		t.Fatalf("unexpected parsed date: %v", result)
	}
}
