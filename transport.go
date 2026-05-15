package tango

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxBackoff       = 10 * time.Second
	apiKeyHeader     = "X-API-KEY"
	rateLimitRemHdr  = "X-RateLimit-Remaining"
	rateLimitLimHdr  = "X-RateLimit-Limit"
	rateLimitRsetHdr = "X-RateLimit-Reset"
	rateLimitTypeHdr = "X-RateLimit-Type"
	retryAfterHdr    = "Retry-After"
)

// RateLimitInfo is a snapshot of the rate-limit headers from the most
// recent response. Mirrors Python's RateLimitInfo dataclass.
type RateLimitInfo struct {
	Remaining  int    // -1 when header was absent or unparseable
	Limit      int    // -1 when absent
	ResetIn    int    // seconds; -1 when absent
	RetryAfter int    // seconds; -1 when absent
	LimitType  string // empty when absent
}

// rateLimitState is the mutable snapshot kept on the client.
type rateLimitState struct {
	mu      sync.RWMutex
	info    *RateLimitInfo
	headers http.Header
}

func (s *rateLimitState) set(info *RateLimitInfo, headers http.Header) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.info = info
	s.headers = headers
}

func (s *rateLimitState) get() (*RateLimitInfo, http.Header) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.info, s.headers
}

func parseIntHeader(h http.Header, key string) int {
	v := h.Get(key)
	if v == "" {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return -1
	}
	return n
}

func parseRateLimit(h http.Header) *RateLimitInfo {
	return &RateLimitInfo{
		Remaining:  parseIntHeader(h, rateLimitRemHdr),
		Limit:      parseIntHeader(h, rateLimitLimHdr),
		ResetIn:    parseIntHeader(h, rateLimitRsetHdr),
		RetryAfter: parseIntHeader(h, retryAfterHdr),
		LimitType:  h.Get(rateLimitTypeHdr),
	}
}

// parseRetryAfter returns the Retry-After delay as a duration. Accepts
// both delta-seconds and HTTP-date forms. Returns 0 on parse failure.
func parseRetryAfter(h http.Header) time.Duration {
	raw := h.Get(retryAfterHdr)
	if raw == "" {
		return 0
	}
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && n >= 0 {
		d := time.Duration(n) * time.Second
		if d > maxBackoff {
			d = maxBackoff
		}
		return d
	}
	if t, err := http.ParseTime(raw); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		if d > maxBackoff {
			d = maxBackoff
		}
		return d
	}
	return 0
}

// requestSpec is the input to (*Client).do.
type requestSpec struct {
	method string
	path   string
	query  url.Values
	body   any
}

// do executes a single HTTP request with retries. The returned []byte is
// the raw response body; callers decode it themselves so we don't need
// generic type machinery in transport.
func (c *Client) do(ctx context.Context, spec requestSpec) ([]byte, error) {
	maxAttempts := c.cfg.retries + 1
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		body, err := c.attempt(ctx, spec)
		if err == nil {
			return body, nil
		}

		lastErr = err
		if !IsRetryable(err) {
			return nil, err
		}

		// Pick wait time: Retry-After hint > exponential backoff.
		var wait time.Duration
		if rle, ok := err.(*RateLimitError); ok && rle.RetryAfter > 0 {
			wait = time.Duration(rle.RetryAfter) * time.Second
			if wait > maxBackoff {
				wait = maxBackoff
			}
		} else {
			exp := c.cfg.retryBackoff * (1 << attempt) // doubles each retry
			if exp > maxBackoff {
				exp = maxBackoff
			}
			wait = exp
		}

		if attempt+1 >= maxAttempts {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}

	return nil, lastErr
}

func (c *Client) attempt(ctx context.Context, spec requestSpec) ([]byte, error) {
	full, err := c.buildURL(spec.path, spec.query)
	if err != nil {
		return nil, &APIError{Message: "build url", Cause: err}
	}

	var bodyReader io.Reader
	if spec.body != nil {
		jsonBody, err := json.Marshal(spec.body)
		if err != nil {
			return nil, &APIError{Message: "encode body", Cause: err}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	reqCtx := ctx
	var cancel context.CancelFunc
	if c.cfg.timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, c.cfg.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, spec.method, full, bodyReader)
	if err != nil {
		return nil, &APIError{Message: "build request", Cause: err}
	}

	req.Header.Set("Accept", "application/json")
	if spec.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.apiKey != "" {
		req.Header.Set(apiKeyHeader, c.cfg.apiKey)
	}
	if c.cfg.userAgent != "" {
		req.Header.Set("User-Agent", c.cfg.userAgent)
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		// Timeout vs other transport failure.
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return nil, &TimeoutError{&APIError{
				Message: fmt.Sprintf("request timed out after %s", c.cfg.timeout),
			}}
		}
		if errors.Is(err, context.Canceled) {
			return nil, ctx.Err()
		}
		return nil, &APIError{Message: "request failed", Cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: "read body", Cause: readErr}
	}

	// Snapshot rate-limit / response headers for observability.
	c.rateLimit.set(parseRateLimit(resp.Header), resp.Header.Clone())

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Some endpoints (DELETE) return empty bodies — that's fine.
		return raw, nil
	}

	// Try to decode an error body. Server may return any JSON-y shape.
	var decoded any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &decoded)
	}

	switch resp.StatusCode {
	case 401:
		return nil, &AuthError{&APIError{
			StatusCode:   401,
			Message:      "invalid API key or authentication required",
			ResponseData: decoded,
		}}
	case 404:
		return nil, &NotFoundError{&APIError{
			StatusCode:   404,
			Message:      "resource not found",
			ResponseData: decoded,
		}}
	case 400:
		return nil, &ValidationError{&APIError{
			StatusCode:   400,
			Message:      extractValidationMessage(decoded),
			ResponseData: decoded,
		}}
	case 429:
		retryAfter := parseRetryAfter(resp.Header)
		return nil, &RateLimitError{
			APIError: &APIError{
				StatusCode:   429,
				Message:      "rate limit exceeded",
				ResponseData: decoded,
			},
			RetryAfter: int(retryAfter.Seconds()),
			LimitType:  resp.Header.Get(rateLimitTypeHdr),
		}
	default:
		return nil, &APIError{
			StatusCode:   resp.StatusCode,
			Message:      fmt.Sprintf("API request failed with status %d", resp.StatusCode),
			ResponseData: decoded,
		}
	}
}

// buildURL resolves a path against baseURL and appends query params.
func (c *Client) buildURL(path string, q url.Values) (string, error) {
	base := strings.TrimRight(c.cfg.baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(base + path)
	if err != nil {
		return "", err
	}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

// extractValidationMessage pulls a useful error string out of a 400 body.
// The Tango API surfaces validation errors in a handful of shapes:
//
//   - {"detail": "..."} or {"message": "..."} or {"error": "..."}
//     (DRF / generic envelope)
//   - {"<field>": ["..."]} (DRF field-error array, one or more keys)
//   - {"<field>": "..."} (single-string field error)
//
// Iteration over field-keyed shapes uses sorted keys so the surfaced
// message is deterministic for any given body — Go's map iteration is
// randomized, so naive walks would yield different errors across runs.
func extractValidationMessage(data any) string {
	const fallback = "invalid request parameters"
	obj, ok := data.(map[string]any)
	if !ok {
		return fallback
	}
	// Envelope keys first, in explicit priority order.
	for _, key := range []string{"detail", "message", "error"} {
		if s, ok := obj[key].(string); ok && s != "" {
			return "invalid request parameters: " + s
		}
	}
	// Field-error shape — walk keys in sorted order for determinism.
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch t := obj[k].(type) {
		case []any:
			if len(t) > 0 {
				if s, ok := t[0].(string); ok && s != "" {
					return "invalid request parameters: " + s
				}
			}
		case string:
			if t != "" {
				return "invalid request parameters: " + t
			}
		}
	}
	return fallback
}
