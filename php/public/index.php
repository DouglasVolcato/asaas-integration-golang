<?php

use AsaasIntegration\{AsaasClient, Config, Database, Env, Repository, Service};

require_once __DIR__ . '/../../vendor/autoload.php';

Env::load(__DIR__ . '/../../.env');

function respondJson($status, $payload): void
{
    http_response_code($status);
    header('Content-Type: application/json');
    echo json_encode($payload);
}

try {
    $config = Config::fromEnv();
    $pdo = Database::connect($config->databaseUrl);
    $repository = new Repository($pdo);
    $repository->ensureSchema();
    $client = new AsaasClient($config->asaasApiUrl, $config->asaasApiToken);
    $service = new Service($repository, $client);
} catch (Throwable $e) {
    respondJson(500, ['error' => $e->getMessage()]);
    exit;
}

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';
$path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH);

if (str_starts_with($path, '/swagger/')) {
    $file = realpath(__DIR__ . '/../../swagger/' . substr($path, 9));
    $base = realpath(__DIR__ . '/../../swagger');
    if ($file && $base && str_starts_with($file, $base) && is_file($file)) {
        $mime = str_ends_with($file, '.yaml') ? 'application/yaml' : 'text/html';
        header('Content-Type: ' . $mime);
        readfile($file);
        exit;
    }
    http_response_code(404);
    exit;
}

function decodeJsonBody(): array
{
    $raw = file_get_contents('php://input');
    if ($raw === false || $raw === '') {
        return [];
    }
    $decoded = json_decode($raw, true);
    if (!is_array($decoded)) {
        throw new RuntimeException('payload inválido');
    }
    return $decoded;
}

function statusForException(Throwable $e): int
{
    if ($e instanceof RuntimeException && str_contains($e->getMessage(), 'not found')) {
        return 404;
    }
    return 502;
}

try {
    switch (trim($path, '/')) {
        case 'customers':
            if ($method === 'POST') {
                $payload = decodeJsonBody();
                $remote = $service->registerCustomer($payload);
                respondJson(201, $remote);
                break;
            }
            if ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respondJson(400, ['error' => 'id é obrigatório']);
                    break;
                }
                $customer = $client->getCustomer($id);
                respondJson(200, $customer);
                break;
            }
            respondJson(405, ['error' => 'método não permitido']);
            break;
        case 'payments':
            if ($method === 'POST') {
                $payload = decodeJsonBody();
                $remote = $service->createPayment($payload);
                respondJson(201, $remote);
                break;
            }
            if ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respondJson(400, ['error' => 'id é obrigatório']);
                    break;
                }
                $payment = $client->getPayment($id);
                respondJson(200, $payment);
                break;
            }
            respondJson(405, ['error' => 'método não permitido']);
            break;
        case 'subscriptions':
            if ($method === 'POST') {
                $payload = decodeJsonBody();
                $remote = $service->createSubscription($payload);
                respondJson(201, $remote);
                break;
            }
            respondJson(405, ['error' => 'método não permitido']);
            break;
        case 'subscriptions/cancel':
            if ($method === 'POST') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respondJson(400, ['error' => 'id é obrigatório']);
                    break;
                }
                $subscription = $client->cancelSubscription($id);
                respondJson(200, $subscription);
                break;
            }
            respondJson(405, ['error' => 'método não permitido']);
            break;
        case 'invoices':
            if ($method === 'POST') {
                $payload = decodeJsonBody();
                $remote = $service->createInvoice($payload);
                respondJson(201, $remote);
                break;
            }
            if ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respondJson(400, ['error' => 'id é obrigatório']);
                    break;
                }
                $invoice = $client->getInvoice($id);
                respondJson(200, $invoice);
                break;
            }
            respondJson(405, ['error' => 'método não permitido']);
            break;
        case 'webhooks/asaas':
            if ($method !== 'POST') {
                respondJson(405, ['error' => 'método não permitido']);
                break;
            }
            $token = $_SERVER['HTTP_ASAAS_ACCESS_TOKEN'] ?? '';
            if ($token !== $config->webhookToken) {
                respondJson(401, ['error' => 'não autorizado']);
                break;
            }
            $payload = decodeJsonBody();
            $service->handleWebhook($payload);
            respondJson(200, ['status' => 'ok']);
            break;
        default:
            respondJson(404, ['error' => 'rota não encontrada']);
    }
} catch (Throwable $e) {
    $status = statusForException($e);
    respondJson($status, ['error' => $e->getMessage()]);
}
