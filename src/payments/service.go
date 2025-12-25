package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CreateCustomer stores customer data locally and registers it in Asaas using the local ID as external reference.
func (c *Client) CreateCustomer(ctx context.Context, payload CustomerRequest) (*CustomerRecord, *CustomerResponse, error) {
	now := time.Now().UTC()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	externalRef := payload.ExternalReference
	if externalRef == "" {
		externalRef = fmt.Sprintf("customer-%d", time.Now().UnixNano())
	}

	var customerID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO customers (name, email, document, external_reference, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		payload.Name, payload.Email, firstNonEmpty(payload.CPF, payload.CNPJ), externalRef, now, now).Scan(&customerID); err != nil {
		return nil, nil, err
	}

	payload.ExternalReference = externalRef

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.baseRequest(ctx, http.MethodPost, "/v3/customers", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("asaas returned status %d", resp.StatusCode)
	}

	var apiResp CustomerResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE customers SET external_id=$1, updated_at=$2 WHERE id=$3`, apiResp.ID, time.Now().UTC(), customerID); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	record := &CustomerRecord{
		ID:         customerID,
		ExternalID: apiResp.ID,
		Name:       payload.Name,
		Email:      payload.Email,
		Document:   firstNonEmpty(payload.CPF, payload.CNPJ),
		CreatedAt:  now,
		UpdatedAt:  time.Now().UTC(),
	}

	return record, &apiResp, nil
}

// CreatePayment stores a payment locally and registers it in Asaas linked to the given customer external ID.
func (c *Client) CreatePayment(ctx context.Context, customerRecord CustomerRecord, payload PaymentRequest) (*PaymentRecord, *PaymentResponse, error) {
	now := time.Now().UTC()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	externalRef := payload.ExternalReference
	if externalRef == "" {
		externalRef = fmt.Sprintf("payment-%d", time.Now().UnixNano())
	}

	var paymentID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO payments (customer_id, external_reference, billing_type, value, status, due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		customerRecord.ID, externalRef, payload.BillingType, payload.Value, "PENDING", parseDate(payload.DueDate), now, now).Scan(&paymentID); err != nil {
		return nil, nil, err
	}

	payload.CustomerID = customerRecord.ExternalID
	payload.ExternalReference = externalRef

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.baseRequest(ctx, http.MethodPost, "/v3/payments", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("asaas returned status %d", resp.StatusCode)
	}

	var apiResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE payments SET external_id=$1, status=$2, updated_at=$3 WHERE id=$4`, apiResp.ID, apiResp.Status, time.Now().UTC(), paymentID); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	record := &PaymentRecord{
		ID:                paymentID,
		CustomerID:        customerRecord.ID,
		ExternalID:        apiResp.ID,
		ExternalReference: payload.ExternalReference,
		BillingType:       payload.BillingType,
		Value:             payload.Value,
		Status:            apiResp.Status,
		DueDate:           parseDate(payload.DueDate),
		CreatedAt:         now,
		UpdatedAt:         time.Now().UTC(),
	}

	return record, &apiResp, nil
}

// CreateSubscription stores subscription data locally before registering it with Asaas.
func (c *Client) CreateSubscription(ctx context.Context, customerRecord CustomerRecord, payload SubscriptionRequest) (*SubscriptionRecord, *SubscriptionResponse, error) {
	now := time.Now().UTC()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	externalRef := payload.ExternalReference
	if externalRef == "" {
		externalRef = fmt.Sprintf("subscription-%d", time.Now().UnixNano())
	}

	var subscriptionID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO subscriptions (customer_id, external_reference, billing_type, value, cycle, status, next_due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		customerRecord.ID, externalRef, payload.BillingType, payload.Value, payload.Cycle, "PENDING", parseDate(payload.NextDueDate), now, now).Scan(&subscriptionID); err != nil {
		return nil, nil, err
	}

	payload.CustomerID = customerRecord.ExternalID
	payload.ExternalReference = externalRef

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.baseRequest(ctx, http.MethodPost, "/v3/subscriptions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("asaas returned status %d", resp.StatusCode)
	}

	var apiResp SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE subscriptions SET external_id=$1, status=$2, updated_at=$3 WHERE id=$4`, apiResp.ID, apiResp.Status, time.Now().UTC(), subscriptionID); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	record := &SubscriptionRecord{
		ID:                subscriptionID,
		CustomerID:        customerRecord.ID,
		ExternalID:        apiResp.ID,
		ExternalReference: payload.ExternalReference,
		BillingType:       payload.BillingType,
		Value:             payload.Value,
		Cycle:             payload.Cycle,
		Status:            apiResp.Status,
		NextDueDate:       parseDate(payload.NextDueDate),
		CreatedAt:         now,
		UpdatedAt:         time.Now().UTC(),
	}

	return record, &apiResp, nil
}

// CreateInvoice stores an invoice locally and registers it with Asaas.
func (c *Client) CreateInvoice(ctx context.Context, customerRecord CustomerRecord, payload InvoiceRequest) (*InvoiceRecord, *InvoiceResponse, error) {
	now := time.Now().UTC()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	externalRef := payload.ExternalReference
	if externalRef == "" {
		externalRef = fmt.Sprintf("invoice-%d", time.Now().UnixNano())
	}

	var invoiceID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO invoices (customer_id, external_reference, value, status, due_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		customerRecord.ID, externalRef, payload.Value, "PENDING", parseDate(payload.DueDate), now, now).Scan(&invoiceID); err != nil {
		return nil, nil, err
	}

	payload.CustomerID = customerRecord.ExternalID
	payload.ExternalReference = externalRef

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.baseRequest(ctx, http.MethodPost, "/v3/invoices", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("asaas returned status %d", resp.StatusCode)
	}

	var apiResp InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE invoices SET external_id=$1, status=$2, updated_at=$3 WHERE id=$4`, apiResp.ID, apiResp.Status, time.Now().UTC(), invoiceID); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	record := &InvoiceRecord{
		ID:                invoiceID,
		CustomerID:        customerRecord.ID,
		ExternalID:        apiResp.ID,
		ExternalReference: payload.ExternalReference,
		Number:            apiResp.Number,
		Status:            apiResp.Status,
		DueDate:           parseDate(payload.DueDate),
		CreatedAt:         now,
		UpdatedAt:         time.Now().UTC(),
	}

	return record, &apiResp, nil
}

// UpdatePaymentStatus persists payment status changes triggered by webhook notifications.
func (c *Client) UpdatePaymentStatus(ctx context.Context, paymentExternalID, status string) error {
	_, err := c.db.ExecContext(ctx, `UPDATE payments SET status=$1, updated_at=$2 WHERE external_id=$3`, status, time.Now().UTC(), paymentExternalID)
	return err
}

// UpdateSubscriptionStatus persists subscription status changes triggered by webhook notifications.
func (c *Client) UpdateSubscriptionStatus(ctx context.Context, subscriptionExternalID, status string) error {
	_, err := c.db.ExecContext(ctx, `UPDATE subscriptions SET status=$1, updated_at=$2 WHERE external_id=$3`, status, time.Now().UTC(), subscriptionExternalID)
	return err
}

// UpdateInvoiceStatus persists invoice status changes triggered by webhook notifications.
func (c *Client) UpdateInvoiceStatus(ctx context.Context, invoiceExternalID, status string) error {
	_, err := c.db.ExecContext(ctx, `UPDATE invoices SET status=$1, updated_at=$2 WHERE external_id=$3`, status, time.Now().UTC(), invoiceExternalID)
	return err
}

// extractInsertID reads the primary key from an INSERT RETURNING statement.
// firstNonEmpty returns the first non-empty string in the provided arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseDate converts date strings to time.Time using the ISO layout expected by Asaas.
func parseDate(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ParseDateForTest exposes date parsing for integration tests without exporting internal helpers broadly.
func ParseDateForTest(value string) time.Time {
	return parseDate(value)
}
