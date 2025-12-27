<?php

namespace App;

use RuntimeException;

class AsaasClient
{
    public function __construct(
        private readonly string $baseUrl,
        private readonly string $token
    ) {
    }

    private function request(string $method, string $endpoint, ?array $payload = null, array $query = []): array
    {
        $url = rtrim($this->baseUrl, '/') . '/' . ltrim($endpoint, '/');
        if (!empty($query)) {
            $url .= '?' . http_build_query($query);
        }

        $ch = curl_init($url);
        if ($ch === false) {
            throw new RuntimeException('Falha ao inicializar requisição');
        }
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'accept: application/json',
            'access_token: ' . $this->token,
        ]);
        if ($payload !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        }

        $responseBody = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        if ($responseBody === false) {
            $err = curl_error($ch);
            curl_close($ch);
            throw new RuntimeException('Falha na requisição: ' . $err);
        }
        curl_close($ch);

        if ($httpCode >= 400) {
            $decoded = json_decode($responseBody, true);
            if (isset($decoded['errors'][0]['description'])) {
                throw new RuntimeException($decoded['errors'][0]['description']);
            }
            throw new RuntimeException('Erro do Asaas ' . $httpCode . ': ' . $responseBody);
        }

        $data = json_decode($responseBody, true);
        return $data ?? [];
    }

    public function createCustomer(array $payload): array
    {
        return $this->request('POST', 'customers', $payload);
    }

    public function getCustomer(string $externalReference): array
    {
        $resp = $this->request('GET', 'customers', null, ['externalReference' => $externalReference]);
        $data = $resp['data'] ?? [];
        if (empty($data)) {
            throw new RuntimeException("cliente não encontrado para externalReference={$externalReference}");
        }
        return $data[0];
    }

    public function createPayment(array $payload): array
    {
        return $this->request('POST', 'payments', $payload);
    }

    public function getPayment(string $externalReference): array
    {
        $resp = $this->request('GET', 'payments', null, ['externalReference' => $externalReference]);
        $data = $resp['data'] ?? [];
        if (empty($data)) {
            throw new RuntimeException("pagamento não encontrado para externalReference={$externalReference}");
        }
        return $data[0];
    }

    public function updatePaymentExternalReference(string $id, string $externalReference): void
    {
        $this->request('POST', 'payments/' . $id, ['externalReference' => $externalReference]);
    }

    public function createSubscription(array $payload): array
    {
        return $this->request('POST', 'subscriptions', $payload);
    }

    public function getSubscription(string $externalReference): array
    {
        $resp = $this->request('GET', 'subscriptions', null, ['externalReference' => $externalReference]);
        $data = $resp['data'] ?? [];
        if (empty($data)) {
            throw new RuntimeException("assinatura não encontrada para externalReference={$externalReference}");
        }
        return $data[0];
    }

    public function getSubscriptionById(string $id): array
    {
        return $this->request('GET', 'subscriptions/' . $id);
    }

    public function cancelSubscription(string $externalReference): array
    {
        $subscription = $this->getSubscription($externalReference);
        return $this->request('DELETE', 'subscriptions/' . $subscription['id']);
    }

    public function createInvoice(array $payload): array
    {
        return $this->request('POST', 'invoices', $payload);
    }

    public function getInvoice(string $externalReference): array
    {
        $resp = $this->request('GET', 'invoices', null, ['externalReference' => $externalReference]);
        $data = $resp['data'] ?? [];
        if (empty($data)) {
            throw new RuntimeException("nota fiscal não encontrada para externalReference={$externalReference}");
        }
        return $data[0];
    }
}
