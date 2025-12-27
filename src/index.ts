import bodyParser from "body-parser";
import dotenv from "dotenv";
import express, { NextFunction, Request, Response } from "express";
import morgan from "morgan";
import path from "path";
import { Pool } from "pg";
import { AsaasClient } from "./payments/asaasClient";
import { loadConfigFromEnv } from "./payments/config";
import { PostgresRepository } from "./payments/repository";
import { PaymentService } from "./payments/service";
import {
  CustomerRequest,
  InvoiceRequest,
  NotificationEvent,
  PaymentRequest,
  SubscriptionRequest,
} from "./payments/types";

dotenv.config();

const app = express();
app.use(bodyParser.json());
app.use(morgan("dev"));

const config = loadConfigFromEnv();
const pool = new Pool({ connectionString: config.databaseUrl });
const repo = new PostgresRepository(pool);
const client = new AsaasClient(config.asaas);
const service = new PaymentService(repo, client);

repo.ensureSchema().catch((err: Error) => {
  console.error("failed to run migrations", err);
  process.exit(1);
});

app.post("/customers", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const payload = req.body as CustomerRequest;
    const { remote } = await service.registerCustomer(payload);
    res.status(201).json(remote);
  } catch (err) {
    next(err);
  }
});

app.get("/customers", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const id = req.query.id as string;
    if (!id) return res.status(400).json({ error: "id é obrigatório" });
    const customer = await client.getCustomer(id);
    res.json(customer);
  } catch (err) {
    next(err);
  }
});

app.post("/payments", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const payload = req.body as PaymentRequest;
    const { remote } = await service.createPayment(payload);
    res.status(201).json(remote);
  } catch (err) {
    next(err);
  }
});

app.get("/payments", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const id = req.query.id as string;
    if (!id) return res.status(400).json({ error: "id é obrigatório" });
    const payment = await client.getPayment(id);
    res.json(payment);
  } catch (err) {
    next(err);
  }
});

app.post("/subscriptions", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const payload = req.body as SubscriptionRequest;
    const { remote } = await service.createSubscription(payload);
    res.status(201).json(remote);
  } catch (err) {
    next(err);
  }
});

app.post("/subscriptions/cancel", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const id = req.query.id as string;
    if (!id) return res.status(400).json({ error: "id é obrigatório" });
    const subscription = await client.cancelSubscription(id);
    res.json(subscription);
  } catch (err) {
    next(err);
  }
});

app.get("/subscriptions", (_req: Request, res: Response) => {
  res.status(405).json({ error: "método não permitido" });
});

app.post("/invoices", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const payload = req.body as InvoiceRequest;
    const { remote } = await service.createInvoice(payload);
    res.status(201).json(remote);
  } catch (err) {
    next(err);
  }
});

app.get("/invoices", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const id = req.query.id as string;
    if (!id) return res.status(400).json({ error: "id é obrigatório" });
    const invoice = await client.getInvoice(id);
    res.json(invoice);
  } catch (err) {
    next(err);
  }
});

app.post("/webhooks/asaas", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const expectedToken = process.env.ASAAS_WEBHOOK_TOKEN;
    if (!expectedToken || req.headers["asaas-access-token"] !== expectedToken) {
      return res.status(401).json({ error: "não autorizado" });
    }
    const event = req.body as NotificationEvent;
    await service.handleWebhookNotification(event);
    res.sendStatus(200);
  } catch (err) {
    next(err);
  }
});

app.use("/swagger", express.static(path.join(__dirname, "..", "swagger")));

app.use((err: any, _req: Request, res: Response, _next: NextFunction) => {
  console.error(err);
  if (err.statusCode) {
    res.status(err.statusCode).json({ error: err.message });
  } else {
    res.status(502).json({ error: err.message || "erro interno do servidor" });
  }
});

app.listen(config.port, () => {
  console.log(`server listening on :${config.port}`);
});
