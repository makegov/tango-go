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

func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL(srv.URL),
		WithRetries(0), // keep tests deterministic
		WithTimeout(2*time.Second),
	)
	return c, srv
}

func TestNewClientDefaults(t *testing.T) {
	t.Setenv("TANGO_API_KEY", "")
	t.Setenv("TANGO_BASE_URL", "")
	c := NewClient()
	if c.BaseURL() != DefaultBaseURL {
		t.Fatalf("expected default base URL %q, got %q", DefaultBaseURL, c.BaseURL())
	}
}

func TestNewClientEnvFallbacks(t *testing.T) {
	t.Setenv("TANGO_API_KEY", "env-key")
	t.Setenv("TANGO_BASE_URL", "https://custom.example/")
	c := NewClient()
	if c.cfg.apiKey != "env-key" {
		t.Errorf("expected env API key, got %q", c.cfg.apiKey)
	}
	if c.BaseURL() != "https://custom.example/" {
		t.Errorf("expected custom base URL from env, got %q", c.BaseURL())
	}
}

func TestSendsAuthHeader(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-KEY"); got != "test-key" {
			t.Errorf("expected X-API-KEY=test-key, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestListContractsAppliesFilters(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("awarding_agency") != "9700" {
			t.Errorf("expected awarding_agency=9700, got %q", q.Get("awarding_agency"))
		}
		if q.Get("ordering") != "-award_date" {
			t.Errorf("expected ordering=-award_date, got %q", q.Get("ordering"))
		}
		if q.Get("shape") != ShapeContractsMinimal {
			t.Errorf("expected shape=%q, got %q", ShapeContractsMinimal, q.Get("shape"))
		}
		if q.Get("limit") != "5" {
			t.Errorf("expected limit=5, got %q", q.Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"results":[{"piid":"X"}]}`))
	})

	page, err := c.ListContracts(context.Background(), &ListContractsOptions{
		ListOptions:    ListOptions{Limit: 5, Shape: ShapeContractsMinimal},
		AwardingAgency: "9700",
		Sort:           "award_date",
		Order:          "desc",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(page.Results) != 1 || page.Results[0]["piid"] != "X" {
		t.Errorf("unexpected results: %#v", page.Results)
	}
}

func Test401IsAuthError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"detail":"bad key"}`))
	})
	_, err := c.ListAgencies(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthError, got %T: %v", err, err)
	}
}

func Test404IsNotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := c.GetAgency(context.Background(), "1234")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func Test400IsValidationErrorWithDetail(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"detail":"bad shape: unknown_field"}`))
	})
	_, err := c.ListContracts(context.Background(), nil)
	var v *ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "bad shape: unknown_field") {
		t.Errorf("expected detail in error, got %q", err.Error())
	}
}

func Test429PopulatesRateLimitError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.Header().Set("X-RateLimit-Type", "burst")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	_, err := c.ListAgencies(context.Background(), nil)
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rle.RetryAfter != 7 {
		t.Errorf("expected RetryAfter=7, got %d", rle.RetryAfter)
	}
	if rle.LimitType != "burst" {
		t.Errorf("expected LimitType=burst, got %q", rle.LimitType)
	}
}

func Test429IsRetried(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"piid":"OK"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(1*time.Millisecond),
		WithTimeout(2*time.Second),
	)
	page, err := c.ListContracts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", calls)
	}
	if len(page.Results) != 1 || page.Results[0]["piid"] != "OK" {
		t.Errorf("unexpected results: %#v", page.Results)
	}
}

func TestRateLimitInfoSnapshot(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	})
	_, err := c.ListAgencies(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	info := c.RateLimitInfo()
	if info == nil {
		t.Fatal("expected RateLimitInfo to be populated")
	}
	if info.Remaining != 42 || info.Limit != 100 {
		t.Errorf("unexpected rate-limit info: %+v", info)
	}
}

func TestPaginatedResponseExtractsCursor(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"count": 3,
			"next": "https://example/api/idvs/?cursor=cD0xMjM=",
			"previous": null,
			"results": [{"key":"k1"},{"key":"k2"}]
		}`))
	})
	page, err := c.ListIDVs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if page.Cursor != "cD0xMjM=" {
		t.Errorf("expected cursor extracted from next URL, got %q", page.Cursor)
	}
	if page.Count != 3 || len(page.Results) != 2 {
		t.Errorf("unexpected pagination state: %+v", page)
	}
}

func TestIteratorWalksAllPages(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			w.Write([]byte(`{"count":4,"next":"http://x/?cursor=p2","results":[{"k":"a"},{"k":"b"}]}`))
		case 2:
			w.Write([]byte(`{"count":4,"next":"http://x/?cursor=p3","results":[{"k":"c"}]}`))
		default:
			w.Write([]byte(`{"count":4,"next":null,"results":[{"k":"d"}]}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if got := strings.Join(keys, ","); got != "a,b,c,d" {
		t.Errorf("expected a,b,c,d; got %s", got)
	}
}

// Regression: a multi-item single-page response must yield every item,
// not just the first. (F1 from verifier-findings.md — the original
// implementation set `done = true` and short-circuited Next() before
// the buffer was drained.)
func TestIteratorSinglePageMultiItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":3,"next":null,"results":[{"k":"a"},{"k":"b"},{"k":"c"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if got := strings.Join(keys, ","); got != "a,b,c" {
		t.Errorf("expected a,b,c; got %q (length %d)", got, len(keys))
	}
}

// Regression: a multi-item LAST page must yield every item.
// Two-page response: 2 on page 1, 3 on the final page = 5 total.
func TestIteratorMultiItemLastPage(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			w.Write([]byte(`{"count":5,"next":"http://x/?cursor=p2","results":[{"k":"a"},{"k":"b"}]}`))
		default:
			w.Write([]byte(`{"count":5,"next":null,"results":[{"k":"c"},{"k":"d"},{"k":"e"}]}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if got := strings.Join(keys, ","); got != "a,b,c,d,e" {
		t.Errorf("expected a,b,c,d,e; got %q (length %d)", got, len(keys))
	}
}

func TestIteratorSeqRangeOverFunc(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			w.Write([]byte(`{"count":3,"next":"http://x/?cursor=p2","results":[{"k":"a"},{"k":"b"}]}`))
		default:
			w.Write([]byte(`{"count":3,"next":null,"results":[{"k":"c"}]}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	var keys []string
	for rec, err := range c.IterateContracts(context.Background(), nil).Seq() {
		if err != nil {
			t.Fatalf("Seq error: %v", err)
		}
		keys = append(keys, rec["k"].(string))
	}
	if got := strings.Join(keys, ","); got != "a,b,c" {
		t.Errorf("expected a,b,c; got %s", got)
	}
}

func TestIteratorSeqEarlyBreakStopsFetching(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":99,"next":"http://x/?cursor=more","results":[{"k":"a"},{"k":"b"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	for rec, err := range c.IterateContracts(context.Background(), nil).Seq() {
		if err != nil {
			t.Fatalf("Seq error: %v", err)
		}
		_ = rec
		break // bail after the first item
	}
	if calls != 1 {
		t.Errorf("expected only 1 fetch (break should stop iteration); got %d", calls)
	}
}

func TestTimeoutBecomesTimeoutError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0), WithTimeout(20*time.Millisecond))
	_, err := c.ListAgencies(context.Background(), nil)
	var te *TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("expected *TimeoutError, got %T: %v", err, err)
	}
}

func TestResolveValidatesInput(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.Resolve(context.Background(), ResolveInput{TargetType: ResolveEntity})
	var v *ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("expected *ValidationError for missing Name, got %T: %v", err, err)
	}
	_, err = c.Resolve(context.Background(), ResolveInput{Name: "Acme", TargetType: "bogus"})
	if !errors.As(err, &v) {
		t.Fatalf("expected *ValidationError for bad TargetType, got %T: %v", err, err)
	}
}

func TestResolvePostsBody(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"candidates":[{"identifier":"UEI123","display_name":"Acme"}]}`))
	})
	res, err := c.Resolve(context.Background(), ResolveInput{Name: "Acme", TargetType: ResolveEntity})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res.Candidates) != 1 || res.Candidates[0].Identifier != "UEI123" {
		t.Errorf("unexpected resolve response: %+v", res)
	}
}
