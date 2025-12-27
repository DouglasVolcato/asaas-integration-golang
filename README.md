# Asaas Integration

Este repositório contém três implementações do mesmo serviço de integração com a API do Asaas (Go, TypeScript e PHP), com cadastro de clientes, cobranças, assinaturas, emissão de notas fiscais e tratamento de webhooks.

## Estrutura do repositório
- `golang/`: serviço HTTP em Go.
- `typescript/`: serviço HTTP em Node.js/Express (TypeScript).
- `php/`: serviço HTTP em PHP.
- `docker-compose.yml`: PostgreSQL para desenvolvimento local.

## Requisitos
- PostgreSQL acessível via `DATABASE_URL` (ou `docker-compose`).
- Go 1.22 (para `golang/`).
- Node.js e npm (para `typescript/`).
- PHP 8.1+ com Composer e extensões `curl` e `pdo_pgsql` (para `php/`).

## Configuração
Variáveis de ambiente usadas nas três implementações:
- `DATABASE_URL`: string de conexão PostgreSQL.
- `PORT`: porta HTTP do serviço (padrão `8080`).
- `ASAAS_API_TOKEN`: token de API do Asaas.
- `ASAAS_WEBHOOK_TOKEN`: token esperado no header `asaas-access-token` dos webhooks.

Variável específica:
- `ASAAS_API_URL`: obrigatória em `golang/` e `php/`; opcional em `typescript/` (padrão `https://sandbox.asaas.com/api/v3`).

Arquivos `.env`:
- Em `golang/`, o arquivo `.env` precisa existir (a inicialização falha se o arquivo não for encontrado).
- Em `typescript/`, o `.env` é carregado se existir.
- Em `php/`, o `.env` é carregado se existir em `php/.env` ou na raiz do repositório.

## Execução
Se precisar de banco local:

```bash
docker-compose up -d
```

### Go (`golang/`)

```bash
cd golang
go run .
```

### TypeScript (`typescript/`)

```bash
cd typescript
npm install
npm run dev
```

Para produção em TypeScript:

```bash
npm run build
npm start
```

### PHP (`php/`)

```bash
cd php
composer install
php -S 0.0.0.0:${PORT:-8080} -t public
```

## Endpoints
As três implementações usam IDs locais (UUID) como `externalReference` no Asaas e persistem os dados no PostgreSQL.
Os parâmetros `id` e `customer` referenciam o ID local (externalReference), não o ID interno do Asaas.

### Go (`golang/`)
- `POST /customers`
- `GET /customers?id=<id_local>`
- `POST /payments`
- `GET /payments?id=<id_local>`
- `POST /subscriptions`
- `POST /subscriptions/cancel?id=<id_local>`
- `POST /invoices`
- `GET /invoices?id=<id_local>`
- `POST /webhooks/asaas`

### TypeScript (`typescript/`)
- `POST /customers`
- `GET /customers?id=<id_local>`
- `POST /payments`
- `GET /payments?id=<id_local>` ou `GET /payments?customer=<id_local>`
- `POST /subscriptions`
- `GET /subscriptions?id=<id_local>` ou `GET /subscriptions?customer=<id_local>`
- `POST /invoices`
- `GET /invoices?id=<id_local>` ou `GET /invoices?customer=<id_local>`
- `POST /subscriptions/cancel` retorna `501` (não implementado).
- `POST /webhooks/asaas`

### PHP (`php/`)
- `POST /customers`
- `GET /customers?id=<id_local>`
- `POST /payments`
- `GET /payments?id=<id_local>`
- `POST /subscriptions`
- `POST /subscriptions/cancel?id=<id_local>`
- `POST /invoices`
- `GET /invoices?id=<id_local>`
- `POST /webhooks/asaas`

## Webhooks
- A rota `/webhooks/asaas` aceita apenas `POST` e exige o header `asaas-access-token` igual a `ASAAS_WEBHOOK_TOKEN`.
- Eventos `PAYMENT_CREATED` originados de assinaturas criam pagamentos locais quando necessário.
- Eventos de pagamento recebido/confirmado/atrasado atualizam o status local e disparam emissão de nota fiscal se ainda não existir.

## Swagger
A documentação OpenAPI está disponível em `/swagger/` quando o servidor está rodando (arquivos em `golang/swagger`, `typescript/swagger` ou `php/swagger`, dependendo da implementação).
