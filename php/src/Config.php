<?php

namespace App;

class Config
{
    public function __construct(
        public readonly string $port,
        public readonly string $databaseUrl,
        public readonly string $asaasApiUrl,
        public readonly string $asaasApiToken,
        public readonly string $webhookToken
    ) {
    }

    public static function load(): self
    {
        $port = getenv('PORT') ?: '8080';
        $databaseUrl = getenv('DATABASE_URL');
        $asaasApiUrl = getenv('ASAAS_API_URL');
        $asaasApiToken = getenv('ASAAS_API_TOKEN');
        $webhookToken = getenv('ASAAS_WEBHOOK_TOKEN') ?: '';

        if (!$databaseUrl) {
            throw new \RuntimeException('DATABASE_URL não está definida');
        }
        if (!$asaasApiUrl) {
            throw new \RuntimeException('ASAAS_API_URL não está definida');
        }
        if (!$asaasApiToken) {
            throw new \RuntimeException('ASAAS_API_TOKEN não está definido');
        }

        return new self($port, $databaseUrl, $asaasApiUrl, $asaasApiToken, $webhookToken);
    }
}
