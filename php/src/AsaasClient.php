<?php

namespace AsaasIntegration;

class AsaasClient
{
    private const TIMEOUT = 30;

    public function __construct(
        private readonly string $baseUrl,
        private readonly string $token
    ) {
    }

    private function request(string $method, string $endpoint, ?array $query, $payload, bool $expectBody = true): array
    {
        $url = rtrim($this->baseUrl, '/') . '/' . ltrim($endpoint, '/');
        if ($query) {
            $url .= '?' . http_build_query($query);
        }

        $ch = curl_init($url);
        $headers = [
            'Content-Type: application/json',
            'accept: application/json',
            'access_token: ' . $this->token,
        ];

        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => self::TIMEOUT,
            CURLOPT_CUSTOMREQUEST => strtoupper($method),
            CURLOPT_HTTPHEADER => $headers,
        ]);

        if ($payload !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        }

        $raw = curl_exec($ch);
        if ($raw === false) {
            $err = curl_error($ch);
            curl_close($ch);
            throw new \RuntimeException('erro de rede: ' . $err);
        }

        $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($status >= 400) {
            throw new \RuntimeException('erro do Asaas ' . $status . ': ' . $raw);
        }

        if (!$expectBody) {
            return [];
        }

        $decoded = json_decode($raw, true);
        if (!is_array($decoded)) {
            throw new \RuntimeException('resposta inesperada do Asaas');
        }
        return $decoded;
    }

    public function createCustomer(array $req): array
    {
        return $this->request('POST', 'customers', null, $req);
    }

    public function getCustomer(string $externalReference): array
    {
        $resp = $this->request('GET', 'customers', ['externalReference' => $externalReference], null);
        return $resp['data'][0] ?? throw new \RuntimeException('cliente não encontrado para externalReference=' . $externalReference);
    }

    public function createPayment(array $req): array
    {
        return $this->request('POST', 'payments', null, $req);
    }

    public function getPayment(string $externalReference): array
    {
        $resp = $this->request('GET', 'payments', ['externalReference' => $externalReference], null);
        return $resp['data'][0] ?? throw new \RuntimeException('pagamento não encontrado para externalReference=' . $externalReference);
    }

    public function updatePaymentExternalReference(string $id, string $externalReference): void
    {
        $this->request('POST', 'payments/' . $id, null, ['externalReference' => $externalReference]);
    }

    public function createSubscription(array $req): array
    {
        return $this->request('POST', 'subscriptions', null, $req);
    }

    public function getSubscription(string $externalReference): array
    {
        $resp = $this->request('GET', 'subscriptions', ['externalReference' => $externalReference], null);
        return $resp['data'][0] ?? throw new \RuntimeException('assinatura não encontrada para externalReference=' . $externalReference);
    }

    public function getSubscriptionById(string $id): array
    {
        return $this->request('GET', 'subscriptions/' . $id, null, null);
    }

    public function cancelSubscription(string $externalReference): array
    {
        $subscription = $this->getSubscription($externalReference);
        return $this->request('DELETE', 'subscriptions/' . $subscription['id'], null, null);
    }

    public function createInvoice(array $req): array
    {
        return $this->request('POST', 'invoices', null, $req);
    }

    public function getInvoice(string $externalReference): array
    {
        $resp = $this->request('GET', 'invoices', ['externalReference' => $externalReference], null);
        return $resp['data'][0] ?? throw new \RuntimeException('nota fiscal não encontrada para externalReference=' . $externalReference);
    }
}
