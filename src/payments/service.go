package payments

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Service orchestrates local persistence and remote Asaas calls.
type Service struct {
	repo   Repository
	client *AsaasClient
}

// NewService creates a payment service.
func NewService(repo Repository, client *AsaasClient) *Service {
	return &Service{repo: repo, client: client}
}

// RegisterCustomer stores a local customer and creates it in Asaas.
func (s *Service) RegisterCustomer(ctx context.Context, req CustomerRequest) (CustomerRecord, CustomerResponse, error) {
	now := time.Now().UTC()
	local := CustomerRecord{
		ID:                   generateID(),
		Name:                 req.Name,
		Email:                req.Email,
		CpfCnpj:              req.CpfCnpj,
		Phone:                req.Phone,
		MobilePhone:          req.MobilePhone,
		Address:              req.Address,
		AddressNumber:        req.AddressNumber,
		Complement:           req.Complement,
		Province:             req.Province,
		PostalCode:           req.PostalCode,
		NotificationDisabled: req.NotificationDisabled,
		AdditionalEmails:     req.AdditionalEmails,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	req.ExternalID = local.ID

	remote, err := s.client.CreateCustomer(ctx, req)
	if err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to create asaas customer: %w", err)
	}

	if err := s.repo.SaveCustomer(ctx, local); err != nil {
		return CustomerRecord{}, CustomerResponse{}, fmt.Errorf("failed to save local customer: %w", err)
	}

	return local, remote, nil
}

// CreatePayment persists the payment locally and in Asaas.
func (s *Service) CreatePayment(ctx context.Context, req PaymentRequest) (PaymentRecord, PaymentResponse, error) {
	customer, err := s.repo.FindCustomerByID(ctx, req.Customer)
	if err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to resolve customer %s: %w", req.Customer, err)
	}

	remoteCustomer, err := s.client.GetCustomer(ctx, customer.ID)
	if err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to fetch asaas customer for id %s:%w", req.Customer, err)
	}

	localID := generateID()
	req.ExternalID = localID
	asaasReq := req
	asaasReq.Customer = remoteCustomer.ID
	remote, err := s.client.CreatePayment(ctx, asaasReq)
	if err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to create asaas payment: %w", err)
	}

	callbackSuccessURL := ""
	callbackAutoRedirect := false
	if req.Callback != nil {
		callbackSuccessURL = req.Callback.SuccessURL
		callbackAutoRedirect = req.Callback.AutoRedirect
	}

	now := time.Now().UTC()
	local := PaymentRecord{
		ID:                    localID,
		CustomerID:            customer.ID,
		BillingType:           req.BillingType,
		Value:                 req.Value,
		DueDate:               parseDate(req.DueDate),
		Description:           req.Description,
		InstallmentCount:      req.InstallmentCount,
		CallbackSuccessURL:    callbackSuccessURL,
		CallbackAutoRedirect:  callbackAutoRedirect,
		Status:                remote.Status,
		InvoiceURL:            remote.InvoiceURL,
		TransactionReceiptURL: remote.TransactionReceiptURL,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.SavePayment(ctx, local); err != nil {
		return PaymentRecord{}, PaymentResponse{}, fmt.Errorf("failed to save local payment: %w", err)
	}

	return local, remote, nil
}

// CreateSubscription persists the subscription locally and remotely.
func (s *Service) CreateSubscription(ctx context.Context, req SubscriptionRequest) (SubscriptionRecord, SubscriptionResponse, error) {
	customer, err := s.repo.FindCustomerByID(ctx, req.Customer)
	if err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to resolve customer %s: %w", req.Customer, err)
	}

	remoteCustomer, err := s.client.GetCustomer(ctx, customer.ID)
	if err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to fetch asaas customer for id %s: %w", req.Customer, err)
	}

	localID := generateID()
	req.ExternalID = localID
	asaasReq := req
	asaasReq.Customer = remoteCustomer.ID
	remote, err := s.client.CreateSubscription(ctx, asaasReq)
	if err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to create asaas subscription: %w", err)
	}

	now := time.Now().UTC()
	local := SubscriptionRecord{
		ID:          localID,
		CustomerID:  customer.ID,
		BillingType: req.BillingType,
		Status:      remote.Status,
		Value:       req.Value,
		Cycle:       req.Cycle,
		NextDueDate: parseDate(req.NextDueDate),
		Description: req.Description,
		EndDate:     parseDate(req.EndDate),
		MaxPayments: req.MaxPayments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.SaveSubscription(ctx, local); err != nil {
		return SubscriptionRecord{}, SubscriptionResponse{}, fmt.Errorf("failed to save local subscription: %w", err)
	}

	return local, remote, nil
}

// CreateInvoice persists the invoice locally and in Asaas.
func (s *Service) CreateInvoice(ctx context.Context, req InvoiceRequest) (InvoiceRecord, InvoiceResponse, error) {
	payment, err := s.repo.FindPaymentByID(ctx, req.Payment)
	if err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to resolve payment %s: %w", req.Payment, err)
	}

	remotePayment, err := s.client.GetPayment(ctx, payment.ID)
	if err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to fetch asaas payment for id %s: %w", req.Payment, err)
	}

	localID := generateID()
	req.ExternalID = localID
	asaasReq := req
	asaasReq.Payment = remotePayment.ID
	remote, err := s.client.CreateInvoice(ctx, asaasReq)
	if err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to create asaas invoice: %w", err)
	}

	now := time.Now().UTC()
	local := InvoiceRecord{
		ID:                   localID,
		PaymentID:            payment.ID,
		ServiceDescription:   req.ServiceDescription,
		Observations:         req.Observations,
		Value:                req.Value,
		Deductions:           req.Deductions,
		EffectiveDate:        parseDate(req.EffectiveDate),
		MunicipalServiceID:   req.MunicipalServiceID,
		MunicipalServiceCode: req.MunicipalServiceCode,
		MunicipalServiceName: req.MunicipalServiceName,
		UpdatePayment:        req.UpdatePayment,
		TaxesRetainISS:       req.Taxes.RetainISS,
		TaxesCofins:          req.Taxes.Cofins,
		TaxesCsll:            req.Taxes.Csll,
		TaxesINSS:            req.Taxes.INSS,
		TaxesIR:              req.Taxes.IR,
		TaxesPIS:             req.Taxes.PIS,
		TaxesISS:             req.Taxes.ISS,
		Status:               remote.Status,
		PaymentLink:          remote.PaymentLink,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.repo.SaveInvoice(ctx, local); err != nil {
		return InvoiceRecord{}, InvoiceResponse{}, fmt.Errorf("failed to save local invoice: %w", err)
	}

	return local, remote, nil
}

// HandleWebhookNotification updates local records based on webhook events.
func (s *Service) HandleWebhookNotification(ctx context.Context, event NotificationEvent) error {
	switch event.Event {
	case "PAYMENT_CREATED":
		return nil
	case "PAYMENT_RECEIVED", "PAYMENT_CONFIRMED", "PAYMENT_OVERDUE":
		if event.Payment == nil {
			return fmt.Errorf("payment payload missing")
		}
		payment, err := s.repo.FindPaymentByID(ctx, event.Payment.ExternalID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		if err := s.repo.UpdatePaymentStatus(ctx, payment.ID, event.Payment.Status, event.Payment.InvoiceURL, event.Payment.TransactionReceiptURL); err != nil {
			return err
		}
		return s.issueInvoiceForPayment(ctx, payment, *event.Payment)
	case "SUBSCRIPTION_CREATED", "SUBSCRIPTION_UPDATED":
		if event.Subscription == nil {
			return fmt.Errorf("subscription payload missing")
		}
		return s.repo.UpdateSubscriptionStatus(ctx, event.Subscription.ExternalID, event.Subscription.Status)
	case "INVOICE_CREATED", "INVOICE_UPDATED", "INVOICE_OVERDUE":
		if event.Invoice == nil {
			return fmt.Errorf("invoice payload missing")
		}
		return s.repo.UpdateInvoiceStatus(ctx, event.Invoice.ExternalID, event.Invoice.Status)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Event)
	}
}

func parseDate(value string) time.Time {
	// Asaas uses yyyy-mm-dd format; parsing errors return zero time for caller validation.
	t, _ := time.Parse("2006-01-02", value)
	return t
}

// ParseDateForTests exposes parseDate for integration tests without changing production API.
func ParseDateForTests(value string) time.Time {
	return parseDate(value)
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:])
}

func (s *Service) issueInvoiceForPayment(ctx context.Context, payment PaymentRecord, payload PaymentResponse) error {
	if _, err := s.repo.FindInvoiceByPaymentID(ctx, payment.ID); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	description := payment.Description
	if description == "" {
		description = fmt.Sprintf("Serviço referente ao pagamento %s", payment.ID)
	}

	req := InvoiceRequest{
		Payment:            payment.ID,
		ServiceDescription: description,
		Observations:       "",
		Value:              payment.Value,
		Deductions:         0,
		EffectiveDate:      payment.DueDate.Format("2006-01-02"),
		MunicipalServiceID: payment.BillingType,
		MunicipalServiceName: func() string {
			if payload.BillingType != "" {
				return fmt.Sprintf("Cobrança %s", payload.BillingType)
			}
			return "Cobrança"
		}(),
		UpdatePayment: true,
		Taxes: InvoiceTaxes{
			RetainISS: false,
			Cofins:    0,
			Csll:      0,
			INSS:      0,
			IR:        0,
			PIS:       0,
			ISS:       0,
		},
	}

	if _, _, err := s.CreateInvoice(ctx, req); err != nil {
		return fmt.Errorf("failed to issue invoice for payment %s: %w", payment.ID, err)
	}

	return nil
}
