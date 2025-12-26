package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"asaas/src/payments"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type AppConfig struct {
	Port        string
	DatabaseDSN string
	Asaas       payments.Config
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	repo := payments.NewPostgresRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	client := payments.NewAsaasClient(cfg.Asaas)
	service := payments.NewService(repo, client)
	webhookHandler := payments.NewWebhookHandler(service)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	registerRoutes(router, service, client, webhookHandler)

	srv := &http.Server{ //nolint:gosec
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func loadConfig() (AppConfig, error) {
	asaasConfig, err := payments.LoadConfigFromEnv()
	if err != nil {
		return AppConfig{}, err
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return AppConfig{}, fmt.Errorf("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return AppConfig{Port: port, DatabaseDSN: dsn, Asaas: asaasConfig}, nil
}

func registerRoutes(r *chi.Mux, service *payments.Service, client *payments.AsaasClient, webhook http.Handler) {
	r.Route("/customers", func(cr chi.Router) {
		cr.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var payload payments.CustomerRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			_, remote, err := service.RegisterCustomer(req.Context(), payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		})

		cr.Get("/", func(w http.ResponseWriter, req *http.Request) {
			externalRef := req.URL.Query().Get("externalReference")
			if externalRef == "" {
				http.Error(w, "externalReference is required", http.StatusBadRequest)
				return
			}
			customer, err := client.GetCustomer(req.Context(), externalRef)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, customer, http.StatusOK)
		})

		cr.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			customerID := chi.URLParam(req, "id")
			customer, err := client.GetCustomer(req.Context(), customerID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, customer, http.StatusOK)
		})
	})

	r.Route("/payments", func(pr chi.Router) {
		pr.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var payload payments.PaymentRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			_, remote, err := service.CreatePayment(req.Context(), payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		})

		pr.Get("/", func(w http.ResponseWriter, req *http.Request) {
			externalRef := req.URL.Query().Get("externalReference")
			if externalRef == "" {
				http.Error(w, "externalReference is required", http.StatusBadRequest)
				return
			}
			payment, err := client.GetPayment(req.Context(), externalRef)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, payment, http.StatusOK)
		})

		pr.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			paymentID := chi.URLParam(req, "id")
			payment, err := client.GetPayment(req.Context(), paymentID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, payment, http.StatusOK)
		})
	})

	r.Route("/subscriptions", func(sr chi.Router) {
		sr.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var payload payments.SubscriptionRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			_, remote, err := service.CreateSubscription(req.Context(), payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		})

		sr.Post("/cancel", func(w http.ResponseWriter, req *http.Request) {
			externalRef := req.URL.Query().Get("externalReference")
			if externalRef == "" {
				http.Error(w, "externalReference is required", http.StatusBadRequest)
				return
			}
			subscription, err := client.CancelSubscription(req.Context(), externalRef)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, subscription, http.StatusOK)
		})
	})

	r.Route("/invoices", func(ir chi.Router) {
		ir.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var payload payments.InvoiceRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			_, remote, err := service.CreateInvoice(req.Context(), payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		})

		ir.Get("/", func(w http.ResponseWriter, req *http.Request) {
			externalRef := req.URL.Query().Get("externalReference")
			if externalRef == "" {
				http.Error(w, "externalReference is required", http.StatusBadRequest)
				return
			}
			invoice, err := client.GetInvoice(req.Context(), externalRef)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, invoice, http.StatusOK)
		})

		ir.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			invoiceID := chi.URLParam(req, "id")
			invoice, err := client.GetInvoice(req.Context(), invoiceID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			respondJSON(w, invoice, http.StatusOK)
		})
	})

	r.Post("/webhooks/asaas", webhook.ServeHTTP)

	r.Handle("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
}

func respondJSON(w http.ResponseWriter, payload any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
