# Asaas Integration

Este serviço implementa integração com a API do Asaas para cadastro de clientes, geração de cobranças, assinaturas e emissão automática de notas fiscais a partir de webhooks.

## Requisitos
- Go 1.22
- PostgreSQL acessível via `DATABASE_URL`

## Configuração
Defina as variáveis de ambiente abaixo antes de iniciar o serviço:

- `DATABASE_URL`: string de conexão PostgreSQL.
- `PORT`: porta HTTP do serviço (padrão `8080`).
- `ASAAS_API_KEY`: token de API usado pelo cliente HTTP do Asaas.
- `ASAAS_ENVIRONMENT`: ambiente do Asaas (`sandbox` ou `production`).
- `ASAAS_WEBHOOK_TOKEN`: token esperado no header `asaas-access-token` dos webhooks.

Opcionalmente crie um arquivo `.env` com as variáveis acima; ele será carregado automaticamente na inicialização.

## Execução

```bash
go run ./...
```

O serviço sobe um HTTP server com rotas JSON sob `/customers`, `/payments`, `/subscriptions`, `/invoices` e `/webhooks/asaas`. Os endpoints utilizam IDs UUID armazenados no banco como referência para o Asaas.

## Webhooks
- A rota `/webhooks/asaas` aceita apenas `POST` e exige o header `asaas-access-token` igual a `ASAAS_WEBHOOK_TOKEN`.
- Eventos `PAYMENT_CREATED` ou pagamentos desconhecidos são ignorados.
- Eventos de pagamento recebido/confirmado/atrasado atualizam o status local e disparam emissão de nota fiscal se ainda não existir.

### Notas fiscais automáticas
Para pagamentos com status atualizado via webhook e ainda sem nota, é enviada ao Asaas a seguinte estrutura padrão (com os valo
res dinâmicos entre chaves):

```json
{
  "payment": "{id do pagamento}",
  "serviceDescription": "{descrição do pagamento}",
  "observations": "NOTA FISCAL EMITIDA POR EMPRESA OPTANTE DO SIMPLES NACIONAL CONFORME LEI COMPLEMENTAR 123/2006. NÃO GERA DIREITO A CRÉDITO DE I.P.I./ICMS.",
  "externalReference": "{referencia externa do pagamento}",
  "value": {valor do pagamento},
  "deductions": 0,
  "effectiveDate": "{data de hoje (yyyy-mm-dd)}",
  "municipalServiceCode": "01.03.01",
  "municipalServiceName": "Processamento, armazenamento ou hospedagem de dados, textos, imagens, vídeos, páginas eletrônicas, aplicativos e sistemas de informação, entre outros formatos, e congêneres",
  "updatePayment": true,
  "taxes": {
    "retainIss": false,
    "cofins": 0,
    "csll": 0,
    "inss": 0,
    "ir": 0,
    "pis": 0,
    "iss": 5
  }
}
```

## Swagger
A documentação OpenAPI está disponível em `/swagger/` quando o servidor está rodando.
