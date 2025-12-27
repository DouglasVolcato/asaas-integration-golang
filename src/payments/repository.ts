import { Pool } from "pg";
import {
  CustomerRecord,
  InvoiceRecord,
  PaymentRecord,
  SubscriptionRecord,
} from "./types";

export class PostgresRepository {
  constructor(private readonly pool: Pool) {}

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
        id, name, email, cpfCnpj, phone, mobile_phone, address, address_number, complement, province, postal_code,
        notification_disabled, additional_emails, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
      [
        customer.id,
        customer.name,
        customer.email,
        customer.cpfCnpj,
        customer.phone,
        customer.mobilePhone,
        customer.address,
        customer.addressNumber,
        customer.complement,
        customer.province,
        customer.postalCode,
        customer.notificationDisabled,
        customer.additionalEmails,
        customer.createdAt,
        customer.updatedAt,
      ]
    );
  }

  async findCustomerById(id: string): Promise<CustomerRecord | null> {
    const result = await this.pool.query(
      `SELECT id, name, email, cpfCnpj, phone, mobile_phone, address, address_number, complement, province, postal_code,
              notification_disabled, additional_emails, created_at, updated_at
       FROM payment_customers WHERE id=$1`,
      [id]
    );
    if (!result.rowCount) return null;
    const row = result.rows[0];
    return {
      id: row.id,
      name: row.name,
      email: row.email,
      cpfCnpj: row.cpfcnpj,
      phone: row.phone,
      mobilePhone: row.mobile_phone,
      address: row.address,
      addressNumber: row.address_number,
      complement: row.complement,
      province: row.province,
      postalCode: row.postal_code,
      notificationDisabled: row.notification_disabled,
      additionalEmails: row.additional_emails,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }

  async savePayment(payment: PaymentRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_payments (
        id, customer_id, subscription_id, billing_type, value, due_date, description, installment_count,
        callback_success_url, callback_auto_redirect, status, invoice_url, transaction_receipt_url, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
      [
        payment.id,
        payment.customerId,
        payment.subscriptionId,
        payment.billingType,
        payment.value,
        payment.dueDate,
        payment.description,
        payment.installmentCount,
        payment.callbackSuccessUrl,
        payment.callbackAutoRedirect,
        payment.status,
        payment.invoiceUrl,
        payment.transactionReceiptUrl,
        payment.createdAt,
        payment.updatedAt,
      ]
    );
  }

  async updatePaymentStatus(id: string, status: string, invoiceUrl?: string, receiptUrl?: string): Promise<void> {
    const result = await this.pool.query(
      `UPDATE payment_payments SET status=$1, invoice_url=$2, transaction_receipt_url=$3, updated_at=$4 WHERE id=$5`,
      [status, invoiceUrl ?? "", receiptUrl ?? "", new Date(), id]
    );
    if (!result.rowCount) {
      throw new Error("payment not found");
    }
  }

  async findPaymentById(id: string): Promise<PaymentRecord | null> {
    const result = await this.pool.query(
      `SELECT id, customer_id, subscription_id, billing_type, value, due_date, description, installment_count,
              callback_success_url, callback_auto_redirect, status, invoice_url, transaction_receipt_url, created_at, updated_at
       FROM payment_payments WHERE id=$1`,
      [id]
    );
    if (!result.rowCount) return null;
    const row = result.rows[0];
    return {
      id: row.id,
      customerId: row.customer_id,
      subscriptionId: row.subscription_id,
      billingType: row.billing_type,
      value: Number(row.value),
      dueDate: row.due_date,
      description: row.description,
      installmentCount: row.installment_count,
      callbackSuccessUrl: row.callback_success_url,
      callbackAutoRedirect: row.callback_auto_redirect,
      status: row.status,
      invoiceUrl: row.invoice_url,
      transactionReceiptUrl: row.transaction_receipt_url,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
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
        subscription.description,
        subscription.endDate,
        subscription.maxPayments,
        subscription.createdAt,
        subscription.updatedAt,
      ]
    );
  }

  async findSubscriptionById(id: string): Promise<SubscriptionRecord | null> {
    const result = await this.pool.query(
      `SELECT id, customer_id, billing_type, status, value, cycle, next_due_date, description, end_date, max_payments,
              created_at, updated_at FROM payment_subscriptions WHERE id=$1`,
      [id]
    );
    if (!result.rowCount) return null;
    const row = result.rows[0];
    return {
      id: row.id,
      customerId: row.customer_id,
      billingType: row.billing_type,
      status: row.status,
      value: Number(row.value),
      cycle: row.cycle,
      nextDueDate: row.next_due_date,
      description: row.description,
      endDate: row.end_date,
      maxPayments: row.max_payments,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }

  async updateSubscriptionStatus(id: string, status: string): Promise<void> {
    const result = await this.pool.query(
      `UPDATE payment_subscriptions SET status=$1, updated_at=$2 WHERE id=$3`,
      [status, new Date(), id]
    );
    if (!result.rowCount) {
      throw new Error("subscription not found");
    }
  }

  async saveInvoice(invoice: InvoiceRecord): Promise<void> {
    await this.pool.query(
      `INSERT INTO payment_invoices (
        id, payment_id, service_description, observations, value, deductions, effective_date, municipal_service_id, municipal_service_code,
        municipal_service_name, update_payment, taxes_retain_iss, taxes_cofins, taxes_csll, taxes_inss, taxes_ir, taxes_pis, taxes_iss,
        status, payment_link, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
      [
        invoice.id,
        invoice.paymentId,
        invoice.serviceDescription,
        invoice.observations,
        invoice.value,
        invoice.deductions,
        invoice.effectiveDate,
        invoice.municipalServiceId,
        invoice.municipalServiceCode,
        invoice.municipalServiceName,
        invoice.updatePayment,
        invoice.taxesRetainIss,
        invoice.taxesCofins,
        invoice.taxesCsll,
        invoice.taxesInss,
        invoice.taxesIr,
        invoice.taxesPis,
        invoice.taxesIss,
        invoice.status,
        invoice.paymentLink,
        invoice.createdAt,
        invoice.updatedAt,
      ]
    );
  }

  async findInvoiceByPaymentId(paymentId: string): Promise<InvoiceRecord | null> {
    const result = await this.pool.query(
      `SELECT id, payment_id, service_description, observations, value, deductions, effective_date, municipal_service_id,
              municipal_service_code, municipal_service_name, update_payment, taxes_retain_iss, taxes_cofins, taxes_csll, taxes_inss,
              taxes_ir, taxes_pis, taxes_iss, status, payment_link, created_at, updated_at
       FROM payment_invoices WHERE payment_id=$1 LIMIT 1`,
      [paymentId]
    );
    if (!result.rowCount) return null;
    const row = result.rows[0];
    return {
      id: row.id,
      paymentId: row.payment_id,
      serviceDescription: row.service_description,
      observations: row.observations,
      value: Number(row.value),
      deductions: Number(row.deductions),
      effectiveDate: row.effective_date,
      municipalServiceId: row.municipal_service_id,
      municipalServiceCode: row.municipal_service_code,
      municipalServiceName: row.municipal_service_name,
      updatePayment: row.update_payment,
      taxesRetainIss: row.taxes_retain_iss,
      taxesCofins: Number(row.taxes_cofins),
      taxesCsll: Number(row.taxes_csll),
      taxesInss: Number(row.taxes_inss),
      taxesIr: Number(row.taxes_ir),
      taxesPis: Number(row.taxes_pis),
      taxesIss: Number(row.taxes_iss),
      status: row.status,
      paymentLink: row.payment_link,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }

  async updateInvoiceStatus(id: string, status: string): Promise<void> {
    const result = await this.pool.query(
      `UPDATE payment_invoices SET status=$1, updated_at=$2 WHERE id=$3`,
      [status, new Date(), id]
    );
    if (!result.rowCount) {
      throw new Error("invoice not found");
    }
  }
}
