package main

import (
	"context"
	"database/sql"
	"log"

	"asaas/src/payments"
)

func main() {
	ctx := context.Background()
	cfg, err := payments.LoadConfigFromEnv()
	if err != nil {
		log.Printf("configuration not loaded: %v", err)
		return
	}

	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatalf("failed to create db connection: %v", err)
	}
	defer db.Close()

	repo := payments.NewPostgresRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	client := payments.NewAsaasClient(cfg)
	_ = payments.NewService(repo, client)
	log.Println("payments module initialized")
}
