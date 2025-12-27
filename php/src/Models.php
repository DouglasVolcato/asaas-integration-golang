<?php

namespace App;

class CustomerRecord
{
    public function __construct(
        public string $id,
        public string $name,
        public string $email,
        public string $cpfCnpj,
        public string $phone,
        public string $mobilePhone,
        public string $address,
        public string $addressNumber,
        public string $complement,
        public string $province,
        public string $postalCode,
        public bool $notificationDisabled,
        public string $additionalEmails,
        public \DateTimeImmutable $createdAt,
        public \DateTimeImmutable $updatedAt
    ) {
    }
}

class PaymentRecord
{
    public function __construct(
        public string $id,
        public string $customerId,
        public ?string $subscriptionId,
        public string $billingType,
        public float $value,
        public \DateTimeImmutable $dueDate,
        public string $description,
        public int $installmentCount,
        public string $callbackSuccessUrl,
        public bool $callbackAutoRedirect,
        public string $status,
        public string $invoiceUrl,
        public string $transactionReceiptUrl,
        public \DateTimeImmutable $createdAt,
        public \DateTimeImmutable $updatedAt
    ) {
    }
}

class SubscriptionRecord
{
    public function __construct(
        public string $id,
        public string $customerId,
        public string $billingType,
        public string $status,
        public float $value,
        public string $cycle,
        public \DateTimeImmutable $nextDueDate,
        public string $description,
        public ?\DateTimeImmutable $endDate,
        public int $maxPayments,
        public \DateTimeImmutable $createdAt,
        public \DateTimeImmutable $updatedAt
    ) {
    }
}

class InvoiceRecord
{
    public function __construct(
        public string $id,
        public string $paymentId,
        public string $serviceDescription,
        public string $observations,
        public float $value,
        public float $deductions,
        public \DateTimeImmutable $effectiveDate,
        public string $municipalServiceId,
        public string $municipalServiceCode,
        public string $municipalServiceName,
        public bool $updatePayment,
        public bool $taxesRetainISS,
        public float $taxesCofins,
        public float $taxesCsll,
        public float $taxesINSS,
        public float $taxesIR,
        public float $taxesPIS,
        public float $taxesISS,
        public string $status,
        public string $paymentLink,
        public \DateTimeImmutable $createdAt,
        public \DateTimeImmutable $updatedAt
    ) {
    }
}
