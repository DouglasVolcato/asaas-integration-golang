import { v4 as uuidv4 } from 'uuid';
import { Pool } from 'pg';
import { AsaasClient } from './asaasClient';
import { PostgresRepository } from './repository';
import {
  CustomerRecord,
  CustomerRequest,
  CustomerResponse,
  InvoiceRecord,
  InvoiceRequest,
  InvoiceResponse,
  PaymentRecord,
  PaymentRequest,
  PaymentResponse,
  SubscriptionRecord,
  SubscriptionRequest,
  SubscriptionResponse,
} from './types';

export class PaymentService {
  private repo: PostgresRepository;
  constructor(private pool: Pool, private client: AsaasClient) {
    this.repo = new PostgresRepository(pool);
  }

  async init(): Promise<void> {
    await this.repo.ensureSchema();
  }

  async registerCustomer(payload: CustomerRequest): Promise<{ local: CustomerRecord; remote: CustomerResponse }> {
    const now = new Date();
    const local: CustomerRecord = {
      id: uuidv4(),
      createdAt: now,
      updatedAt: now,
      ...payload,
    };
    const remote = await this.client.createCustomer({ ...payload, externalReference: local.id });
    await this.repo.saveCustomer(local);
    return { local, remote };
  }

  async createPayment(payload: PaymentRequest): Promise<{ local: PaymentRecord; remote: PaymentResponse }> {
    const customer = await this.repo.findCustomerById(payload.customer);
    if (!customer) {
      throw new Error(`falha ao localizar cliente ${payload.customer}`);
    }
    const remoteCustomer = await this.client.getCustomer(customer.id);
    const id = uuidv4();
    const remote = await this.client.createPayment({ ...payload, customer: remoteCustomer.id, externalReference: id });

    const now = new Date();
    const local: PaymentRecord = {
      ...payload,
      id,
      customerId: customer.id,
      subscriptionId: payload.externalReference,
      callbackSuccessUrl: payload.callback?.successUrl ?? '',
      callbackAutoRedirect: payload.callback?.autoRedirect ?? false,
      status: remote.status,
      invoiceUrl: remote.invoiceUrl,
      transactionReceiptUrl: remote.transactionReceiptUrl,
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.savePayment(local);
    return { local, remote };
  }

  async createSubscription(
    payload: SubscriptionRequest,
  ): Promise<{ local: SubscriptionRecord; remote: SubscriptionResponse }> {
    const customer = await this.repo.findCustomerById(payload.customer);
    if (!customer) {
      throw new Error(`falha ao localizar cliente ${payload.customer}`);
    }
    const remoteCustomer = await this.client.getCustomer(customer.id);
    const id = uuidv4();
    const remote = await this.client.createSubscription({ ...payload, customer: remoteCustomer.id, externalReference: id });

    const now = new Date();
    const local: SubscriptionRecord = {
      ...payload,
      id,
      customerId: customer.id,
      status: remote.status,
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.saveSubscription(local);
    return { local, remote };
  }

  async createInvoice(payload: InvoiceRequest): Promise<{ local: InvoiceRecord; remote: InvoiceResponse }> {
    const payment = await this.repo.findPaymentById(payload.payment);
    if (!payment) {
      throw new Error(`falha ao localizar pagamento ${payload.payment}`);
    }
    const remotePayment = await this.client.getPayment(payment.id);
    const id = payload.externalReference || payment.id;
    const remote = await this.client.createInvoice({ ...payload, payment: remotePayment.id, externalReference: id });

    const now = new Date();
    const local: InvoiceRecord = {
      ...payload,
      id,
      paymentId: payment.id,
      status: remote.status,
      paymentLink: remote.paymentLink,
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.saveInvoice(local);
    return { local, remote };
  }
}
