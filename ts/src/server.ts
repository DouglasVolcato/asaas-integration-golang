import express from 'express';
import morgan from 'morgan';
import { Pool } from 'pg';
import path from 'path';
import { loadConfig } from './config';
import { AsaasClient } from './asaasClient';
import { PaymentService } from './service';
import { InvoiceRequest, PaymentRequest, SubscriptionRequest } from './types';

async function bootstrap() {
  const config = loadConfig();
  const pool = new Pool({ connectionString: config.databaseUrl });
  const client = new AsaasClient(config.asaasApiUrl, config.asaasApiToken);
  const service = new PaymentService(pool, client);
  await service.init();

  const app = express();
  app.use(express.json());
  app.use(morgan('combined'));

  app.post('/customers', async (req, res) => {
    try {
      const { remote } = await service.registerCustomer(req.body);
      res.status(201).json(remote);
    } catch (err: any) {
      res.status(400).json({ error: err.message });
    }
  });

  app.get('/customers', async (req, res) => {
    const id = req.query.id as string;
    if (!id) return res.status(400).json({ error: 'id é obrigatório' });
    try {
      const customer = await client.getCustomer(id);
      res.json(customer);
    } catch (err: any) {
      res.status(502).json({ error: err.message });
    }
  });

  app.post('/payments', async (req, res) => {
    try {
      const { remote } = await service.createPayment(req.body as PaymentRequest);
      res.status(201).json(remote);
    } catch (err: any) {
      res.status(400).json({ error: err.message });
    }
  });

  app.get('/payments', async (req, res) => {
    const id = req.query.id as string;
    const customer = req.query.customer as string;
    if (id) {
      try {
        const payment = await client.getPayment(id);
        res.json(payment);
      } catch (err: any) {
        res.status(502).json({ error: err.message });
      }
      return;
    }
    if (!customer) return res.status(400).json({ error: 'customer é obrigatório' });
    try {
      const payments = await client.listPayments(customer);
      res.json(payments);
    } catch (err: any) {
      res.status(502).json({ error: err.message });
    }
  });

  app.post('/subscriptions', async (req, res) => {
    try {
      const { remote } = await service.createSubscription(req.body as SubscriptionRequest);
      res.status(201).json(remote);
    } catch (err: any) {
      res.status(400).json({ error: err.message });
    }
  });

  app.get('/subscriptions', async (req, res) => {
    const id = req.query.id as string;
    const customer = req.query.customer as string;
    if (id) {
      try {
        const subscription = await client.getSubscription(id);
        res.json(subscription);
      } catch (err: any) {
        res.status(502).json({ error: err.message });
      }
      return;
    }
    if (!customer) return res.status(400).json({ error: 'customer é obrigatório' });
    try {
      const subscriptions = await client.listSubscriptions(customer);
      res.json(subscriptions);
    } catch (err: any) {
      res.status(502).json({ error: err.message });
    }
  });

  app.post('/invoices', async (req, res) => {
    try {
      const { remote } = await service.createInvoice(req.body as InvoiceRequest);
      res.status(201).json(remote);
    } catch (err: any) {
      res.status(400).json({ error: err.message });
    }
  });

  app.get('/invoices', async (req, res) => {
    const id = req.query.id as string;
    const customer = req.query.customer as string;
    if (id) {
      try {
        const invoice = await client.getInvoice(id);
        res.json(invoice);
      } catch (err: any) {
        res.status(502).json({ error: err.message });
      }
      return;
    }
    if (!customer) return res.status(400).json({ error: 'customer é obrigatório' });
    try {
      const invoices = await client.listInvoices(customer);
      res.json(invoices);
    } catch (err: any) {
      res.status(502).json({ error: err.message });
    }
  });

  app.post('/subscriptions/cancel', async (req, res) => {
    res.status(501).json({ error: 'cancelamento ainda não implementado' });
  });

  app.post('/webhooks/asaas', async (req, res) => {
    if (!config.webhookToken || req.headers['asaas-access-token'] !== config.webhookToken) {
      return res.status(401).json({ error: 'não autorizado' });
    }
    try {
      // for now just acknowledge; downstream handling can be added as needed
      res.status(200).json({ received: true });
    } catch (err: any) {
      res.status(400).json({ error: err.message });
    }
  });

  const swaggerPath = path.resolve(process.cwd(), 'swagger');
  app.use('/swagger', express.static(swaggerPath));

  app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
    // eslint-disable-next-line no-console
    console.error(err);
    res.status(500).json({ error: 'erro interno do servidor' });
  });

  app.listen(config.port, () => {
    // eslint-disable-next-line no-console
    console.log(`server listening on :${config.port}`);
  });
}

bootstrap().catch((err) => {
  // eslint-disable-next-line no-console
  console.error('failed to start server', err);
  process.exit(1);
});
