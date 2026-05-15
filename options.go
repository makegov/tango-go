package tango

import (
	"net/http"
	"time"
)

// Option configures a Client. Pass any number of these to NewClient.
type Option func(*clientConfig)

type clientConfig struct {
	apiKey        string
	baseURL       string
	timeout       time.Duration
	retries       int
	retryBackoff  time.Duration
	httpClient    *http.Client
	userAgent     string
}

// WithAPIKey sets the Tango API key. When omitted, the client reads the
// TANGO_API_KEY environment variable. Passing an explicit empty string
// is equivalent to not passing this option.
func WithAPIKey(key string) Option {
	return func(c *clientConfig) {
		if key != "" {
			c.apiKey = key
		}
	}
}

// WithBaseURL overrides the API base URL. Precedence: this option >
// TANGO_BASE_URL env var > DefaultBaseURL.
func WithBaseURL(url string) Option {
	return func(c *clientConfig) {
		if url != "" {
			c.baseURL = url
		}
	}
}

// WithTimeout sets the per-request timeout. Defaults to 30s.
// A timeout of 0 disables the deadline (matching net/http semantics).
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

// WithRetries sets the number of retry attempts on retryable failures
// (5xx, 408, 429, transport errors). The initial attempt is not counted
// as a retry, so the total attempt count is retries + 1. Defaults to 3.
func WithRetries(n int) Option {
	return func(c *clientConfig) {
		if n < 0 {
			n = 0
		}
		c.retries = n
	}
}

// WithRetryBackoff sets the initial backoff for retries. The actual wait
// doubles each retry, capped at 10s. The server's Retry-After header
// (when present on 429/503) overrides this. Defaults to 250ms.
func WithRetryBackoff(d time.Duration) Option {
	return func(c *clientConfig) {
		if d < 0 {
			d = 0
		}
		c.retryBackoff = d
	}
}

// WithHTTPClient supplies a custom *http.Client. Useful for injecting
// custom transports (proxies, tracing, etc.). The client's Timeout field
// is ignored — request deadlines are managed via context.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithUserAgent sets a custom User-Agent header. Defaults to
// "tango-go/<version>".
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}
