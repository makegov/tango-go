package tango

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRetriesOnGenericRetryable5xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(1*time.Millisecond),
		WithTimeout(2*time.Second),
	)
	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", calls)
	}
}

func TestRetries408IsRetryable(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(408)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(1*time.Millisecond),
		WithTimeout(2*time.Second),
	)
	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected success after 408 retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestNetworkErrorIsRetryable(t *testing.T) {
	// Spin up a server so we get a real URL, then close it immediately
	// so the first request fails at the network layer. On the second
	// attempt we bring a fresh server up at a different URL — we do this
	// with a custom approach: record attempts and succeed on 2nd.
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(srv.Close)

	// Close the server to simulate a transport-level error on first attempt,
	// then use a custom RoundTripper that re-opens on retry.
	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(1*time.Millisecond),
		WithTimeout(2*time.Second),
		WithHTTPClient(&http.Client{
			Transport: &countingTransport{
				inner:   http.DefaultTransport,
				failOn:  1,
				realURL: srv.URL,
			},
		}),
	)
	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected success after network error retry, got: %v", err)
	}
}

// countingTransport fails on attempt failOn (1-based), succeeds otherwise.
type countingTransport struct {
	inner   http.RoundTripper
	failOn  int
	attempt int
	realURL string
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.attempt++
	if t.attempt == t.failOn {
		return nil, errors.New("simulated network error")
	}
	return t.inner.RoundTrip(req)
}

func TestContextCancellationStopsRetry(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(503)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(5),
		WithRetryBackoff(200*time.Millisecond), // backoff > ctx timeout
		WithTimeout(1*time.Second),
	)
	_, err := c.ListAgencies(ctx, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestValidationErrorExtractsDetailFromVariousShapes(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{
			name:    "detail key",
			body:    `{"detail":"bad shape value"}`,
			wantMsg: "bad shape value",
		},
		{
			name:    "message key",
			body:    `{"message":"field is required"}`,
			wantMsg: "field is required",
		},
		{
			name:    "error key",
			body:    `{"error":"unknown filter"}`,
			wantMsg: "unknown filter",
		},
		{
			// Only array, no other string fields to avoid map-iteration non-determinism
			name:    "field array only",
			body:    `{"award_date":["enter a valid date"]}`,
			wantMsg: "enter a valid date",
		},
		{
			// Only string field — no other competing entries
			name:    "field string",
			body:    `{"award_date":"this field is required"}`,
			wantMsg: "this field is required",
		},
		{
			name:    "fallback",
			body:    `{}`,
			wantMsg: "invalid request parameters",
		},
		{
			// Multi-key body: sorted-key iteration picks "award_date"
			// (alphabetically first) over "zzz", deterministically.
			name:    "multi-key deterministic",
			body:    `{"zzz":"later","award_date":["enter a valid date"]}`,
			wantMsg: "enter a valid date",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
				w.Write([]byte(body))
			})
			_, err := c.ListContracts(context.Background(), nil)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *ValidationError, got %T: %v", err, err)
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("expected error to contain %q, got %q", tc.wantMsg, err.Error())
			}
		})
	}
}

func Test5xxReturnsAPIErrorAfterExhausted(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(1*time.Millisecond),
		WithTimeout(2*time.Second),
	)
	_, err := c.ListAgencies(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected StatusCode 500, got %d", apiErr.StatusCode)
	}
	if calls != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestIsRetryableStatuses(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{&APIError{StatusCode: 0}, true}, // network error
		{&APIError{StatusCode: 408}, true},
		{&APIError{StatusCode: 429}, true},
		{&APIError{StatusCode: 500}, true},
		{&APIError{StatusCode: 503}, true},
		{&APIError{StatusCode: 400}, false},
		{&APIError{StatusCode: 401}, false},
		{&APIError{StatusCode: 404}, false},
		{&AuthError{&APIError{StatusCode: 401}}, false},
		{&ValidationError{&APIError{StatusCode: 400}}, false},
		{&NotFoundError{&APIError{StatusCode: 404}}, false},
		{&RateLimitError{APIError: &APIError{StatusCode: 429}}, true},
		{&TimeoutError{&APIError{StatusCode: 0}}, true},
	}
	for _, tc := range cases {
		got := IsRetryable(tc.err)
		if got != tc.want {
			t.Errorf("IsRetryable(%T status=%d): want %v, got %v",
				tc.err, func() int {
					if a, ok := tc.err.(interface{ Unwrap() error }); ok {
						_ = a
					}
					return 0
				}(), tc.want, got)
		}
	}
}

func TestLastResponseHeadersPopulated(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "hello")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	})
	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	h := c.LastResponseHeaders()
	if h == nil {
		t.Fatal("expected LastResponseHeaders to be populated")
	}
	if h.Get("X-Custom-Header") != "hello" {
		t.Errorf("expected X-Custom-Header=hello, got %q", h.Get("X-Custom-Header"))
	}
}

func TestParseRetryAfterDeltaSeconds(t *testing.T) {
	// 429 with Retry-After: 0 → RateLimitError with RetryAfter=0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(429)
	})
	_, err := c.ListAgencies(context.Background(), nil)
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rle.RetryAfter != 3 {
		t.Errorf("expected RetryAfter=3, got %d", rle.RetryAfter)
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	// A past HTTP-date should return 0 (clamp to 0).
	d := parseRetryAfter(http.Header{
		"Retry-After": []string{"Thu, 01 Jan 2015 00:00:00 GMT"},
	})
	if d != 0 {
		t.Errorf("expected 0 for past HTTP-date, got %v", d)
	}
}

func TestParseRetryAfterCapAtMaxBackoff(t *testing.T) {
	// Very large delta-seconds should be capped at maxBackoff (10s).
	d := parseRetryAfter(http.Header{
		"Retry-After": []string{"999"},
	})
	if d != maxBackoff {
		t.Errorf("expected %v (maxBackoff), got %v", maxBackoff, d)
	}
}

func TestBuildURLAppendsQuery(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("https://api.example.com"))
	got, err := c.buildURL("/api/agencies/", map[string][]string{"page": {"2"}})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(got, "page=2") {
		t.Errorf("expected query param page=2 in %q", got)
	}
}

func TestParseIntHeaderMissing(t *testing.T) {
	h := http.Header{}
	got := parseIntHeader(h, "X-RateLimit-Remaining")
	if got != -1 {
		t.Errorf("expected -1 for missing header, got %d", got)
	}
}

func TestParseIntHeaderInvalid(t *testing.T) {
	h := http.Header{"X-RateLimit-Remaining": []string{"not-a-number"}}
	got := parseIntHeader(h, "X-RateLimit-Remaining")
	if got != -1 {
		t.Errorf("expected -1 for non-numeric header, got %d", got)
	}
}
