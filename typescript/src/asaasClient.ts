import axios, { AxiosInstance } from 'axios';
import {
  CustomerRequest,
  CustomerResponse,
  InvoiceRequest,
  InvoiceResponse,
  PaymentRequest,
  PaymentResponse,
  SubscriptionRequest,
  SubscriptionResponse,
} from './types';

type AsaasListResponse<T> = {
  data: T[];
};

export class AsaasClient {
  private http: AxiosInstance;

  constructor(baseURL: string, token: string) {
    this.http = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        access_token: token,
      },
    });
  }

  private pickFirst<T>(payload: AsaasListResponse<T> | T, notFoundMessage: string): T {
    const list = (payload as AsaasListResponse<T>)?.data;
    if (Array.isArray(list)) {
      const item = list[0];
      if (!item) {
        throw new Error(notFoundMessage);
      }
      return item;
    }
    if (payload && typeof payload === 'object' && 'id' in payload) {
      return payload as T;
    }
    throw new Error(notFoundMessage);
  }

  async createCustomer(payload: CustomerRequest): Promise<CustomerResponse> {
    const { data } = await this.http.post('/customers', payload);
    return data;
  }

  async getCustomer(id: string): Promise<CustomerResponse> {
    const { data } = await this.http.get<AsaasListResponse<CustomerResponse>>('/customers', {
      params: { externalReference: id },
    });
    return this.pickFirst(data, `cliente nao encontrado para externalReference=${id}`);
  }

  async createPayment(payload: PaymentRequest): Promise<PaymentResponse> {
    const { data } = await this.http.post('/payments', payload);
    return data;
  }

  async getPayment(id: string): Promise<PaymentResponse> {
    const { data } = await this.http.get<AsaasListResponse<PaymentResponse>>('/payments', {
      params: { externalReference: id },
    });
    return this.pickFirst(data, `pagamento nao encontrado para externalReference=${id}`);
  }

  async updatePaymentExternalReference(id: string, externalReference: string): Promise<void> {
    await this.http.post(`/payments/${id}`, { externalReference });
  }

  async listPayments(customerId: string): Promise<PaymentResponse[]> {
    const { data } = await this.http.get('/payments', { params: { customer: customerId } });
    return data?.data ?? [];
  }

  async createSubscription(payload: SubscriptionRequest): Promise<SubscriptionResponse> {
    const { data } = await this.http.post('/subscriptions', payload);
    return data;
  }

  async getSubscription(id: string): Promise<SubscriptionResponse> {
    const { data } = await this.http.get<AsaasListResponse<SubscriptionResponse>>('/subscriptions', {
      params: { externalReference: id },
    });
    return this.pickFirst(data, `assinatura nao encontrada para externalReference=${id}`);
  }

  async getSubscriptionById(id: string): Promise<SubscriptionResponse> {
    const { data } = await this.http.get(`/subscriptions/${id}`);
    return data;
  }

  async listSubscriptions(customerId: string): Promise<SubscriptionResponse[]> {
    const { data } = await this.http.get('/subscriptions', { params: { customer: customerId } });
    return data?.data ?? [];
  }

  async createInvoice(payload: InvoiceRequest): Promise<InvoiceResponse> {
    const { data } = await this.http.post('/invoices', payload);
    return data;
  }

  async getInvoice(id: string): Promise<InvoiceResponse> {
    const { data } = await this.http.get<AsaasListResponse<InvoiceResponse>>('/invoices', {
      params: { externalReference: id },
    });
    return this.pickFirst(data, `nota fiscal nao encontrada para externalReference=${id}`);
  }

  async listInvoices(customerId: string): Promise<InvoiceResponse[]> {
    const { data } = await this.http.get('/invoices', { params: { customer: customerId } });
    return data?.data ?? [];
  }
}
