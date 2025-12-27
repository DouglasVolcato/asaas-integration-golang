import dotenv from 'dotenv';

dotenv.config();

export interface AppConfig {
  port: number;
  databaseUrl: string;
  asaasApiUrl: string;
  asaasApiToken: string;
  webhookToken: string;
}

export function loadConfig(): AppConfig {
  const {
    PORT,
    DATABASE_URL,
    ASAAS_API_URL,
    ASAAS_API_TOKEN,
    ASAAS_WEBHOOK_TOKEN,
  } = process.env;

  if (!DATABASE_URL) {
    throw new Error('DATABASE_URL não está definida');
  }
  if (!ASAAS_API_TOKEN) {
    throw new Error('ASAAS_API_TOKEN não está definida');
  }

  return {
    port: Number(PORT || 8080),
    databaseUrl: DATABASE_URL,
    asaasApiUrl: ASAAS_API_URL || 'https://sandbox.asaas.com/api/v3',
    asaasApiToken: ASAAS_API_TOKEN,
    webhookToken: ASAAS_WEBHOOK_TOKEN || '',
  };
}
