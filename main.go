package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"asaas/src/payments"

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

	handler := buildHandler(service, client)

	srv := &http.Server{ //nolint:gosec
		Addr:         ":" + cfg.Port,
		Handler:      handler,
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
		return AppConfig{}, fmt.Errorf("DATABASE_URL n\u00e3o est\u00e1 definida")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return AppConfig{Port: port, DatabaseDSN: dsn, Asaas: asaasConfig}, nil
}

func buildHandler(service *payments.Service, client *payments.AsaasClient) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, service, client)
	return withRecovery(withRequestLogging(mux))
}

func registerRoutes(mux *http.ServeMux, service *payments.Service, client *payments.AsaasClient) {
	customerHandler := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			var payload payments.CustomerRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				respondError(w, http.StatusBadRequest, "payload inv\u00e1lido")
				return
			}
			_, remote, err := service.RegisterCustomer(req.Context(), payload)
			if err != nil {
				respondError(w, statusForError(err), err.Error())
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		case http.MethodGet:
			id := req.URL.Query().Get("id")
			if id == "" {
				respondError(w, http.StatusBadRequest, "id \u00e9 obrigat\u00f3rio")
				return
			}
			customer, err := client.GetCustomer(req.Context(), id)
			if err != nil {
				respondError(w, http.StatusBadGateway, err.Error())
				return
			}
			respondJSON(w, customer, http.StatusOK)
		default:
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
		}
	}

	paymentHandler := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			var payload payments.PaymentRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				respondError(w, http.StatusBadRequest, "payload inv\u00e1lido")
				return
			}
			_, remote, err := service.CreatePayment(req.Context(), payload)
			if err != nil {
				respondError(w, statusForError(err), err.Error())
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		case http.MethodGet:
			id := req.URL.Query().Get("id")
			if id == "" {
				respondError(w, http.StatusBadRequest, "id \u00e9 obrigat\u00f3rio")
				return
			}
			payment, err := client.GetPayment(req.Context(), id)
			if err != nil {
				respondError(w, http.StatusBadGateway, err.Error())
				return
			}
			respondJSON(w, payment, http.StatusOK)
		default:
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
		}
	}

	subscriptionHandler := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			var payload payments.SubscriptionRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				respondError(w, http.StatusBadRequest, "payload inv\u00e1lido")
				return
			}
			_, remote, err := service.CreateSubscription(req.Context(), payload)
			if err != nil {
				respondError(w, statusForError(err), err.Error())
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		default:
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
		}
	}

	subscriptionCancelHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
			return
		}
		id := req.URL.Query().Get("id")
		if id == "" {
			respondError(w, http.StatusBadRequest, "id \u00e9 obrigat\u00f3rio")
			return
		}
		subscription, err := client.CancelSubscription(req.Context(), id)
		if err != nil {
			respondError(w, http.StatusBadGateway, err.Error())
			return
		}
		respondJSON(w, subscription, http.StatusOK)
	}

	invoiceHandler := func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			var payload payments.InvoiceRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				respondError(w, http.StatusBadRequest, "payload inv\u00e1lido")
				return
			}
			_, remote, err := service.CreateInvoice(req.Context(), payload)
			if err != nil {
				respondError(w, statusForError(err), err.Error())
				return
			}
			respondJSON(w, remote, http.StatusCreated)
		case http.MethodGet:
			id := req.URL.Query().Get("id")
			if id == "" {
				respondError(w, http.StatusBadRequest, "id \u00e9 obrigat\u00f3rio")
				return
			}
			invoice, err := client.GetInvoice(req.Context(), id)
			if err != nil {
				respondError(w, http.StatusBadGateway, err.Error())
				return
			}
			respondJSON(w, invoice, http.StatusOK)
		default:
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
		}
	}

	webhookHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "m\u00e9todo n\u00e3o permitido")
			return
		}

		expectedToken := os.Getenv("ASAAS_WEBHOOK_TOKEN")
		if expectedToken == "" || req.Header.Get("asaas-access-token") != expectedToken {
			respondError(w, http.StatusUnauthorized, "n\u00e3o autorizado")
			return
		}

		payload, err := io.ReadAll(req.Body)
		if err != nil {
			respondError(w, http.StatusBadRequest, "n\u00e3o foi poss\u00edvel ler o corpo")
			return
		}
		defer req.Body.Close()

		if err := service.HandleWebhookPayload(req.Context(), payload); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("/customers", customerHandler)
	mux.HandleFunc("/customers/", customerHandler)
	mux.HandleFunc("/payments", paymentHandler)
	mux.HandleFunc("/payments/", paymentHandler)
	mux.HandleFunc("/subscriptions", subscriptionHandler)
	mux.HandleFunc("/subscriptions/", subscriptionHandler)
	mux.HandleFunc("/subscriptions/cancel", subscriptionCancelHandler)
	mux.HandleFunc("/subscriptions/cancel/", subscriptionCancelHandler)
	mux.HandleFunc("/invoices", invoiceHandler)
	mux.HandleFunc("/invoices/", invoiceHandler)
	mux.HandleFunc("/webhooks/asaas", webhookHandler)

	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s completed in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				respondError(w, http.StatusInternalServerError, "erro interno do servidor")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, payload any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		respondError(w, http.StatusInternalServerError, "erro interno do servidor")
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func statusForError(err error) int {
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound
	}
	return http.StatusBadGateway
}
