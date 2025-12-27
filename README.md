# Asaas Integration

Este repositório contém duas implementações do mesmo serviço de integração com a API do Asaas (Go e TypeScript), com cadastro de clientes, cobranças, assinaturas, emissão de notas fiscais e tratamento de webhooks.

## Estrutura do repositório
- `golang/`: serviço HTTP em Go.
- `typescript/`: serviço HTTP em Node.js/Express (TypeScript).
- `docker-compose.yml`: PostgreSQL para desenvolvimento local.

## Requisitos
- PostgreSQL acessível via `DATABASE_URL` (ou `docker-compose`).
- Go 1.22 (para `golang/`).
- Node.js e npm (para `typescript/`).

## Configuração
Variáveis de ambiente usadas nas duas implementações:
- `DATABASE_URL`: string de conexão PostgreSQL.
- `PORT`: porta HTTP do serviço (padrão `8080`).
- `ASAAS_API_TOKEN`: token de API do Asaas.
- `ASAAS_WEBHOOK_TOKEN`: token esperado no header `asaas-access-token` dos webhooks.

Variável específica:
- `ASAAS_API_URL`: obrigatória em `golang/`; opcional em `typescript/` (padrão `https://sandbox.asaas.com/api/v3`).

Arquivos `.env`:
- Em `golang/`, o arquivo `.env` precisa existir (a inicialização falha se o arquivo não for encontrado).
- Em `typescript/`, o `.env` é carregado se existir.

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

## Endpoints
As duas implementações usam IDs locais (UUID) como `externalReference` no Asaas e persistem os dados no PostgreSQL.

### Go (`golang/`)
- `POST /customers`
- `GET /customers?id=<id_asaas>`
- `POST /payments`
- `GET /payments?id=<id_asaas>`
- `POST /subscriptions`
- `POST /subscriptions/cancel?id=<id_asaas>`
- `POST /invoices`
- `GET /invoices?id=<id_asaas>`
- `POST /webhooks/asaas`

### TypeScript (`typescript/`)
- `POST /customers`
- `GET /customers?id=<id_asaas>`
- `POST /payments`
- `GET /payments?id=<id_asaas>` ou `GET /payments?customer=<id_local>`
- `POST /subscriptions`
- `GET /subscriptions?id=<id_asaas>` ou `GET /subscriptions?customer=<id_local>`
- `POST /invoices`
- `GET /invoices?id=<id_asaas>` ou `GET /invoices?customer=<id_local>`
- `POST /subscriptions/cancel` retorna `501` (não implementado).
- `POST /webhooks/asaas`

### Implementação PHP

Uma versão equivalente em PHP está disponível em `php/` com o mesmo Swagger e lógica de webhooks. Para executar:

```bash
cd php
composer install
php -S 0.0.0.0:${PORT:-8080} -t public
```

Certifique-se de configurar as mesmas variáveis de ambiente usadas pela versão Go (`DATABASE_URL`, `ASAAS_API_URL`, `ASAAS_API_TOKEN`, `ASAAS_WEBHOOK_TOKEN` e `PORT`).

## Webhooks
- A rota `/webhooks/asaas` aceita apenas `POST` e exige o header `asaas-access-token` igual a `ASAAS_WEBHOOK_TOKEN`.
- Eventos `PAYMENT_CREATED` originados de assinaturas criam pagamentos locais quando necessário.
- Eventos de pagamento recebido/confirmado/atrasado atualizam o status local e disparam emissão de nota fiscal se ainda não existir.

## Swagger
A documentação OpenAPI está disponível em `/swagger/` quando o servidor está rodando (arquivos em `golang/swagger` ou `typescript/swagger`, dependendo da implementação).
