<?php

namespace AsaasIntegration;

use DateTimeImmutable;

class Service
{
    public function __construct(private readonly Repository $repository, private readonly AsaasClient $client)
    {
    }

    public function registerCustomer(array $payload): array
    {
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $localId = self::generateId();
        $payload['externalReference'] = $localId;

        $remote = $this->client->createCustomer($payload);

        $this->repository->saveCustomer([
            'id' => $localId,
            'name' => $payload['name'] ?? '',
            'email' => $payload['email'] ?? '',
            'cpfCnpj' => $payload['cpfCnpj'] ?? '',
            'phone' => $payload['phone'] ?? '',
            'mobile_phone' => $payload['mobilePhone'] ?? '',
            'address' => $payload['address'] ?? '',
            'address_number' => $payload['addressNumber'] ?? '',
            'complement' => $payload['complement'] ?? '',
            'province' => $payload['province'] ?? '',
            'postal_code' => $payload['postalCode'] ?? '',
            'notification_disabled' => !empty($payload['notificationDisabled']),
            'additional_emails' => $payload['additionalEmails'] ?? '',
            'created_at' => $now->format('Y-m-d H:i:sP'),
            'updated_at' => $now->format('Y-m-d H:i:sP'),
        ]);

        return $remote;
    }

    public function createPayment(array $payload): array
    {
        $customer = $this->repository->findCustomerById($payload['customer'] ?? '');
        $remoteCustomer = $this->client->getCustomer($customer['id']);

        $localId = self::generateId();
        $payload['externalReference'] = $localId;
        $payload['customer'] = $remoteCustomer['id'];

        $remote = $this->client->createPayment($payload);
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));

        $callback = $payload['callback'] ?? [];

        $this->repository->savePayment([
            'id' => $localId,
            'customer_id' => $customer['id'],
            'subscription_id' => $payload['subscription'] ?? null,
            'billing_type' => $payload['billingType'] ?? '',
            'value' => $payload['value'] ?? 0,
            'due_date' => self::parseDate($payload['dueDate'] ?? ''),
            'description' => $payload['description'] ?? '',
            'installment_count' => $payload['installmentCount'] ?? 0,
            'callback_success_url' => $callback['successUrl'] ?? '',
            'callback_auto_redirect' => !empty($callback['autoRedirect']),
            'status' => $remote['status'] ?? '',
            'invoice_url' => $remote['invoiceUrl'] ?? '',
            'transaction_receipt_url' => $remote['transactionReceiptUrl'] ?? '',
            'created_at' => $now->format('Y-m-d H:i:sP'),
            'updated_at' => $now->format('Y-m-d H:i:sP'),
        ]);

        return $remote;
    }

    public function createSubscription(array $payload): array
    {
        $customer = $this->repository->findCustomerById($payload['customer'] ?? '');
        $remoteCustomer = $this->client->getCustomer($customer['id']);

        $localId = self::generateId();
        $payload['externalReference'] = $localId;
        $payload['customer'] = $remoteCustomer['id'];

        $remote = $this->client->createSubscription($payload);
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));

        $this->repository->saveSubscription([
            'id' => $localId,
            'customer_id' => $customer['id'],
            'billing_type' => $payload['billingType'] ?? '',
            'status' => $remote['status'] ?? '',
            'value' => $payload['value'] ?? 0,
            'cycle' => $payload['cycle'] ?? '',
            'next_due_date' => self::parseDate($payload['nextDueDate'] ?? ''),
            'description' => $payload['description'] ?? '',
            'end_date' => self::parseDate($payload['endDate'] ?? ''),
            'max_payments' => $payload['maxPayments'] ?? 0,
            'created_at' => $now->format('Y-m-d H:i:sP'),
            'updated_at' => $now->format('Y-m-d H:i:sP'),
        ]);

        return $remote;
    }

    public function createInvoice(array $payload): array
    {
        $payment = $this->repository->findPaymentById($payload['payment'] ?? '');
        $remotePayment = $this->client->getPayment($payment['id']);

        $external = $payload['externalReference'] ?? '';
        if ($external === '') {
            $external = $payment['id'];
        }
        $payload['externalReference'] = $external;
        $payload['payment'] = $remotePayment['id'];

        $remote = $this->client->createInvoice($payload);
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));

        $taxes = $payload['taxes'] ?? [];

        $this->repository->saveInvoice([
            'id' => $external,
            'payment_id' => $payment['id'],
            'service_description' => $payload['serviceDescription'] ?? '',
            'observations' => $payload['observations'] ?? '',
            'value' => $payload['value'] ?? 0,
            'deductions' => $payload['deductions'] ?? 0,
            'effective_date' => self::parseDate($payload['effectiveDate'] ?? ''),
            'municipal_service_id' => $payload['municipalServiceId'] ?? '',
            'municipal_service_code' => $payload['municipalServiceCode'] ?? '',
            'municipal_service_name' => $payload['municipalServiceName'] ?? '',
            'update_payment' => !empty($payload['updatePayment']),
            'taxes_retain_iss' => !empty($taxes['retainIss']),
            'taxes_cofins' => $taxes['cofins'] ?? 0,
            'taxes_csll' => $taxes['csll'] ?? 0,
            'taxes_inss' => $taxes['inss'] ?? 0,
            'taxes_ir' => $taxes['ir'] ?? 0,
            'taxes_pis' => $taxes['pis'] ?? 0,
            'taxes_iss' => $taxes['iss'] ?? 0,
            'status' => $remote['status'] ?? '',
            'payment_link' => $remote['paymentLink'] ?? '',
            'created_at' => $now->format('Y-m-d H:i:sP'),
            'updated_at' => $now->format('Y-m-d H:i:sP'),
        ]);

        return $remote;
    }

    public function handleWebhook(array $event): void
    {
        $type = $event['event'] ?? '';
        switch ($type) {
            case 'PAYMENT_CREATED':
                $payment = $event['payment'] ?? null;
                if (!$payment || empty($payment['subscription'])) {
                    return;
                }
                if (!empty($payment['externalReference'])) {
                    try {
                        $this->repository->findPaymentById($payment['externalReference']);
                        return;
                    } catch (\RuntimeException) {
                    }
                }
                $subscription = $this->client->getSubscriptionById($payment['subscription']);
                if (empty($subscription['externalReference'])) {
                    return;
                }
                try {
                    $localSubscription = $this->repository->findSubscriptionById($subscription['externalReference']);
                } catch (\RuntimeException) {
                    return;
                }

                $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
                $localId = self::generateId();
                $this->repository->savePayment([
                    'id' => $localId,
                    'customer_id' => $localSubscription['customer_id'],
                    'subscription_id' => $localSubscription['id'],
                    'billing_type' => $payment['billingType'] ?? '',
                    'value' => $payment['value'] ?? 0,
                    'due_date' => self::parseDate($payment['dueDate'] ?? ''),
                    'description' => $payment['description'] ?? '',
                    'installment_count' => 0,
                    'callback_success_url' => '',
                    'callback_auto_redirect' => false,
                    'status' => $payment['status'] ?? '',
                    'invoice_url' => $payment['invoiceUrl'] ?? '',
                    'transaction_receipt_url' => $payment['transactionReceiptUrl'] ?? '',
                    'created_at' => $now->format('Y-m-d H:i:sP'),
                    'updated_at' => $now->format('Y-m-d H:i:sP'),
                ]);

                if (!empty($payment['id']) && ($payment['externalReference'] ?? '') !== $localId) {
                    $this->client->updatePaymentExternalReference($payment['id'], $localId);
                }
                return;
            case 'INVOICE_CREATED':
            case 'SUBSCRIPTION_CREATED':
                return;
            case 'PAYMENT_AUTHORIZED':
            case 'PAYMENT_APPROVED_BY_RISK_ANALYSIS':
            case 'PAYMENT_CONFIRMED':
            case 'PAYMENT_ANTICIPATED':
            case 'PAYMENT_DELETED':
            case 'PAYMENT_REFUNDED':
            case 'PAYMENT_REFUND_DENIED':
            case 'PAYMENT_CHARGEBACK_REQUESTED':
            case 'PAYMENT_AWAITING_CHARGEBACK_REVERSAL':
            case 'PAYMENT_DUNNING_REQUESTED':
            case 'PAYMENT_CHECKOUT_VIEWED':
            case 'PAYMENT_PARTIALLY_REFUNDED':
            case 'PAYMENT_SPLIT_DIVERGENCE_BLOCK':
            case 'PAYMENT_AWAITING_RISK_ANALYSIS':
            case 'PAYMENT_REPROVED_BY_RISK_ANALYSIS':
            case 'PAYMENT_UPDATED':
            case 'PAYMENT_RECEIVED':
            case 'PAYMENT_OVERDUE':
            case 'PAYMENT_RESTORED':
            case 'PAYMENT_REFUND_IN_PROGRESS':
            case 'PAYMENT_RECEIVED_IN_CASH_UNDONE':
            case 'PAYMENT_CHARGEBACK_DISPUTE':
            case 'PAYMENT_DUNNING_RECEIVED':
            case 'PAYMENT_BANK_SLIP_VIEWED':
            case 'PAYMENT_CREDIT_CARD_CAPTURE_REFUSED':
            case 'PAYMENT_SPLIT_CANCELLED':
            case 'PAYMENT_SPLIT_DIVERGENCE_BLOCK_FINISHED':
                $payment = $event['payment'] ?? null;
                if (!$payment || empty($payment['externalReference'])) {
                    return;
                }
                try {
                    $local = $this->repository->findPaymentById($payment['externalReference']);
                } catch (\RuntimeException) {
                    return;
                }
                $this->repository->updatePaymentStatus($local['id'], $payment['status'] ?? '', $payment['invoiceUrl'] ?? '', $payment['transactionReceiptUrl'] ?? '');
                $this->issueInvoiceForPayment($local, $payment);
                return;
            case 'SUBSCRIPTION_INACTIVATED':
            case 'SUBSCRIPTION_SPLIT_DISABLED':
            case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK_FINISHED':
            case 'SUBSCRIPTION_UPDATED':
            case 'SUBSCRIPTION_DELETED':
            case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK':
                $subscription = $event['subscription'] ?? null;
                if (!$subscription || empty($subscription['externalReference'])) {
                    return;
                }
                $this->repository->updateSubscriptionStatus($subscription['externalReference'], $subscription['status'] ?? '');
                return;
            case 'INVOICE_SYNCHRONIZED':
            case 'INVOICE_PROCESSING_CANCELLATION':
            case 'INVOICE_CANCELLATION_DENIED':
            case 'INVOICE_UPDATED':
            case 'INVOICE_AUTHORIZED':
            case 'INVOICE_CANCELED':
            case 'INVOICE_ERROR':
                $invoice = $event['invoice'] ?? null;
                if (!$invoice || empty($invoice['externalReference'])) {
                    return;
                }
                $this->repository->updateInvoiceStatus($invoice['externalReference'], $invoice['status'] ?? '');
                return;
            default:
                throw new \RuntimeException('tipo de evento não suportado: ' . $type);
        }
    }

    private function issueInvoiceForPayment(array $payment, array $payload): void
    {
        try {
            $this->repository->findInvoiceByPaymentId($payment['id']);
            return;
        } catch (\RuntimeException) {
        }

        $req = [
            'payment' => $payment['id'],
            'serviceDescription' => $payment['description'] !== '' ? $payment['description'] : 'Pagamento ' . $payment['id'],
            'observations' => 'NOTA FISCAL EMITIDA POR EMPRESA OPTANTE DO SIMPLES NACIONAL CONFORME LEI COMPLEMENTAR 123/2006. NÃO GERA DIREITO A CRÉDITO DE I.P.I./ICMS.',
            'externalReference' => $payment['id'],
            'value' => $payment['value'],
            'deductions' => 0,
            'effectiveDate' => (new DateTimeImmutable('now', new \DateTimeZone('UTC')))->format('Y-m-d'),
            'municipalServiceCode' => '01.03.01',
            'municipalServiceName' => 'Processamento, armazenamento ou hospedagem de dados, textos, imagens, vídeos, páginas eletrônicas, aplicativos e sistemas de informação, entre outros formatos, e congêneres',
            'updatePayment' => true,
            'taxes' => [
                'retainIss' => false,
                'cofins' => 0,
                'csll' => 0,
                'inss' => 0,
                'ir' => 0,
                'pis' => 0,
                'iss' => 5,
            ],
        ];

        $this->createInvoice($req);
    }

    public static function parseDate(string $date): string
    {
        if ($date === '') {
            return (new DateTimeImmutable('now', new \DateTimeZone('UTC')))->format('Y-m-d');
        }
        return (new DateTimeImmutable($date))->format('Y-m-d');
    }

    public static function generateId(): string
    {
        $bytes = random_bytes(16);
        $bytes[6] = chr((ord($bytes[6]) & 0x0f) | 0x40);
        $bytes[8] = chr((ord($bytes[8]) & 0x3f) | 0x80);
        $hex = bin2hex($bytes);
        return sprintf('%s-%s-%s-%s-%s', substr($hex, 0, 8), substr($hex, 8, 4), substr($hex, 12, 4), substr($hex, 16, 4), substr($hex, 20));
    }
}
