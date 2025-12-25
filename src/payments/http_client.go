package payments

import (
	"net"
	"net/http"
	"time"
)

// HTTPClient abstracts the Do method used by the Asaas client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewDefaultHTTPClient returns an HTTP client with sensible defaults for external calls.
func NewDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}
