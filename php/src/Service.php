<?php

namespace App;

use DateTimeImmutable;
use RuntimeException;

class Service
{
    public function __construct(
        private readonly Repository $repo,
        private readonly AsaasClient $client
    ) {
    }

    public function registerCustomer(array $payload): array
    {
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $localId = self::generateId();
        $payload['externalReference'] = $localId;

        $remote = $this->client->createCustomer($payload);

        $customer = new CustomerRecord(
            $localId,
            $payload['name'] ?? '',
            $payload['email'] ?? '',
            $payload['cpfCnpj'] ?? '',
            $payload['phone'] ?? '',
            $payload['mobilePhone'] ?? '',
            $payload['address'] ?? '',
            $payload['addressNumber'] ?? '',
            $payload['complement'] ?? '',
            $payload['province'] ?? '',
            $payload['postalCode'] ?? '',
            (bool)($payload['notificationDisabled'] ?? false),
            $payload['additionalEmails'] ?? '',
            $now,
            $now
        );

        $this->repo->saveCustomer($customer);
        return $remote;
    }

    public function createPayment(array $payload): array
    {
        $customer = $this->repo->findCustomerById($payload['customer'] ?? '');
        $remoteCustomer = $this->client->getCustomer($customer->id);

        $localId = self::generateId();
        $payload['externalReference'] = $localId;
        $asaasPayload = $payload;
        $asaasPayload['customer'] = $remoteCustomer['id'];
        $remote = $this->client->createPayment($asaasPayload);

        $callbackSuccessUrl = '';
        $callbackAutoRedirect = false;
        if (isset($payload['callback'])) {
            $callbackSuccessUrl = $payload['callback']['successUrl'] ?? '';
            $callbackAutoRedirect = (bool)($payload['callback']['autoRedirect'] ?? false);
        }

        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $payment = new PaymentRecord(
            $localId,
            $customer->id,
            null,
            $payload['billingType'] ?? '',
            (float)($payload['value'] ?? 0),
            self::parseDate($payload['dueDate'] ?? ''),
            $payload['description'] ?? '',
            (int)($payload['installmentCount'] ?? 0),
            $callbackSuccessUrl,
            $callbackAutoRedirect,
            $remote['status'] ?? '',
            $remote['invoiceUrl'] ?? '',
            $remote['transactionReceiptUrl'] ?? '',
            $now,
            $now
        );

        $this->repo->savePayment($payment);
        return $remote;
    }

    public function createSubscription(array $payload): array
    {
        $customer = $this->repo->findCustomerById($payload['customer'] ?? '');
        $remoteCustomer = $this->client->getCustomer($customer->id);

        $localId = self::generateId();
        $payload['externalReference'] = $localId;
        $asaasPayload = $payload;
        $asaasPayload['customer'] = $remoteCustomer['id'];
        $remote = $this->client->createSubscription($asaasPayload);

        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $subscription = new SubscriptionRecord(
            $localId,
            $customer->id,
            $payload['billingType'] ?? '',
            $remote['status'] ?? '',
            (float)($payload['value'] ?? 0),
            $payload['cycle'] ?? '',
            self::parseDate($payload['nextDueDate'] ?? ''),
            $payload['description'] ?? '',
            ($payload['endDate'] ?? '') !== '' ? self::parseDate($payload['endDate']) : null,
            (int)($payload['maxPayments'] ?? 0),
            $now,
            $now
        );
        $this->repo->saveSubscription($subscription);
        return $remote;
    }

    public function createInvoice(array $payload): array
    {
        $payment = $this->repo->findPaymentById($payload['payment'] ?? '');
        $remotePayment = $this->client->getPayment($payment->id);

        $localId = $payload['externalReference'] ?? '';
        if ($localId === '') {
            $localId = $payment->id;
        }
        $payload['externalReference'] = $localId;
        $asaasPayload = $payload;
        $asaasPayload['payment'] = $remotePayment['id'];
        $remote = $this->client->createInvoice($asaasPayload);

        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $taxes = is_array($payload['taxes'] ?? null) ? $payload['taxes'] : [];

        $invoice = new InvoiceRecord(
            $localId,
            $payment->id,
            $payload['serviceDescription'] ?? '',
            $payload['observations'] ?? '',
            (float)($payload['value'] ?? 0),
            (float)($payload['deductions'] ?? 0),
            self::parseDate($payload['effectiveDate'] ?? ''),
            $payload['municipalServiceId'] ?? '',
            $payload['municipalServiceCode'] ?? '',
            $payload['municipalServiceName'] ?? '',
            (bool)($payload['updatePayment'] ?? false),
            (bool)($taxes['retainIss'] ?? false),
            (float)($taxes['cofins'] ?? 0),
            (float)($taxes['csll'] ?? 0),
            (float)($taxes['inss'] ?? 0),
            (float)($taxes['ir'] ?? 0),
            (float)($taxes['pis'] ?? 0),
            (float)($taxes['iss'] ?? 0),
            $remote['status'] ?? '',
            $remote['paymentLink'] ?? '',
            $now,
            $now
        );

        $this->repo->saveInvoice($invoice);
        return $remote;
    }

    public function handleWebhook(array $payload): void
    {
        $event = $payload['event'] ?? '';
        switch ($event) {
            case 'PAYMENT_CREATED':
                $this->handlePaymentCreated($payload);
                break;
            case 'INVOICE_CREATED':
            case 'SUBSCRIPTION_CREATED':
                return; // ignored
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
                $payment = $payload['payment'] ?? null;
                if (!$payment) {
                    throw new RuntimeException('payload de pagamento ausente');
                }
                $this->handlePaymentStatusChange($payment);
                break;
            case 'SUBSCRIPTION_INACTIVATED':
            case 'SUBSCRIPTION_SPLIT_DISABLED':
            case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK_FINISHED':
            case 'SUBSCRIPTION_UPDATED':
            case 'SUBSCRIPTION_DELETED':
            case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK':
                $subscription = $payload['subscription'] ?? null;
                if (!$subscription) {
                    throw new RuntimeException('payload de assinatura ausente');
                }
                $this->repo->updateSubscriptionStatus($subscription['externalReference'] ?? '', $subscription['status'] ?? '');
                break;
            case 'INVOICE_SYNCHRONIZED':
            case 'INVOICE_PROCESSING_CANCELLATION':
            case 'INVOICE_CANCELLATION_DENIED':
            case 'INVOICE_UPDATED':
            case 'INVOICE_AUTHORIZED':
            case 'INVOICE_CANCELED':
            case 'INVOICE_ERROR':
                $invoice = $payload['invoice'] ?? null;
                if (!$invoice) {
                    throw new RuntimeException('payload de nota fiscal ausente');
                }
                $this->repo->updateInvoiceStatus($invoice['externalReference'] ?? '', $invoice['status'] ?? '');
                break;
            default:
                throw new RuntimeException('tipo de evento não suportado: ' . $event);
        }
    }

    private function handlePaymentCreated(array $payload): void
    {
        $paymentPayload = $payload['payment'] ?? null;
        if (!$paymentPayload) {
            throw new RuntimeException('payload de pagamento ausente');
        }
        if (($paymentPayload['subscription'] ?? '') === '') {
            return;
        }
        if (($paymentPayload['externalReference'] ?? '') !== '') {
            try {
                $this->repo->findPaymentById($paymentPayload['externalReference']);
                return;
            } catch (NotFoundException) {
                // create below
            }
        }

        $subscription = $this->client->getSubscriptionById($paymentPayload['subscription']);
        if (($subscription['externalReference'] ?? '') === '') {
            throw new RuntimeException('externalReference da assinatura ausente para id ' . $paymentPayload['subscription']);
        }

        try {
            $localSubscription = $this->repo->findSubscriptionById($subscription['externalReference']);
        } catch (NotFoundException) {
            return;
        }

        $localId = self::generateId();
        $now = new DateTimeImmutable('now', new \DateTimeZone('UTC'));
        $payment = new PaymentRecord(
            $localId,
            $localSubscription->customerId,
            $localSubscription->id,
            $paymentPayload['billingType'] ?? '',
            (float)($paymentPayload['value'] ?? 0),
            self::parseDate($paymentPayload['dueDate'] ?? ''),
            $paymentPayload['description'] ?? '',
            0,
            '',
            false,
            $paymentPayload['status'] ?? '',
            $paymentPayload['invoiceUrl'] ?? '',
            $paymentPayload['transactionReceiptUrl'] ?? '',
            $now,
            $now
        );
        $this->repo->savePayment($payment);
        if (($paymentPayload['id'] ?? '') !== '' && ($paymentPayload['externalReference'] ?? '') !== $localId) {
            $this->client->updatePaymentExternalReference($paymentPayload['id'], $localId);
        }
    }

    private function handlePaymentStatusChange(array $paymentPayload): void
    {
        $externalRef = $paymentPayload['externalReference'] ?? '';
        if ($externalRef === '') {
            return;
        }
        try {
            $payment = $this->repo->findPaymentById($externalRef);
        } catch (NotFoundException) {
            return;
        }

        $this->repo->updatePaymentStatus($payment->id, $paymentPayload['status'] ?? '', $paymentPayload['invoiceUrl'] ?? '', $paymentPayload['transactionReceiptUrl'] ?? '');
        $this->issueInvoiceForPayment($payment, $paymentPayload);
    }

    private function issueInvoiceForPayment(PaymentRecord $payment, array $payload): void
    {
        try {
            $this->repo->findInvoiceByPaymentId($payment->id);
            return;
        } catch (NotFoundException) {
            // continue issuing invoice
        }

        $req = [
            'payment' => $payment->id,
            'serviceDescription' => $payment->description !== '' ? $payment->description : 'Pagamento ' . $payment->id,
            'observations' => 'NOTA FISCAL EMITIDA POR EMPRESA OPTANTE DO SIMPLES NACIONAL CONFORME LEI COMPLEMENTAR 123/2006. NÃO GERA DIREITO A CRÉDITO DE I.P.I./ICMS.',
            'externalReference' => $payment->id,
            'value' => $payment->value,
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

    public static function parseDate(string $value): DateTimeImmutable
    {
        $dt = DateTimeImmutable::createFromFormat('Y-m-d', $value, new \DateTimeZone('UTC'));
        if ($dt === false) {
            return new DateTimeImmutable('@0');
        }
        return $dt;
    }

    public static function generateId(): string
    {
        $data = random_bytes(16);
        $data[6] = chr(ord($data[6]) & 0x0f | 0x40);
        $data[8] = chr(ord($data[8]) & 0x3f | 0x80);
        $hex = bin2hex($data);
        return sprintf('%s-%s-%s-%s-%s',
            substr($hex, 0, 8),
            substr($hex, 8, 4),
            substr($hex, 12, 4),
            substr($hex, 16, 4),
            substr($hex, 20, 12)
        );
    }
}
