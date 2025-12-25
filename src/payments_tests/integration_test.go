package payments_tests

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"asaas/src/payments"
)

func TestPaymentLifecycleAndWebhooks(t *testing.T) {
	ctx := context.Background()
	db := newFakeDatabase()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/customers":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payments.CustomerResponse{ID: "cus_1", ExternalReference: "customer-ref"})
		case "/v3/payments":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payments.PaymentResponse{ID: "pay_1", Status: "CONFIRMED", ExternalReference: "payment-ref"})
		case "/v3/subscriptions":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payments.SubscriptionResponse{ID: "sub_1", Status: "ACTIVE", ExternalReference: "subscription-ref"})
		case "/v3/invoices":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payments.InvoiceResponse{ID: "inv_1", Status: "ISSUED", ExternalReference: "invoice-ref"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	os.Setenv("ASAAS_API_URL", apiServer.URL)
	os.Setenv("ASAAS_API_TOKEN", "token")

	client, err := payments.NewClientWithDatabase(ctx, db, payments.WithBaseURL(apiServer.URL), payments.WithHTTPClient(apiServer.Client()))
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}

	customerPayload := payments.CustomerRequest{Name: "Alice", Email: "alice@example.com", CPF: "12345678900"}
	customerRecord, apiCustomer, err := client.CreateCustomer(ctx, customerPayload)
	if err != nil {
		t.Fatalf("customer creation failed: %v", err)
	}
	if apiCustomer.ID != "cus_1" || customerRecord.ID == 0 {
		t.Fatalf("unexpected customer response: %+v", apiCustomer)
	}

	paymentPayload := payments.PaymentRequest{BillingType: "CREDIT_CARD", Value: 100.0, DueDate: "2024-01-10"}
	paymentRecord, apiPayment, err := client.CreatePayment(ctx, *customerRecord, paymentPayload)
	if err != nil {
		t.Fatalf("payment creation failed: %v", err)
	}
	if apiPayment.Status != "CONFIRMED" || paymentRecord.ExternalID != "pay_1" {
		t.Fatalf("unexpected payment response")
	}

	subscriptionPayload := payments.SubscriptionRequest{BillingType: "BOLETO", Value: 200.0, Cycle: "MONTHLY", NextDueDate: "2024-02-01"}
	subRecord, apiSubscription, err := client.CreateSubscription(ctx, *customerRecord, subscriptionPayload)
	if err != nil {
		t.Fatalf("subscription creation failed: %v", err)
	}
	if apiSubscription.Status != "ACTIVE" || subRecord.ExternalID != "sub_1" {
		t.Fatalf("unexpected subscription response")
	}

	invoicePayload := payments.InvoiceRequest{Value: 300.0, Description: "Service", DueDate: "2024-03-01"}
	invoiceRecord, apiInvoice, err := client.CreateInvoice(ctx, *customerRecord, invoicePayload)
	if err != nil {
		t.Fatalf("invoice creation failed: %v", err)
	}
	if apiInvoice.Status != "ISSUED" || invoiceRecord.ExternalID != "inv_1" {
		t.Fatalf("unexpected invoice response")
	}

	webhookBody, _ := json.Marshal(payments.WebhookEvent{Event: "PAYMENT_CONFIRMED", Payment: &payments.PaymentResponse{ID: "pay_1", Status: "CONFIRMED"}})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(webhookBody))
	rr := httptest.NewRecorder()
	client.WebhookHandler(ctx, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected webhook status: %d", rr.Code)
	}

	if !db.allOperationsSucceeded() {
		t.Fatalf("database expectations not satisfied")
	}
}

// --- Test fakes ---

type fakeDatabase struct {
	inserted map[string]int64
	updates  []string
}

func newFakeDatabase() *fakeDatabase {
	return &fakeDatabase{inserted: map[string]int64{"customers": 1, "payments": 2, "subscriptions": 3, "invoices": 4}}
}

func (f *fakeDatabase) PingContext(ctx context.Context) error { return nil }

func (f *fakeDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (payments.DBTransaction, error) {
	return &fakeTx{db: f}, nil
}

func (f *fakeDatabase) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	f.updates = append(f.updates, query)
	return stubResult(1), nil
}

func (f *fakeDatabase) allOperationsSucceeded() bool {
	return len(f.updates) >= 4 // updates triggered for customers, payments, subscriptions, invoices, webhook
}

type fakeTx struct {
	db *fakeDatabase
}

func (t *fakeTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	t.db.updates = append(t.db.updates, query)
	return stubResult(1), nil
}

func (t *fakeTx) QueryRowContext(ctx context.Context, query string, args ...any) payments.RowScanner {
	switch {
	case strings.Contains(query, "customers"):
		return stubRow{id: t.db.inserted["customers"]}
	case strings.Contains(query, "payments"):
		return stubRow{id: t.db.inserted["payments"]}
	case strings.Contains(query, "subscriptions"):
		return stubRow{id: t.db.inserted["subscriptions"]}
	case strings.Contains(query, "invoices"):
		return stubRow{id: t.db.inserted["invoices"]}
	default:
		return stubRow{id: 0, err: driver.ErrBadConn}
	}
}

func (t *fakeTx) Commit() error   { return nil }
func (t *fakeTx) Rollback() error { return nil }

// stubRow satisfies payments.RowScanner.
type stubRow struct {
	id  int64
	err error
}

func (r stubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) == 0 {
		return nil
	}
	if ptr, ok := dest[0].(*int64); ok {
		*ptr = r.id
	}
	return nil
}

type stubResult int64

func (r stubResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r stubResult) RowsAffected() (int64, error) { return 1, nil }
