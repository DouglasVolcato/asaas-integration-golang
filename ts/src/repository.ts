import { Pool } from 'pg';
import {
  CustomerRecord,
  InvoiceRecord,
  PaymentRecord,
  SubscriptionRecord,
} from './types';

export class PostgresRepository {
  constructor(private pool: Pool) {}

  async ensureSchema(): Promise<void> {
    const stmts = [
      `CREATE TABLE IF NOT EXISTS payment_customers (
        id UUID PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT DEFAULT '',
        cpfCnpj TEXT DEFAULT '',
        phone TEXT DEFAULT '',
        mobile_phone TEXT DEFAULT '',
        address TEXT DEFAULT '',
        address_number TEXT DEFAULT '',
        complement TEXT DEFAULT '',
        province TEXT DEFAULT '',
        postal_code TEXT DEFAULT '',
        notification_disabled BOOLEAN NOT NULL DEFAULT FALSE,
        additional_emails TEXT DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
      );`,
      `CREATE TABLE IF NOT EXISTS payment_payments (
        id UUID PRIMARY KEY,
        customer_id UUID NOT NULL REFERENCES payment_customers(id),
        subscription_id UUID DEFAULT NULL,
        billing_type TEXT NOT NULL,
        value NUMERIC NOT NULL,
        due_date TIMESTAMPTZ NOT NULL,
        description TEXT DEFAULT '',
        installment_count INTEGER NOT NULL DEFAULT 0,
        callback_success_url TEXT DEFAULT '',
        callback_auto_redirect BOOLEAN NOT NULL DEFAULT FALSE,
        status TEXT DEFAULT '',
        invoice_url TEXT DEFAULT '',
        transaction_receipt_url TEXT DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
      );`,
      `ALTER TABLE payment_payments ADD COLUMN IF NOT EXISTS subscription_id UUID;`,
      `CREATE TABLE IF NOT EXISTS payment_subscriptions (
        id UUID PRIMARY KEY,
        customer_id UUID NOT NULL REFERENCES payment_customers(id),
        billing_type TEXT NOT NULL,
        status TEXT DEFAULT '',
        value NUMERIC NOT NULL,
        cycle TEXT NOT NULL,
        next_due_date TIMESTAMPTZ NOT NULL,
        description TEXT DEFAULT '',
        end_date TIMESTAMPTZ,
        max_payments INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
      );`,
      `CREATE TABLE IF NOT EXISTS payment_invoices (
        id UUID PRIMARY KEY,
        payment_id UUID NOT NULL REFERENCES payment_payments(id),
        service_description TEXT NOT NULL,
        observations TEXT NOT NULL,
        value NUMERIC NOT NULL,
        deductions NUMERIC NOT NULL DEFAULT 0,
        effective_date TIMESTAMPTZ NOT NULL,
        municipal_service_id TEXT DEFAULT '',
        municipal_service_code TEXT DEFAULT '',
        municipal_service_name TEXT NOT NULL,
        update_payment BOOLEAN NOT NULL DEFAULT FALSE,
        taxes_retain_iss BOOLEAN NOT NULL DEFAULT FALSE,
        taxes_cofins NUMERIC NOT NULL DEFAULT 0,
        taxes_csll NUMERIC NOT NULL DEFAULT 0,
        taxes_inss NUMERIC NOT NULL DEFAULT 0,
        taxes_ir NUMERIC NOT NULL DEFAULT 0,
        taxes_pis NUMERIC NOT NULL DEFAULT 0,
        taxes_iss NUMERIC NOT NULL DEFAULT 0,
        status TEXT DEFAULT '',
        payment_link TEXT DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
      );`,
    ];

    for (const stmt of stmts) {
      await this.pool.query(stmt);
    }
  }

  async saveCustomer(customer: CustomerRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_customers (
        id, name, email, cpfCnpj, phone, mobile_phone, address, address_number, complement, province, postal_code, notification_disabled, additional_emails, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
      [
        customer.id,
        customer.name,
        customer.email ?? '',
        customer.cpfCnpj ?? '',
        customer.phone ?? '',
        customer.mobilePhone ?? '',
        customer.address ?? '',
        customer.addressNumber ?? '',
        customer.complement ?? '',
        customer.province ?? '',
        customer.postalCode ?? '',
        customer.notificationDisabled ?? false,
        customer.additionalEmails ?? '',
        customer.createdAt,
        customer.updatedAt,
      ],
    );
  }

  async findCustomerById(id: string): Promise<CustomerRecord | null> {
    const { rows } = await this.pool.query<CustomerRecord>(
      `SELECT id, name, email, cpfCnpj as "cpfCnpj", phone, mobile_phone as "mobilePhone", address, address_number as "addressNumber", complement, province, postal_code as "postalCode", notification_disabled as "notificationDisabled", additional_emails as "additionalEmails", created_at as "createdAt", updated_at as "updatedAt" FROM payment_customers WHERE id=$1`,
      [id],
    );
    return rows[0] || null;
  }

  async savePayment(payment: PaymentRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_payments (
        id, customer_id, subscription_id, billing_type, value, due_date, description, installment_count, callback_success_url, callback_auto_redirect, status, invoice_url, transaction_receipt_url, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
      [
        payment.id,
        payment.customerId,
        payment.subscriptionId ?? null,
        payment.billingType,
        payment.value,
        payment.dueDate,
        payment.description ?? '',
        payment.installmentCount ?? 0,
        payment.callbackSuccessUrl,
        payment.callbackAutoRedirect,
        payment.status,
        payment.invoiceUrl ?? '',
        payment.transactionReceiptUrl ?? '',
        payment.createdAt,
        payment.updatedAt,
      ],
    );
  }

  async findPaymentById(id: string): Promise<PaymentRecord | null> {
    const { rows } = await this.pool.query<PaymentRecord>(
      `SELECT id, customer_id as "customerId", subscription_id as "subscriptionId", billing_type as "billingType", value, due_date as "dueDate", description, installment_count as "installmentCount", callback_success_url as "callbackSuccessUrl", callback_auto_redirect as "callbackAutoRedirect", status, invoice_url as "invoiceUrl", transaction_receipt_url as "transactionReceiptUrl", created_at as "createdAt", updated_at as "updatedAt" FROM payment_payments WHERE id=$1`,
      [id],
    );
    return rows[0] || null;
  }

  async saveSubscription(subscription: SubscriptionRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_subscriptions (
        id, customer_id, billing_type, status, value, cycle, next_due_date, description, end_date, max_payments, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
      [
        subscription.id,
        subscription.customerId,
        subscription.billingType,
        subscription.status,
        subscription.value,
        subscription.cycle,
        subscription.nextDueDate,
        subscription.description ?? '',
        subscription.endDate ?? null,
        subscription.maxPayments ?? 0,
        subscription.createdAt,
        subscription.updatedAt,
      ],
    );
  }

  async saveInvoice(invoice: InvoiceRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_invoices (
        id, payment_id, service_description, observations, value, deductions, effective_date, municipal_service_id, municipal_service_code, municipal_service_name, update_payment, taxes_retain_iss, taxes_cofins, taxes_csll, taxes_inss, taxes_ir, taxes_pis, taxes_iss, status, payment_link, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
      [
        invoice.id,
        invoice.paymentId,
        invoice.serviceDescription,
        invoice.observations,
        invoice.value,
        invoice.deductions,
        invoice.effectiveDate,
        invoice.municipalServiceId ?? '',
        invoice.municipalServiceCode ?? '',
        invoice.municipalServiceName,
        invoice.updatePayment ?? false,
        invoice.taxes.retainIss,
        invoice.taxes.cofins,
        invoice.taxes.csll,
        invoice.taxes.inss,
        invoice.taxes.ir,
        invoice.taxes.pis,
        invoice.taxes.iss,
        invoice.status,
        invoice.paymentLink ?? '',
        invoice.createdAt,
        invoice.updatedAt,
      ],
    );
  }
}
