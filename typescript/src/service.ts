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
  WebhookPayload,
} from './types';

export class PaymentService {
  private repo: PostgresRepository;
  constructor(private pool: Pool, private client: AsaasClient) {
    this.repo = new PostgresRepository(pool);
  }

  async init(): Promise<void> {
    await this.repo.ensureSchema();
  }

  private async resolveCustomer(
    id: string,
  ): Promise<{ local: CustomerRecord; remote: CustomerResponse }> {
    const local = await this.repo.findCustomerById(id);
    if (!local) {
      throw new Error(`falha ao localizar cliente ${id}`);
    }
    const remote = await this.client.getCustomer(local.id);
    return { local, remote };
  }

  private async resolvePayment(id: string): Promise<{ local: PaymentRecord; remote: PaymentResponse }> {
    const local = await this.repo.findPaymentById(id);
    if (!local) {
      throw new Error(`falha ao localizar pagamento ${id}`);
    }
    const remote = await this.client.getPayment(local.id);
    return { local, remote };
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
    const { local: customer, remote: remoteCustomer } = await this.resolveCustomer(payload.customer);
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
    const { local: customer, remote: remoteCustomer } = await this.resolveCustomer(payload.customer);
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
    const { local: payment, remote: remotePayment } = await this.resolvePayment(payload.payment);
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

  async listPaymentsByCustomer(customerId: string): Promise<PaymentResponse[]> {
    const { remote } = await this.resolveCustomer(customerId);
    return this.client.listPayments(remote.id);
  }

  async listSubscriptionsByCustomer(customerId: string): Promise<SubscriptionResponse[]> {
    const { remote } = await this.resolveCustomer(customerId);
    return this.client.listSubscriptions(remote.id);
  }

  async listInvoicesByCustomer(customerId: string): Promise<InvoiceResponse[]> {
    const { remote } = await this.resolveCustomer(customerId);
    return this.client.listInvoices(remote.id);
  }

  async handleWebhookPayload(payload: string): Promise<void> {
    let event: WebhookPayload;
    try {
      event = JSON.parse(payload) as WebhookPayload;
    } catch {
      throw new Error('payload inválido');
    }
    await this.handleWebhookNotification(event);
  }

  private async handleWebhookNotification(event: WebhookPayload): Promise<void> {
    const eventType = event.event ?? '';
    switch (eventType) {
      case 'PAYMENT_CREATED': {
        if (!event.payment) {
          throw new Error('payload de pagamento ausente');
        }
        if (!event.payment.subscription) {
          return;
        }
        if (event.payment.externalReference) {
          const existingPayment = await this.repo.findPaymentById(event.payment.externalReference);
          if (existingPayment) {
            return;
          }
        }

        let subscription: SubscriptionResponse;
        try {
          subscription = await this.client.getSubscriptionById(event.payment.subscription);
        } catch (err) {
          throw new Error(
            `falha ao buscar assinatura ${event.payment.subscription}: ${this.formatError(err)}`,
          );
        }
        if (!subscription.externalReference) {
          throw new Error(
            `externalReference da assinatura ausente para id ${event.payment.subscription}`,
          );
        }

        const localSubscription = await this.repo.findSubscriptionById(subscription.externalReference);
        if (!localSubscription) {
          return;
        }

        const localId = uuidv4();
        const now = new Date();
        const localPayment: PaymentRecord = {
          id: localId,
          customer: localSubscription.customerId,
          billingType: event.payment.billingType ?? '',
          value: event.payment.value ?? 0,
          dueDate: this.parseDate(event.payment.dueDate),
          description: event.payment.description,
          installmentCount: 0,
          customerId: localSubscription.customerId,
          subscriptionId: localSubscription.id,
          callbackSuccessUrl: '',
          callbackAutoRedirect: false,
          status: event.payment.status ?? '',
          invoiceUrl: event.payment.invoiceUrl,
          transactionReceiptUrl: event.payment.transactionReceiptUrl,
          createdAt: now,
          updatedAt: now,
        };

        try {
          await this.repo.savePayment(localPayment);
        } catch (err) {
          throw new Error(`falha ao salvar pagamento local: ${this.formatError(err)}`);
        }

        if (event.payment.id && event.payment.externalReference !== localId) {
          try {
            await this.client.updatePaymentExternalReference(event.payment.id, localId);
          } catch (err) {
            throw new Error(
              `falha ao atualizar externalReference do pagamento: ${this.formatError(err)}`,
            );
          }
        }
        return;
      }
      case 'INVOICE_CREATED':
      case 'SUBSCRIPTION_CREATED':
        return;
      case 'PAYMENT_AUTHORIZED':
      case 'PAYMENT_APPROVED_BY_RISK_ANALYSIS':
      case 'PAYMENT_CONFIRMED':
      case 'PAYMENT_ANTICIPATED':
      case 'PAYMENT_DELETED':
      case 'PAYMENT_REFUNDED':
      case 'PAYMENT_REFUND_DENIED':
      case 'PAYMENT_CHARGEBACK_REQUESTED':
      case 'PAYMENT_AWAITING_CHARGEBACK_REVERSAL':
      case 'PAYMENT_DUNNING_REQUESTED':
      case 'PAYMENT_CHECKOUT_VIEWED':
      case 'PAYMENT_PARTIALLY_REFUNDED':
      case 'PAYMENT_SPLIT_DIVERGENCE_BLOCK':
      case 'PAYMENT_AWAITING_RISK_ANALYSIS':
      case 'PAYMENT_REPROVED_BY_RISK_ANALYSIS':
      case 'PAYMENT_UPDATED':
      case 'PAYMENT_RECEIVED':
      case 'PAYMENT_OVERDUE':
      case 'PAYMENT_RESTORED':
      case 'PAYMENT_REFUND_IN_PROGRESS':
      case 'PAYMENT_RECEIVED_IN_CASH_UNDONE':
      case 'PAYMENT_CHARGEBACK_DISPUTE':
      case 'PAYMENT_DUNNING_RECEIVED':
      case 'PAYMENT_BANK_SLIP_VIEWED':
      case 'PAYMENT_CREDIT_CARD_CAPTURE_REFUSED':
      case 'PAYMENT_SPLIT_CANCELLED':
      case 'PAYMENT_SPLIT_DIVERGENCE_BLOCK_FINISHED': {
        if (!event.payment) {
          throw new Error('payload de pagamento ausente');
        }
        const externalReference = event.payment.externalReference ?? '';
        if (!externalReference) {
          return;
        }
        const payment = await this.repo.findPaymentById(externalReference);
        if (!payment) {
          return;
        }
        await this.repo.updatePaymentStatus(
          payment.id,
          event.payment.status ?? '',
          event.payment.invoiceUrl ?? '',
          event.payment.transactionReceiptUrl ?? '',
        );
        await this.issueInvoiceForPayment(payment, event.payment);
        return;
      }
      case 'SUBSCRIPTION_INACTIVATED':
      case 'SUBSCRIPTION_SPLIT_DISABLED':
      case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK_FINISHED':
      case 'SUBSCRIPTION_UPDATED':
      case 'SUBSCRIPTION_DELETED':
      case 'SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK':
        if (!event.subscription) {
          throw new Error('payload de assinatura ausente');
        }
        await this.repo.updateSubscriptionStatus(
          event.subscription.externalReference ?? '',
          event.subscription.status ?? '',
        );
        return;
      case 'INVOICE_SYNCHRONIZED':
      case 'INVOICE_PROCESSING_CANCELLATION':
      case 'INVOICE_CANCELLATION_DENIED':
      case 'INVOICE_UPDATED':
      case 'INVOICE_AUTHORIZED':
      case 'INVOICE_CANCELED':
      case 'INVOICE_ERROR':
        if (!event.invoice) {
          throw new Error('payload de nota fiscal ausente');
        }
        await this.repo.updateInvoiceStatus(
          event.invoice.externalReference ?? '',
          event.invoice.status ?? '',
        );
        return;
      default:
        throw new Error(`tipo de evento não suportado: ${eventType}`);
    }
  }

  private async issueInvoiceForPayment(payment: PaymentRecord, _payload: PaymentResponse): Promise<void> {
    const existingInvoice = await this.repo.findInvoiceByPaymentId(payment.id);
    if (existingInvoice) {
      return;
    }

    const request: InvoiceRequest = {
      payment: payment.id,
      serviceDescription: payment.description ? payment.description : `Pagamento ${payment.id}`,
      observations:
        'NOTA FISCAL EMITIDA POR EMPRESA OPTANTE DO SIMPLES NACIONAL CONFORME LEI COMPLEMENTAR 123/2006. NÃO GERA DIREITO A CRÉDITO DE I.P.I./ICMS.',
      externalReference: payment.id,
      value: payment.value,
      deductions: 0,
      effectiveDate: new Date().toISOString().slice(0, 10),
      municipalServiceCode: '01.03.01',
      municipalServiceName:
        'Processamento, armazenamento ou hospedagem de dados, textos, imagens, vídeos, páginas eletrônicas, aplicativos e sistemas de informação, entre outros formatos, e congêneres',
      updatePayment: true,
      taxes: {
        retainIss: false,
        cofins: 0,
        csll: 0,
        inss: 0,
        ir: 0,
        pis: 0,
        iss: 5,
      },
    };

    try {
      await this.createInvoice(request);
    } catch (err) {
      throw new Error(
        `falha ao emitir nota fiscal para o pagamento ${payment.id}: ${this.formatError(err)}`,
      );
    }
  }

  private formatError(error: unknown): string {
    if (error instanceof Error) {
      return error.message;
    }
    return String(error);
  }

  private parseDate(value?: string): string {
    if (!value) {
      return '0001-01-01';
    }
    const parsed = new Date(`${value}T00:00:00Z`);
    if (Number.isNaN(parsed.getTime())) {
      return '0001-01-01';
    }
    return value;
  }
}
