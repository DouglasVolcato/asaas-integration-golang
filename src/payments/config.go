package payments

import (
	"fmt"
	"os"
)

// Config holds credentials and endpoints for the Asaas API.
type Config struct {
	APIURL   string
	APIToken string
}

// LoadConfigFromEnv builds a Config using environment variables.
func LoadConfigFromEnv() (Config, error) {
	apiURL := os.Getenv("ASAAS_API_URL")
	token := os.Getenv("ASAAS_API_TOKEN")
	if apiURL == "" {
		return Config{}, fmt.Errorf("ASAAS_API_URL is not set")
	}
	if token == "" {
		return Config{}, fmt.Errorf("ASAAS_API_TOKEN is not set")
	}
	return Config{APIURL: apiURL, APIToken: token}, nil
}
