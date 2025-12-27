export interface CustomerRequest {
  name: string;
  email?: string;
  cpfCnpj?: string;
  phone?: string;
  mobilePhone?: string;
  address?: string;
  addressNumber?: string;
  complement?: string;
  province?: string;
  postalCode?: string;
  externalReference?: string;
  notificationDisabled?: boolean;
  additionalEmails?: string;
}

export interface CustomerResponse {
  id: string;
  name: string;
  email: string;
  externalReference: string;
}

export interface PaymentCallback {
  successUrl: string;
  autoRedirect: boolean;
}

export interface PaymentRequest {
  customer: string;
  billingType: string;
  value: number;
  dueDate: string;
  description?: string;
  installmentCount?: number;
  externalReference?: string;
  callback?: PaymentCallback;
}

export interface PaymentResponse {
  id: string;
  customer: string;
  billingType: string;
  value: number;
  status: string;
  description?: string;
  dueDate?: string;
  externalReference: string;
  subscription?: string;
  invoiceUrl?: string;
  transactionReceiptUrl?: string;
}

export interface SubscriptionRequest {
  customer: string;
  billingType: string;
  value: number;
  nextDueDate: string;
  cycle: string;
  externalReference?: string;
  description?: string;
  endDate?: string;
  maxPayments?: number;
}

export interface SubscriptionResponse {
  id: string;
  customer: string;
  status: string;
  value: number;
  externalReference: string;
}

export interface InvoiceTaxes {
  retainIss: boolean;
  cofins: number;
  csll: number;
  inss: number;
  ir: number;
  pis: number;
  iss: number;
}

export interface InvoiceRequest {
  payment: string;
  serviceDescription: string;
  observations: string;
  externalReference?: string;
  value: number;
  deductions: number;
  effectiveDate: string;
  municipalServiceId?: string;
  municipalServiceCode?: string;
  municipalServiceName: string;
  updatePayment?: boolean;
  taxes: InvoiceTaxes;
}

export interface InvoiceResponse {
  id: string;
  customer: string;
  status: string;
  value: number;
  externalReference: string;
  paymentLink: string;
}

export interface NotificationEvent {
  event: string;
  payment?: PaymentResponse;
  invoice?: InvoiceResponse;
  subscription?: SubscriptionResponse;
}

export interface CustomerRecord {
  id: string;
  name: string;
  email: string;
  cpfCnpj: string;
  phone: string;
  mobilePhone: string;
  address: string;
  addressNumber: string;
  complement: string;
  province: string;
  postalCode: string;
  notificationDisabled: boolean;
  additionalEmails: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface PaymentRecord {
  id: string;
  customerId: string;
  subscriptionId: string | null;
  billingType: string;
  value: number;
  dueDate: Date;
  description: string;
  installmentCount: number;
  callbackSuccessUrl: string;
  callbackAutoRedirect: boolean;
  status: string;
  invoiceUrl: string;
  transactionReceiptUrl: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface SubscriptionRecord {
  id: string;
  customerId: string;
  billingType: string;
  status: string;
  value: number;
  cycle: string;
  nextDueDate: Date;
  description: string;
  endDate: Date | null;
  maxPayments: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface InvoiceRecord {
  id: string;
  paymentId: string;
  serviceDescription: string;
  observations: string;
  value: number;
  deductions: number;
  effectiveDate: Date;
  municipalServiceId: string;
  municipalServiceCode: string;
  municipalServiceName: string;
  updatePayment: boolean;
  taxesRetainIss: boolean;
  taxesCofins: number;
  taxesCsll: number;
  taxesInss: number;
  taxesIr: number;
  taxesPis: number;
  taxesIss: number;
  status: string;
  paymentLink: string;
  createdAt: Date;
  updatedAt: Date;
}
