package payments

import (
    "fmt"
    "os"
)

// Config holds API credentials and base URL for Asaas.
type Config struct {
    APIBaseURL string
    APIToken   string
}

const (
    envAPIURL   = "ASAAS_API_URL"
    envAPIToken = "ASAAS_API_TOKEN"
)

// LoadConfig reads configuration from environment variables.
func LoadConfig() (*Config, error) {
    baseURL := os.Getenv(envAPIURL)
    token := os.Getenv(envAPIToken)

    if baseURL == "" {
        return nil, fmt.Errorf("environment variable %s is required", envAPIURL)
    }

    if token == "" {
        return nil, fmt.Errorf("environment variable %s is required", envAPIToken)
    }

    return &Config{
        APIBaseURL: baseURL,
        APIToken:   token,
    }, nil
}
