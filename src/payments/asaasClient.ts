import axios, { AxiosInstance } from "axios";
import { URL } from "url";
import {
  CustomerRequest,
  CustomerResponse,
  InvoiceRequest,
  InvoiceResponse,
  NotificationEvent,
  PaymentRequest,
  PaymentResponse,
  SubscriptionRequest,
  SubscriptionResponse,
} from "./types";
import { AsaasConfig } from "./config";

export class AsaasError extends Error {
  statusCode: number;
  code?: string;
  raw?: string;

  constructor(message: string, statusCode: number, code?: string, raw?: string) {
    super(message);
    this.statusCode = statusCode;
    this.code = code;
    this.raw = raw;
  }
}

export class AsaasClient {
  private readonly http: AxiosInstance;

  constructor(private readonly config: AsaasConfig) {
    this.http = axios.create({
      baseURL: config.apiUrl,
      timeout: 30_000,
      headers: {
        "Content-Type": "application/json",
        accept: "application/json",
        access_token: config.apiToken,
      },
    });
  }

  private async request<T>(
    method: string,
    endpoint: string,
    data?: unknown,
    params?: Record<string, string | number>
  ): Promise<T> {
    try {
      const url = new URL(endpoint, this.config.apiUrl);
      const response = await this.http.request<T>({ method, url: url.toString(), data, params });
      return response.data;
    } catch (err: any) {
      if (err.response) {
        const { status, data: respData } = err.response;
        if (respData?.errors?.length) {
          throw new AsaasError(respData.errors[0].description ?? "erro do Asaas", status, respData.errors[0].code);
        }
        throw new AsaasError("erro do Asaas", status, undefined, JSON.stringify(respData));
      }
      throw new AsaasError(err.message, 500);
    }
  }

  async createCustomer(payload: CustomerRequest): Promise<CustomerResponse> {
    return this.request<CustomerResponse>("post", "customers", payload);
  }

  async getCustomer(externalReference: string): Promise<CustomerResponse> {
    const resp = await this.request<{ data: CustomerResponse[] }>("get", "customers", undefined, {
      externalReference,
    });
    if (!resp.data.length) {
      throw new AsaasError(`cliente não encontrado para externalReference=${externalReference}`, 404);
    }
    return resp.data[0];
  }

  async createPayment(payload: PaymentRequest): Promise<PaymentResponse> {
    return this.request<PaymentResponse>("post", "payments", payload);
  }

  async getPayment(externalReference: string): Promise<PaymentResponse> {
    const resp = await this.request<{ data: PaymentResponse[] }>("get", "payments", undefined, {
      externalReference,
    });
    if (!resp.data.length) {
      throw new AsaasError(`pagamento não encontrado para externalReference=${externalReference}`, 404);
    }
    return resp.data[0];
  }

  async updatePaymentExternalReference(id: string, externalReference: string): Promise<void> {
    await this.request<void>("post", `payments/${id}`, { externalReference });
  }

  async createSubscription(payload: SubscriptionRequest): Promise<SubscriptionResponse> {
    return this.request<SubscriptionResponse>("post", "subscriptions", payload);
  }

  async getSubscription(externalReference: string): Promise<SubscriptionResponse> {
    const resp = await this.request<{ data: SubscriptionResponse[] }>("get", "subscriptions", undefined, {
      externalReference,
    });
    if (!resp.data.length) {
      throw new AsaasError(`assinatura não encontrada para externalReference=${externalReference}`, 404);
    }
    return resp.data[0];
  }

  async getSubscriptionById(id: string): Promise<SubscriptionResponse> {
    return this.request<SubscriptionResponse>("get", `subscriptions/${id}`);
  }

  async cancelSubscription(externalReference: string): Promise<SubscriptionResponse> {
    const subscription = await this.getSubscription(externalReference);
    return this.request<SubscriptionResponse>("delete", `subscriptions/${subscription.id}`);
  }

  async createInvoice(payload: InvoiceRequest): Promise<InvoiceResponse> {
    return this.request<InvoiceResponse>("post", "invoices", payload);
  }

  async getInvoice(externalReference: string): Promise<InvoiceResponse> {
    const resp = await this.request<{ data: InvoiceResponse[] }>("get", "invoices", undefined, {
      externalReference,
    });
    if (!resp.data.length) {
      throw new AsaasError(`nota fiscal não encontrada para externalReference=${externalReference}`, 404);
    }
    return resp.data[0];
  }
}

export type WebhookEvent = NotificationEvent;
