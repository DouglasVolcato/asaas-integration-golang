<?php

namespace App;

use PDO;
use RuntimeException;

class Repository
{
    public function __construct(private PDO $db)
    {
    }

    public function ensureSchema(): void
    {
        $statements = [
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

        foreach ($statements as $sql) {
            $this->db->exec($sql);
        }
    }

    public function saveCustomer(CustomerRecord $customer): void
    {
        $stmt = $this->db->prepare('INSERT INTO payment_customers (
            id, name, email, cpfCnpj, phone, mobile_phone, address, address_number, complement, province, postal_code, notification_disabled, additional_emails, created_at, updated_at
        ) VALUES (:id,:name,:email,:cpf,:phone,:mobile,:address,:addressNumber,:complement,:province,:postalCode,:notificationDisabled,:additionalEmails,:createdAt,:updatedAt)');
        $stmt->execute([
            ':id' => $customer->id,
            ':name' => $customer->name,
            ':email' => $customer->email,
            ':cpf' => $customer->cpfCnpj,
            ':phone' => $customer->phone,
            ':mobile' => $customer->mobilePhone,
            ':address' => $customer->address,
            ':addressNumber' => $customer->addressNumber,
            ':complement' => $customer->complement,
            ':province' => $customer->province,
            ':postalCode' => $customer->postalCode,
            ':notificationDisabled' => $customer->notificationDisabled,
            ':additionalEmails' => $customer->additionalEmails,
            ':createdAt' => $customer->createdAt->format('c'),
            ':updatedAt' => $customer->updatedAt->format('c'),
        ]);
    }

    public function findCustomerById(string $id): CustomerRecord
    {
        $stmt = $this->db->prepare('SELECT * FROM payment_customers WHERE id = :id');
        $stmt->execute([':id' => $id]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
        if (!$row) {
            throw new NotFoundException('not found');
        }
        return new CustomerRecord(
            $row['id'],
            $row['name'],
            $row['email'],
            $row['cpfcnpj'],
            $row['phone'],
            $row['mobile_phone'],
            $row['address'],
            $row['address_number'],
            $row['complement'],
            $row['province'],
            $row['postal_code'],
            (bool)$row['notification_disabled'],
            $row['additional_emails'],
            new \DateTimeImmutable($row['created_at']),
            new \DateTimeImmutable($row['updated_at'])
        );
    }

    public function savePayment(PaymentRecord $payment): void
    {
        $stmt = $this->db->prepare('INSERT INTO payment_payments (
            id, customer_id, subscription_id, billing_type, value, due_date, description, installment_count, callback_success_url, callback_auto_redirect, status, invoice_url, transaction_receipt_url, created_at, updated_at
        ) VALUES (:id,:customerId,:subscriptionId,:billingType,:value,:dueDate,:description,:installmentCount,:callbackSuccessUrl,:callbackAutoRedirect,:status,:invoiceUrl,:receiptUrl,:createdAt,:updatedAt)');
        $stmt->execute([
            ':id' => $payment->id,
            ':customerId' => $payment->customerId,
            ':subscriptionId' => $payment->subscriptionId,
            ':billingType' => $payment->billingType,
            ':value' => $payment->value,
            ':dueDate' => $payment->dueDate->format('c'),
            ':description' => $payment->description,
            ':installmentCount' => $payment->installmentCount,
            ':callbackSuccessUrl' => $payment->callbackSuccessUrl,
            ':callbackAutoRedirect' => $payment->callbackAutoRedirect,
            ':status' => $payment->status,
            ':invoiceUrl' => $payment->invoiceUrl,
            ':receiptUrl' => $payment->transactionReceiptUrl,
            ':createdAt' => $payment->createdAt->format('c'),
            ':updatedAt' => $payment->updatedAt->format('c'),
        ]);
    }

    public function updatePaymentStatus(string $id, string $status, string $invoiceUrl, string $receiptUrl): void
    {
        $stmt = $this->db->prepare('UPDATE payment_payments SET status=:status, invoice_url=:invoiceUrl, transaction_receipt_url=:receiptUrl, updated_at=:updatedAt WHERE id=:id');
        $stmt->execute([
            ':status' => $status,
            ':invoiceUrl' => $invoiceUrl,
            ':receiptUrl' => $receiptUrl,
            ':updatedAt' => (new \DateTimeImmutable())->format('c'),
            ':id' => $id,
        ]);
        if ($stmt->rowCount() === 0) {
            throw new NotFoundException('not found');
        }
    }

    public function findPaymentById(string $id): PaymentRecord
    {
        $stmt = $this->db->prepare('SELECT * FROM payment_payments WHERE id = :id');
        $stmt->execute([':id' => $id]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
        if (!$row) {
            throw new NotFoundException('not found');
        }
        return new PaymentRecord(
            $row['id'],
            $row['customer_id'],
            $row['subscription_id'] ?: null,
            $row['billing_type'],
            (float)$row['value'],
            new \DateTimeImmutable($row['due_date']),
            $row['description'],
            (int)$row['installment_count'],
            $row['callback_success_url'],
            (bool)$row['callback_auto_redirect'],
            $row['status'],
            $row['invoice_url'],
            $row['transaction_receipt_url'],
            new \DateTimeImmutable($row['created_at']),
            new \DateTimeImmutable($row['updated_at'])
        );
    }

    public function saveSubscription(SubscriptionRecord $subscription): void
    {
        $stmt = $this->db->prepare('INSERT INTO payment_subscriptions (
            id, customer_id, billing_type, status, value, cycle, next_due_date, description, end_date, max_payments, created_at, updated_at
        ) VALUES (:id,:customerId,:billingType,:status,:value,:cycle,:nextDueDate,:description,:endDate,:maxPayments,:createdAt,:updatedAt)');
        $stmt->execute([
            ':id' => $subscription->id,
            ':customerId' => $subscription->customerId,
            ':billingType' => $subscription->billingType,
            ':status' => $subscription->status,
            ':value' => $subscription->value,
            ':cycle' => $subscription->cycle,
            ':nextDueDate' => $subscription->nextDueDate->format('c'),
            ':description' => $subscription->description,
            ':endDate' => $subscription->endDate?->format('c'),
            ':maxPayments' => $subscription->maxPayments,
            ':createdAt' => $subscription->createdAt->format('c'),
            ':updatedAt' => $subscription->updatedAt->format('c'),
        ]);
    }

    public function findSubscriptionById(string $id): SubscriptionRecord
    {
        $stmt = $this->db->prepare('SELECT * FROM payment_subscriptions WHERE id = :id');
        $stmt->execute([':id' => $id]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
        if (!$row) {
            throw new NotFoundException('not found');
        }
        return new SubscriptionRecord(
            $row['id'],
            $row['customer_id'],
            $row['billing_type'],
            $row['status'],
            (float)$row['value'],
            $row['cycle'],
            new \DateTimeImmutable($row['next_due_date']),
            $row['description'],
            $row['end_date'] ? new \DateTimeImmutable($row['end_date']) : null,
            (int)$row['max_payments'],
            new \DateTimeImmutable($row['created_at']),
            new \DateTimeImmutable($row['updated_at'])
        );
    }

    public function updateSubscriptionStatus(string $id, string $status): void
    {
        $stmt = $this->db->prepare('UPDATE payment_subscriptions SET status=:status, updated_at=:updatedAt WHERE id=:id');
        $stmt->execute([
            ':status' => $status,
            ':updatedAt' => (new \DateTimeImmutable())->format('c'),
            ':id' => $id,
        ]);
        if ($stmt->rowCount() === 0) {
            throw new NotFoundException('not found');
        }
    }

    public function saveInvoice(InvoiceRecord $invoice): void
    {
        $stmt = $this->db->prepare('INSERT INTO payment_invoices (
            id, payment_id, service_description, observations, value, deductions, effective_date, municipal_service_id, municipal_service_code, municipal_service_name, update_payment, taxes_retain_iss, taxes_cofins, taxes_csll, taxes_inss, taxes_ir, taxes_pis, taxes_iss, status, payment_link, created_at, updated_at
        ) VALUES (:id,:paymentId,:serviceDescription,:observations,:value,:deductions,:effectiveDate,:municipalServiceId,:municipalServiceCode,:municipalServiceName,:updatePayment,:taxesRetainIss,:taxesCofins,:taxesCsll,:taxesInss,:taxesIr,:taxesPis,:taxesIss,:status,:paymentLink,:createdAt,:updatedAt)');
        $stmt->execute([
            ':id' => $invoice->id,
            ':paymentId' => $invoice->paymentId,
            ':serviceDescription' => $invoice->serviceDescription,
            ':observations' => $invoice->observations,
            ':value' => $invoice->value,
            ':deductions' => $invoice->deductions,
            ':effectiveDate' => $invoice->effectiveDate->format('c'),
            ':municipalServiceId' => $invoice->municipalServiceId,
            ':municipalServiceCode' => $invoice->municipalServiceCode,
            ':municipalServiceName' => $invoice->municipalServiceName,
            ':updatePayment' => $invoice->updatePayment,
            ':taxesRetainIss' => $invoice->taxesRetainISS,
            ':taxesCofins' => $invoice->taxesCofins,
            ':taxesCsll' => $invoice->taxesCsll,
            ':taxesInss' => $invoice->taxesINSS,
            ':taxesIr' => $invoice->taxesIR,
            ':taxesPis' => $invoice->taxesPIS,
            ':taxesIss' => $invoice->taxesISS,
            ':status' => $invoice->status,
            ':paymentLink' => $invoice->paymentLink,
            ':createdAt' => $invoice->createdAt->format('c'),
            ':updatedAt' => $invoice->updatedAt->format('c'),
        ]);
    }

    public function findInvoiceByPaymentId(string $paymentId): InvoiceRecord
    {
        $stmt = $this->db->prepare('SELECT * FROM payment_invoices WHERE payment_id = :paymentId LIMIT 1');
        $stmt->execute([':paymentId' => $paymentId]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
        if (!$row) {
            throw new NotFoundException('not found');
        }
        return new InvoiceRecord(
            $row['id'],
            $row['payment_id'],
            $row['service_description'],
            $row['observations'],
            (float)$row['value'],
            (float)$row['deductions'],
            new \DateTimeImmutable($row['effective_date']),
            $row['municipal_service_id'],
            $row['municipal_service_code'],
            $row['municipal_service_name'],
            (bool)$row['update_payment'],
            (bool)$row['taxes_retain_iss'],
            (float)$row['taxes_cofins'],
            (float)$row['taxes_csll'],
            (float)$row['taxes_inss'],
            (float)$row['taxes_ir'],
            (float)$row['taxes_pis'],
            (float)$row['taxes_iss'],
            $row['status'],
            $row['payment_link'],
            new \DateTimeImmutable($row['created_at']),
            new \DateTimeImmutable($row['updated_at'])
        );
    }

    public function updateInvoiceStatus(string $id, string $status): void
    {
        $stmt = $this->db->prepare('UPDATE payment_invoices SET status=:status, updated_at=:updatedAt WHERE id=:id');
        $stmt->execute([
            ':status' => $status,
            ':updatedAt' => (new \DateTimeImmutable())->format('c'),
            ':id' => $id,
        ]);
        if ($stmt->rowCount() === 0) {
            throw new NotFoundException('not found');
        }
    }
}
