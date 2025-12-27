export interface AsaasConfig {
  apiUrl: string;
  apiToken: string;
}

export interface AppConfig {
  port: number;
  databaseUrl: string;
  asaas: AsaasConfig;
}

export function loadConfigFromEnv(): AppConfig {
  const apiUrl = process.env.ASAAS_API_URL;
  const apiToken = process.env.ASAAS_API_TOKEN;
  const databaseUrl = process.env.DATABASE_URL;
  const port = process.env.PORT ? parseInt(process.env.PORT, 10) : 8080;

  if (!apiUrl) {
    throw new Error("ASAAS_API_URL não está definida");
  }
  if (!apiToken) {
    throw new Error("ASAAS_API_TOKEN não está definido");
  }
  if (!databaseUrl) {
    throw new Error("DATABASE_URL não está definida");
  }

  return {
    port,
    databaseUrl,
    asaas: { apiUrl, apiToken },
  };
}
