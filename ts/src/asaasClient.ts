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

  async createCustomer(payload: CustomerRequest): Promise<CustomerResponse> {
    const { data } = await this.http.post('/customers', payload);
    return data;
  }

  async getCustomer(id: string): Promise<CustomerResponse> {
    const { data } = await this.http.get(`/customers/${id}`);
    return data;
  }

  async createPayment(payload: PaymentRequest): Promise<PaymentResponse> {
    const { data } = await this.http.post('/payments', payload);
    return data;
  }

  async getPayment(id: string): Promise<PaymentResponse> {
    const { data } = await this.http.get(`/payments/${id}`);
    return data;
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
    const { data } = await this.http.get(`/invoices/${id}`);
    return data;
  }

  async listInvoices(customerId: string): Promise<InvoiceResponse[]> {
    const { data } = await this.http.get('/invoices', { params: { customer: customerId } });
    return data?.data ?? [];
  }
}
