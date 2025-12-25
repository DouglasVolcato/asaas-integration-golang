package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client provides access to Asaas HTTP APIs and a Postgres database.
// All operations require an active context for cancellation and deadlines.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
	db         Database
}

// Option configures a Client instance.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithBaseURL overrides the API base URL. Useful for tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// WithAPIToken overrides the API token. Useful for tests.
func WithAPIToken(token string) Option {
	return func(c *Client) {
		if token != "" {
			c.apiToken = token
		}
	}
}

// NewClient builds a Client using environment variables as defaults and applies optional overrides.
// Expected environment variables: ASAAS_API_URL and ASAAS_API_TOKEN.
func NewClient(ctx context.Context, db *sql.DB, opts ...Option) (*Client, error) {
	return NewClientWithDatabase(ctx, NewSQLDatabase(db), opts...)
}

// NewClientWithDatabase allows injecting custom database implementations, useful for tests.
func NewClientWithDatabase(ctx context.Context, database Database, opts ...Option) (*Client, error) {
	baseURL := os.Getenv("ASAAS_API_URL")
	token := os.Getenv("ASAAS_API_TOKEN")

	if baseURL == "" || token == "" {
		return nil, errors.New("ASAAS_API_URL and ASAAS_API_TOKEN must be set")
	}

	client := &Client{
		baseURL:    baseURL,
		apiToken:   token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		db:         database,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.db == nil {
		return nil, errors.New("database connection is required")
	}

	if err := client.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	return client, nil
}

// baseRequest returns a prepared HTTP request with authentication and JSON headers applied.
func (c *Client) baseRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// do executes an HTTP request and returns the response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.httpClient == nil {
		return nil, errors.New("http client is not configured")
	}
	return c.httpClient.Do(req)
}
