module asaas

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/lib/pq v1.10.9
)

require github.com/joho/godotenv v1.5.1

replace github.com/go-chi/chi/v5 => ./internal/chi

replace github.com/go-chi/chi/v5/middleware => ./internal/chi/middleware
