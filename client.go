// Package tango is the official Go SDK for the Tango API.
//
// The Tango API provides programmatic access to federal-contracting data —
// contracts, IDVs, entities, opportunities, grants, vehicles, and more —
// with dynamic response shaping so callers fetch only the fields they need.
//
// Quick start:
//
//	client := tango.NewClient(tango.WithAPIKey("…"))
//
//	page, err := client.ListContracts(ctx, &tango.ListContractsOptions{
//	    AwardingAgency: "9700",
//	    Shape:          tango.ShapeContractsMinimal,
//	    Limit:          25,
//	})
//
// Documentation: https://docs.makegov.com
package tango

import (
	"net/http"
	"os"
	"time"
)

// Client is the Tango API client. Construct one with NewClient and pass
// it around — Client is safe for concurrent use.
type Client struct {
	cfg       clientConfig
	rateLimit *rateLimitState
}

// NewClient constructs a Client. The API key is taken from the WithAPIKey
// option, falling back to the TANGO_API_KEY environment variable. The
// base URL is taken from WithBaseURL, falling back to TANGO_BASE_URL,
// then DefaultBaseURL.
//
// Calling NewClient with no options is valid and useful in environments
// where TANGO_API_KEY is set (CI, container, dev shell).
func NewClient(opts ...Option) *Client {
	cfg := clientConfig{
		apiKey:       os.Getenv("TANGO_API_KEY"),
		baseURL:      firstNonEmpty(os.Getenv("TANGO_BASE_URL"), DefaultBaseURL),
		timeout:      30 * time.Second,
		retries:      3,
		retryBackoff: 250 * time.Millisecond,
		userAgent:    "tango-go/" + Version,
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{}
	}
	return &Client{
		cfg:       cfg,
		rateLimit: &rateLimitState{},
	}
}

// BaseURL returns the resolved base URL the client will hit.
func (c *Client) BaseURL() string { return c.cfg.baseURL }

// RateLimitInfo returns a snapshot of the rate-limit headers from the
// most recent response. Returns nil if no request has completed yet.
func (c *Client) RateLimitInfo() *RateLimitInfo {
	info, _ := c.rateLimit.get()
	return info
}

// LastResponseHeaders returns the response headers from the most recent
// completed request (useful for X-Request-Id, X-Tango-Trace-Id, etc.).
// Returns nil if no request has completed.
func (c *Client) LastResponseHeaders() http.Header {
	_, h := c.rateLimit.get()
	return h
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
