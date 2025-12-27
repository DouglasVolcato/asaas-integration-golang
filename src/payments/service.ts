import crypto from "crypto";
import {
  CustomerRequest,
  CustomerRecord,
  CustomerResponse,
  InvoiceRecord,
  InvoiceRequest,
  InvoiceResponse,
  InvoiceTaxes,
  NotificationEvent,
  PaymentRecord,
  PaymentRequest,
  PaymentResponse,
  SubscriptionRecord,
  SubscriptionRequest,
  SubscriptionResponse,
} from "./types";
import { PostgresRepository } from "./repository";
import { AsaasClient } from "./asaasClient";

export class PaymentService {
  constructor(private readonly repo: PostgresRepository, private readonly client: AsaasClient) {}

  async registerCustomer(req: CustomerRequest): Promise<{ local: CustomerRecord; remote: CustomerResponse }> {
    const now = new Date();
    const id = generateId();
    const local: CustomerRecord = {
      id,
      name: req.name,
      email: req.email ?? "",
      cpfCnpj: req.cpfCnpj ?? "",
      phone: req.phone ?? "",
      mobilePhone: req.mobilePhone ?? "",
      address: req.address ?? "",
      addressNumber: req.addressNumber ?? "",
      complement: req.complement ?? "",
      province: req.province ?? "",
      postalCode: req.postalCode ?? "",
      notificationDisabled: req.notificationDisabled ?? false,
      additionalEmails: req.additionalEmails ?? "",
      createdAt: now,
      updatedAt: now,
    };

    const remote = await this.client.createCustomer({ ...req, externalReference: id });
    await this.repo.saveCustomer(local);
    return { local, remote };
  }

  async createPayment(req: PaymentRequest): Promise<{ local: PaymentRecord; remote: PaymentResponse }> {
    const customer = await this.repo.findCustomerById(req.customer);
    if (!customer) {
      throw new Error(`falha ao localizar cliente ${req.customer}`);
    }
    const remoteCustomer = await this.client.getCustomer(customer.id);
    const id = generateId();
    const now = new Date();
    const callback = req.callback ?? { successUrl: "", autoRedirect: false };
    const remote = await this.client.createPayment({ ...req, externalReference: id, customer: remoteCustomer.id });
    const local: PaymentRecord = {
      id,
      customerId: customer.id,
      subscriptionId: null,
      billingType: req.billingType,
      value: req.value,
      dueDate: parseDate(req.dueDate),
      description: req.description ?? "",
      installmentCount: req.installmentCount ?? 0,
      callbackSuccessUrl: callback.successUrl,
      callbackAutoRedirect: callback.autoRedirect,
      status: remote.status,
      invoiceUrl: remote.invoiceUrl ?? "",
      transactionReceiptUrl: remote.transactionReceiptUrl ?? "",
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.savePayment(local);
    return { local, remote };
  }

  async createSubscription(req: SubscriptionRequest): Promise<{ local: SubscriptionRecord; remote: SubscriptionResponse }> {
    const customer = await this.repo.findCustomerById(req.customer);
    if (!customer) {
      throw new Error(`falha ao localizar cliente ${req.customer}`);
    }
    const remoteCustomer = await this.client.getCustomer(customer.id);
    const id = generateId();
    const remote = await this.client.createSubscription({ ...req, externalReference: id, customer: remoteCustomer.id });
    const now = new Date();
    const local: SubscriptionRecord = {
      id,
      customerId: customer.id,
      billingType: req.billingType,
      status: remote.status,
      value: req.value,
      cycle: req.cycle,
      nextDueDate: parseDate(req.nextDueDate),
      description: req.description ?? "",
      endDate: req.endDate ? parseDate(req.endDate) : null,
      maxPayments: req.maxPayments ?? 0,
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.saveSubscription(local);
    return { local, remote };
  }

  async createInvoice(req: InvoiceRequest): Promise<{ local: InvoiceRecord; remote: InvoiceResponse }> {
    const payment = await this.repo.findPaymentById(req.payment);
    if (!payment) {
      throw new Error(`falha ao localizar pagamento ${req.payment}`);
    }
    const remotePayment = await this.client.getPayment(payment.id);
    const id = req.externalReference ?? payment.id;
    const remote = await this.client.createInvoice({ ...req, externalReference: id, payment: remotePayment.id });
    const now = new Date();
    const local: InvoiceRecord = {
      id,
      paymentId: payment.id,
      serviceDescription: req.serviceDescription,
      observations: req.observations,
      value: req.value,
      deductions: req.deductions,
      effectiveDate: parseDate(req.effectiveDate),
      municipalServiceId: req.municipalServiceId ?? "",
      municipalServiceCode: req.municipalServiceCode ?? "",
      municipalServiceName: req.municipalServiceName,
      updatePayment: req.updatePayment ?? false,
      taxesRetainIss: req.taxes.retainIss,
      taxesCofins: req.taxes.cofins,
      taxesCsll: req.taxes.csll,
      taxesInss: req.taxes.inss,
      taxesIr: req.taxes.ir,
      taxesPis: req.taxes.pis,
      taxesIss: req.taxes.iss,
      status: remote.status,
      paymentLink: remote.paymentLink,
      createdAt: now,
      updatedAt: now,
    };
    await this.repo.saveInvoice(local);
    return { local, remote };
  }

  async handleWebhookNotification(event: NotificationEvent): Promise<void> {
    switch (event.event) {
      case "PAYMENT_CREATED":
        if (!event.payment) throw new Error("payload de pagamento ausente");
        if (!event.payment.subscription) return;
        if (event.payment.externalReference) {
          const existing = await this.repo.findPaymentById(event.payment.externalReference);
          if (existing) return;
        }
        const subscription = await this.client.getSubscriptionById(event.payment.subscription);
        if (!subscription.externalReference) {
          throw new Error(`externalReference da assinatura ausente para id ${event.payment.subscription}`);
        }
        const localSubscription = await this.repo.findSubscriptionById(subscription.externalReference);
        if (!localSubscription) return;

        const id = generateId();
        const now = new Date();
        const payment: PaymentRecord = {
          id,
          customerId: localSubscription.customerId,
          subscriptionId: localSubscription.id,
          billingType: event.payment.billingType,
          value: event.payment.value,
          dueDate: parseDate(event.payment.dueDate ?? ""),
          description: event.payment.description ?? "",
          installmentCount: 0,
          callbackSuccessUrl: "",
          callbackAutoRedirect: false,
          status: event.payment.status,
          invoiceUrl: event.payment.invoiceUrl ?? "",
          transactionReceiptUrl: event.payment.transactionReceiptUrl ?? "",
          createdAt: now,
          updatedAt: now,
        };
        await this.repo.savePayment(payment);
        if (event.payment.id && event.payment.externalReference !== id) {
          await this.client.updatePaymentExternalReference(event.payment.id, id);
        }
        return;
      case "INVOICE_CREATED":
      case "SUBSCRIPTION_CREATED":
        return;
      case "PAYMENT_AUTHORIZED":
      case "PAYMENT_APPROVED_BY_RISK_ANALYSIS":
      case "PAYMENT_CONFIRMED":
      case "PAYMENT_ANTICIPATED":
      case "PAYMENT_DELETED":
      case "PAYMENT_REFUNDED":
      case "PAYMENT_REFUND_DENIED":
      case "PAYMENT_CHARGEBACK_REQUESTED":
      case "PAYMENT_AWAITING_CHARGEBACK_REVERSAL":
      case "PAYMENT_DUNNING_REQUESTED":
      case "PAYMENT_CHECKOUT_VIEWED":
      case "PAYMENT_PARTIALLY_REFUNDED":
      case "PAYMENT_SPLIT_DIVERGENCE_BLOCK":
      case "PAYMENT_AWAITING_RISK_ANALYSIS":
      case "PAYMENT_REPROVED_BY_RISK_ANALYSIS":
      case "PAYMENT_UPDATED":
      case "PAYMENT_RECEIVED":
      case "PAYMENT_OVERDUE":
      case "PAYMENT_RESTORED":
      case "PAYMENT_REFUND_IN_PROGRESS":
      case "PAYMENT_RECEIVED_IN_CASH_UNDONE":
      case "PAYMENT_CHARGEBACK_DISPUTE":
      case "PAYMENT_DUNNING_RECEIVED":
      case "PAYMENT_BANK_SLIP_VIEWED":
      case "PAYMENT_CREDIT_CARD_CAPTURE_REFUSED":
      case "PAYMENT_SPLIT_CANCELLED":
      case "PAYMENT_SPLIT_DIVERGENCE_BLOCK_FINISHED": {
        if (!event.payment) throw new Error("payload de pagamento ausente");
        const payment = await this.repo.findPaymentById(event.payment.externalReference);
        if (!payment) return;
        await this.repo.updatePaymentStatus(
          payment.id,
          event.payment.status,
          event.payment.invoiceUrl,
          event.payment.transactionReceiptUrl
        );
        await this.issueInvoiceForPayment(payment, event.payment);
        return;
      }
      case "SUBSCRIPTION_INACTIVATED":
      case "SUBSCRIPTION_SPLIT_DISABLED":
      case "SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK_FINISHED":
      case "SUBSCRIPTION_UPDATED":
      case "SUBSCRIPTION_DELETED":
      case "SUBSCRIPTION_SPLIT_DIVERGENCE_BLOCK": {
        if (!event.subscription) throw new Error("payload de assinatura ausente");
        await this.repo.updateSubscriptionStatus(event.subscription.externalReference, event.subscription.status);
        return;
      }
      case "INVOICE_SYNCHRONIZED":
      case "INVOICE_PROCESSING_CANCELLATION":
      case "INVOICE_CANCELLATION_DENIED":
      case "INVOICE_UPDATED":
      case "INVOICE_AUTHORIZED":
      case "INVOICE_CANCELED":
      case "INVOICE_ERROR": {
        if (!event.invoice) throw new Error("payload de nota fiscal ausente");
        await this.repo.updateInvoiceStatus(event.invoice.externalReference, event.invoice.status);
        return;
      }
      default:
        throw new Error(`tipo de evento não suportado: ${event.event}`);
    }
  }

  private async issueInvoiceForPayment(payment: PaymentRecord, payload: PaymentResponse): Promise<void> {
    const existing = await this.repo.findInvoiceByPaymentId(payment.id);
    if (existing) return;

    const req: InvoiceRequest = {
      payment: payment.id,
      serviceDescription: payment.description || `Pagamento ${payment.id}`,
      observations:
        "NOTA FISCAL EMITIDA POR EMPRESA OPTANTE DO SIMPLES NACIONAL CONFORME LEI COMPLEMENTAR 123/2006. NÃO GERA DIREITO A CRÉDITO DE I.P.I./ICMS.",
      externalReference: payment.id,
      value: payment.value,
      deductions: 0,
      effectiveDate: new Date().toISOString().slice(0, 10),
      municipalServiceCode: "01.03.01",
      municipalServiceName:
        "Processamento, armazenamento ou hospedagem de dados, textos, imagens, vídeos, páginas eletrônicas, aplicativos e sistemas de informação, entre outros formatos, e congêneres",
      updatePayment: true,
      taxes: defaultTaxes(),
    };
    await this.createInvoice(req);
  }
}

export function parseDate(value: string): Date {
  if (!value) return new Date(0);
  return new Date(`${value}T00:00:00Z`);
}

export function generateId(): string {
  const buf = crypto.randomBytes(16);
  buf[6] = (buf[6] & 0x0f) | 0x40;
  buf[8] = (buf[8] & 0x3f) | 0x80;
  const hex = buf.toString("hex");
  return `${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}`;
}

function defaultTaxes(): InvoiceTaxes {
  return {
    retainIss: false,
    cofins: 0,
    csll: 0,
    inss: 0,
    ir: 0,
    pis: 0,
    iss: 5,
  };
}
