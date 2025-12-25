package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AsaasClient handles authenticated requests against the Asaas API.
type AsaasClient struct {
	baseURL    string
	token      string
	httpClient HTTPClient
}

// NewAsaasClient constructs an AsaasClient using the provided configuration.
func NewAsaasClient(cfg Config, client HTTPClient) *AsaasClient {
	httpClient := client
	if httpClient == nil {
		httpClient = NewDefaultHTTPClient()
	}

	return &AsaasClient{
		baseURL:    cfg.APIBaseURL,
		token:      cfg.APIToken,
		httpClient: httpClient,
	}
}

// doRequest executes the HTTP request, decoding the response into the provided value.
func (c *AsaasClient) doRequest(ctx context.Context, method, path string, payload any, v any) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.baseURL, path), body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		data, _ := io.ReadAll(res.Body)
		return fmt.Errorf("asaas api error: status=%d body=%s", res.StatusCode, string(data))
	}

	if v == nil {
		return nil
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
