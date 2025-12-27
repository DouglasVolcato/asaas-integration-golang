<?php

namespace AsaasIntegration;

use PDO;
use PDOException;

class Repository
{
    public function __construct(private readonly PDO $pdo)
    {
    }

    public function ensureSchema(): void
    {
        $stmts = [
            "CREATE TABLE IF NOT EXISTS payment_customers (
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
            );",
            "CREATE TABLE IF NOT EXISTS payment_payments (
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
            );",
            "ALTER TABLE payment_payments ADD COLUMN IF NOT EXISTS subscription_id UUID;",
            "CREATE TABLE IF NOT EXISTS payment_subscriptions (
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
            );",
            "CREATE TABLE IF NOT EXISTS payment_invoices (
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
            );",
        ];

        foreach ($stmts as $sql) {
            $this->pdo->exec($sql);
        }
    }

    public function saveCustomer(array $customer): void
    {
        $stmt = $this->pdo->prepare(
            'INSERT INTO payment_customers (
                id, name, email, cpfCnpj, phone, mobile_phone, address, address_number, complement, province, postal_code,
                notification_disabled, additional_emails, created_at, updated_at
            ) VALUES (
                :id, :name, :email, :cpfCnpj, :phone, :mobile_phone, :address, :address_number, :complement, :province, :postal_code,
                :notification_disabled, :additional_emails, :created_at, :updated_at
            )'
        );
        $stmt->execute($customer);
    }

    public function findCustomerById(string $id): array
    {
        $stmt = $this->pdo->prepare('SELECT * FROM payment_customers WHERE id = :id');
        $stmt->execute(['id' => $id]);
        $row = $stmt->fetch();
        if (!$row) {
            throw new \RuntimeException('customer not found');
        }
        return $row;
    }

    public function savePayment(array $payment): void
    {
        $stmt = $this->pdo->prepare(
            'INSERT INTO payment_payments (
                id, customer_id, subscription_id, billing_type, value, due_date, description, installment_count, callback_success_url,
                callback_auto_redirect, status, invoice_url, transaction_receipt_url, created_at, updated_at
            ) VALUES (
                :id, :customer_id, :subscription_id, :billing_type, :value, :due_date, :description, :installment_count,
                :callback_success_url, :callback_auto_redirect, :status, :invoice_url, :transaction_receipt_url, :created_at, :updated_at
            )'
        );
        $stmt->execute($payment);
    }

    public function findPaymentById(string $id): array
    {
        $stmt = $this->pdo->prepare('SELECT * FROM payment_payments WHERE id = :id');
        $stmt->execute(['id' => $id]);
        $row = $stmt->fetch();
        if (!$row) {
            throw new \RuntimeException('payment not found');
        }
        return $row;
    }

    public function updatePaymentStatus(string $id, string $status, string $invoiceUrl, string $transactionReceiptUrl): void
    {
        $stmt = $this->pdo->prepare('UPDATE payment_payments SET status = :status, invoice_url = :invoice_url, transaction_receipt_url = :transaction_receipt_url, updated_at = NOW() WHERE id = :id');
        $stmt->execute([
            'status' => $status,
            'invoice_url' => $invoiceUrl,
            'transaction_receipt_url' => $transactionReceiptUrl,
            'id' => $id,
        ]);
    }

    public function saveSubscription(array $subscription): void
    {
        $stmt = $this->pdo->prepare(
            'INSERT INTO payment_subscriptions (
                id, customer_id, billing_type, status, value, cycle, next_due_date, description, end_date, max_payments, created_at, updated_at
            ) VALUES (
                :id, :customer_id, :billing_type, :status, :value, :cycle, :next_due_date, :description, :end_date, :max_payments, :created_at, :updated_at
            )'
        );
        $stmt->execute($subscription);
    }

    public function findSubscriptionById(string $id): array
    {
        $stmt = $this->pdo->prepare('SELECT * FROM payment_subscriptions WHERE id = :id');
        $stmt->execute(['id' => $id]);
        $row = $stmt->fetch();
        if (!$row) {
            throw new \RuntimeException('subscription not found');
        }
        return $row;
    }

    public function updateSubscriptionStatus(string $id, string $status): void
    {
        $stmt = $this->pdo->prepare('UPDATE payment_subscriptions SET status = :status, updated_at = NOW() WHERE id = :id');
        $stmt->execute(['status' => $status, 'id' => $id]);
    }

    public function saveInvoice(array $invoice): void
    {
        $stmt = $this->pdo->prepare(
            'INSERT INTO payment_invoices (
                id, payment_id, service_description, observations, value, deductions, effective_date, municipal_service_id, municipal_service_code,
                municipal_service_name, update_payment, taxes_retain_iss, taxes_cofins, taxes_csll, taxes_inss, taxes_ir, taxes_pis, taxes_iss,
                status, payment_link, created_at, updated_at
            ) VALUES (
                :id, :payment_id, :service_description, :observations, :value, :deductions, :effective_date, :municipal_service_id, :municipal_service_code,
                :municipal_service_name, :update_payment, :taxes_retain_iss, :taxes_cofins, :taxes_csll, :taxes_inss, :taxes_ir, :taxes_pis, :taxes_iss,
                :status, :payment_link, :created_at, :updated_at
            )'
        );
        $stmt->execute($invoice);
    }

    public function findInvoiceByPaymentId(string $paymentId): array
    {
        $stmt = $this->pdo->prepare('SELECT * FROM payment_invoices WHERE payment_id = :payment_id');
        $stmt->execute(['payment_id' => $paymentId]);
        $row = $stmt->fetch();
        if (!$row) {
            throw new \RuntimeException('invoice not found');
        }
        return $row;
    }

    public function updateInvoiceStatus(string $externalId, string $status): void
    {
        $stmt = $this->pdo->prepare('UPDATE payment_invoices SET status = :status, updated_at = NOW() WHERE id = :id');
        $stmt->execute(['status' => $status, 'id' => $externalId]);
    }
}
