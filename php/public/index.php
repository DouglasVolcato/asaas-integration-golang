<?php

require_once __DIR__ . '/../vendor/autoload.php';

use App\{Env, Config, Database, Repository, AsaasClient, Service, BadRequestException, NotFoundException};

Env::load();

try {
    $config = Config::load();
    $pdo = Database::connect($config->databaseUrl);
    $repo = new Repository($pdo);
    $repo->ensureSchema();
    $client = new AsaasClient($config->asaasApiUrl, $config->asaasApiToken);
    $service = new Service($repo, $client);
} catch (Throwable $e) {
    http_response_code(500);
    header('Content-Type: application/json');
    echo json_encode(['error' => $e->getMessage()]);
    exit;
}

$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH) ?? '/';

if (str_starts_with($path, '/swagger/')) {
    $swaggerDir = realpath(__DIR__ . '/../../swagger');
    if ($swaggerDir && file_exists($swaggerDir . substr($path, 8))) {
        $file = $swaggerDir . substr($path, 8);
        $ext = pathinfo($file, PATHINFO_EXTENSION);
        $mime = $ext === 'yaml' || $ext === 'yml' ? 'application/yaml' : 'text/html';
        header('Content-Type: ' . $mime);
        readfile($file);
        exit;
    }
    http_response_code(404);
    exit;
}

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';

function jsonInput(): array
{
    $raw = file_get_contents('php://input');
    if ($raw === false || trim($raw) === '') {
        throw new BadRequestException('payload inválido');
    }
    $data = json_decode($raw, true);
    if (!is_array($data)) {
        throw new BadRequestException('payload inválido');
    }
    return $data;
}

function respond(mixed $payload, int $status = 200): void
{
    http_response_code($status);
    header('Content-Type: application/json');
    echo json_encode($payload);
}

try {
    switch (true) {
        case $path === '/customers' || str_starts_with($path, '/customers/'):
            if ($method === 'POST') {
                $body = jsonInput();
                $remote = $service->registerCustomer($body);
                respond($remote, 201);
            } elseif ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respond(['error' => 'id é obrigatório'], 400);
                    break;
                }
                $customer = $client->getCustomer($id);
                respond($customer);
            } else {
                respond(['error' => 'método não permitido'], 405);
            }
            break;
        case $path === '/payments' || str_starts_with($path, '/payments/'):
            if ($method === 'POST') {
                $body = jsonInput();
                $remote = $service->createPayment($body);
                respond($remote, 201);
            } elseif ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respond(['error' => 'id é obrigatório'], 400);
                    break;
                }
                $payment = $client->getPayment($id);
                respond($payment);
            } else {
                respond(['error' => 'método não permitido'], 405);
            }
            break;
        case $path === '/subscriptions' || str_starts_with($path, '/subscriptions/'):
            if ($path === '/subscriptions/cancel' || str_starts_with($path, '/subscriptions/cancel/')) {
                if ($method !== 'POST') {
                    respond(['error' => 'método não permitido'], 405);
                    break;
                }
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respond(['error' => 'id é obrigatório'], 400);
                    break;
                }
                $subscription = $client->cancelSubscription($id);
                respond($subscription);
                break;
            }

            if ($method === 'POST') {
                $body = jsonInput();
                $remote = $service->createSubscription($body);
                respond($remote, 201);
            } else {
                respond(['error' => 'método não permitido'], 405);
            }
            break;
        case $path === '/invoices' || str_starts_with($path, '/invoices/'):
            if ($method === 'POST') {
                $body = jsonInput();
                $remote = $service->createInvoice($body);
                respond($remote, 201);
            } elseif ($method === 'GET') {
                $id = $_GET['id'] ?? '';
                if ($id === '') {
                    respond(['error' => 'id é obrigatório'], 400);
                    break;
                }
                $invoice = $client->getInvoice($id);
                respond($invoice);
            } else {
                respond(['error' => 'método não permitido'], 405);
            }
            break;
        case $path === '/webhooks/asaas':
            if ($method !== 'POST') {
                respond(['error' => 'método não permitido'], 405);
                break;
            }
            if ($config->webhookToken === '' || ($_SERVER['HTTP_ASAAS_ACCESS_TOKEN'] ?? '') !== $config->webhookToken) {
                respond(['error' => 'não autorizado'], 401);
                break;
            }
            $body = file_get_contents('php://input') ?: '';
            $data = json_decode($body, true);
            if (!is_array($data)) {
                respond(['error' => 'payload inválido'], 400);
                break;
            }
            try {
                $service->handleWebhook($data);
            } catch (Throwable $e) {
                respond(['error' => $e->getMessage()], 400);
                break;
            }
            respond(['status' => 'ok'], 200);
            break;
        default:
            respond(['error' => 'rota não encontrada'], 404);
            break;
    }
} catch (BadRequestException $e) {
    respond(['error' => $e->getMessage()], 400);
} catch (NotFoundException $e) {
    respond(['error' => $e->getMessage()], 404);
} catch (RuntimeException $e) {
    respond(['error' => $e->getMessage()], 502);
} catch (Throwable $e) {
    respond(['error' => $e->getMessage()], 502);
}
