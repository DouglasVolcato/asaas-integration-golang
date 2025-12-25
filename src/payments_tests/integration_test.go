//go:build integration
// +build integration

package payments_tests

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "testing"
    "time"

    "asaas/src/payments"

    _ "github.com/jackc/pgx/v5/stdlib"
)

// Integration test that hits the real Asaas API when credentials are provided.
func TestAsaasIntegration(t *testing.T) {
    apiURL := os.Getenv("ASAAS_API_URL")
    apiToken := os.Getenv("ASAAS_API_TOKEN")
    dsn := os.Getenv("POSTGRES_DSN")

    if apiURL == "" || apiToken == "" || dsn == "" {
        t.Skip("ASAAS_API_URL, ASAAS_API_TOKEN and POSTGRES_DSN must be set for integration tests")
    }

    ctx := context.Background()

    cfg, err := payments.LoadConfig()
    if err != nil {
        t.Fatalf("load config: %v", err)
    }

    db, err := sql.Open("pgx", dsn)
    if err != nil {
        t.Fatalf("open database: %v", err)
    }
    defer db.Close()

    repo := payments.NewRepository(db)
    if err := repo.InitSchema(ctx); err != nil {
        t.Fatalf("init schema: %v", err)
    }

    service := payments.NewService(payments.NewClient(cfg), repo)

    customerPayload := payments.Customer{
        Name:        fmt.Sprintf("Integration Client %d", time.Now().Unix()),
        Email:       "integration@example.com",
        CPF:         "12345678909",
        ExternalRef: fmt.Sprintf("ext-%d", time.Now().UnixNano()),
    }

    customerResp, err := service.CreateCustomer(ctx, customerPayload)
    if err != nil {
        t.Fatalf("create customer: %v", err)
    }

    paymentPayload := payments.Payment{
        CustomerID:        customerResp.ID,
        BillingType:       "BOLETO",
        Value:             100.50,
        DueDate:           time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
        Description:       "Integration payment",
        ExternalReference: fmt.Sprintf("pay-%d", time.Now().UnixNano()),
    }

    paymentResp, err := service.CreatePayment(ctx, paymentPayload)
    if err != nil {
        t.Fatalf("create payment: %v", err)
    }

    subscriptionPayload := payments.Subscription{
        CustomerID:        customerResp.ID,
        BillingType:       "BOLETO",
        Value:             50.0,
        NextDueDate:       time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
        Cycle:             "MONTHLY",
        Description:       "Integration subscription",
        ExternalReference: fmt.Sprintf("sub-%d", time.Now().UnixNano()),
    }

    if _, err := service.CreateSubscription(ctx, subscriptionPayload); err != nil {
        t.Fatalf("create subscription: %v", err)
    }

    invoicePayload := payments.Invoice{
        CustomerID:        customerResp.ID,
        BillingType:       "BOLETO",
        Value:             10.0,
        DueDate:           time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
        Description:       "Integration invoice",
        ExternalReference: fmt.Sprintf("inv-%d", time.Now().UnixNano()),
    }

    if _, err := service.CreateInvoice(ctx, invoicePayload); err != nil {
        t.Fatalf("create invoice: %v", err)
    }

    if _, err := service.GetCustomer(ctx, customerResp.ID); err != nil {
        t.Fatalf("fetch stored customer: %v", err)
    }

    if _, err := service.GetPayment(ctx, paymentResp.ID); err != nil {
        t.Fatalf("fetch stored payment: %v", err)
    }
}
