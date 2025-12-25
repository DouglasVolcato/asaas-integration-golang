package payments

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Client handles HTTP communication with the Asaas API.
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

// NewClient builds a new Client from the provided configuration.
func NewClient(cfg *Config) *Client {
    return &Client{
        baseURL: cfg.APIBaseURL,
        token:   cfg.APIToken,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// doRequest executes an HTTP request with JSON encoding/decoding.
func (c *Client) doRequest(ctx context.Context, method, path string, payload any, result any) error {
    var body io.Reader
    if payload != nil {
        buf, err := json.Marshal(payload)
        if err != nil {
            return fmt.Errorf("marshal payload: %w", err)
        }
        body = bytes.NewBuffer(buf)
    }

    req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.baseURL, path), body)
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("access_token", c.token)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        data, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("asaas returned status %d: %s", resp.StatusCode, string(data))
    }

    if result == nil {
        return nil
    }

    decoder := json.NewDecoder(resp.Body)
    if err := decoder.Decode(result); err != nil {
        return fmt.Errorf("decode response: %w", err)
    }

    return nil
}
