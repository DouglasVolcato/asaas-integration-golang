package payments

import (
	"fmt"
	"os"
)

// Config holds environment driven configuration for Asaas API access.
type Config struct {
	APIBaseURL string
	APIToken   string
}

// LoadConfig reads configuration from environment variables.
//
// Expected variables:
//   - ASAAS_API_URL:   base URL of the Asaas API (for example, https://api.asaas.com/v3)
//   - ASAAS_API_TOKEN: secret token used to authorize requests
func LoadConfig() (Config, error) {
	baseURL := os.Getenv("ASAAS_API_URL")
	token := os.Getenv("ASAAS_API_TOKEN")

	if baseURL == "" {
		return Config{}, fmt.Errorf("missing ASAAS_API_URL environment variable")
	}

	if token == "" {
		return Config{}, fmt.Errorf("missing ASAAS_API_TOKEN environment variable")
	}

	return Config{APIBaseURL: baseURL, APIToken: token}, nil
}
