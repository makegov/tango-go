# Client Configuration

`*tango.Client` is the entry point for every API call. This guide covers the constructor, options, environment variables, retry semantics, error handling, rate-limit observability, and the iterator surface.

For per-method signatures, see [`API_REFERENCE.md`](API_REFERENCE.md). For webhook signing and the CRUD endpoint methods, see [`WEBHOOKS.md`](WEBHOOKS.md). For response shaping, see [`SHAPES.md`](SHAPES.md).

## Constructor

```go
import (
    "context"
    "github.com/makegov/tango-go"
)

client := tango.NewClient(
    tango.WithAPIKey("your-api-key"),                  // optional if TANGO_API_KEY env is set
    tango.WithBaseURL("https://tango.makegov.com"),    // default
    tango.WithTimeout(30 * time.Second),               // default
    tango.WithRetries(3),                              // default
    tango.WithRetryBackoff(250 * time.Millisecond),    // default
)
```

`NewClient(...Option)` always returns a usable `*Client`. Passing no options is valid when `TANGO_API_KEY` is set in the environment.

`*Client` is safe for concurrent use across goroutines. There is no per-request configuration — to deviate from the defaults on a specific call, construct a new client.

### Options

| Option | Signature | Default | Description |
| ------ | --------- | ------- | ----------- |
| `WithAPIKey` | `WithAPIKey(key string) Option` | reads `TANGO_API_KEY` env | Tango API key. Sent as `X-API-KEY` header on every request. Empty strings are ignored. |
| `WithBaseURL` | `WithBaseURL(url string) Option` | `https://tango.makegov.com` | API base URL. Override for staging/local. Trailing slash is normalized. |
| `WithTimeout` | `WithTimeout(d time.Duration) Option` | `30 * time.Second` | Per-request timeout. Each attempt is wrapped with `context.WithTimeout`. `0` disables the deadline. |
| `WithRetries` | `WithRetries(n int) Option` | `3` | Number of retry attempts on retryable failures. `0` disables retries (one attempt only). Negative values clamp to 0. |
| `WithRetryBackoff` | `WithRetryBackoff(d time.Duration) Option` | `250 * time.Millisecond` | Initial backoff between retries. Doubles each retry, capped at 10s. |
| `WithHTTPClient` | `WithHTTPClient(hc *http.Client) Option` | `&http.Client{}` | Custom `*http.Client`. Useful for injecting custom transports (proxies, tracing, instrumentation). The client's `Timeout` field is **ignored** — timeouts are managed via context. |
| `WithUserAgent` | `WithUserAgent(ua string) Option` | `tango-go/<version>` | Custom `User-Agent` header. |

### Environment variables

| Var | Purpose |
| --- | ------- |
| `TANGO_API_KEY` | Default API key when `WithAPIKey` is not passed. |
| `TANGO_BASE_URL` | Default base URL when `WithBaseURL` is not passed. Falls through to `tango.DefaultBaseURL` if neither is set. |

Precedence: explicit option > env var > default. `WithAPIKey("")` is a no-op (does not override the env value).

### Defaults table

| Setting | Default | Source |
| ------- | ------- | ------ |
| Base URL | `https://tango.makegov.com` (`tango.DefaultBaseURL`) | `shapes.go` |
| Timeout | `30 * time.Second` | `client.go` |
| Retries | `3` | `client.go` |
| Retry backoff base | `250 * time.Millisecond` | `client.go` |
| Max backoff cap | `10 * time.Second` | `transport.go` (`maxBackoff`) |
| Retryable statuses | `5xx`, `408`, `429`, transport errors | `errors.go` (`IsRetryable`) |
| User-Agent | `tango-go/<Version>` | `client.go` |

These match the Node and Python SDKs exactly.

## Retry semantics

A request is retried when:

- The HTTP status is `5xx` (any server error), `408` (Request Timeout), or `429` (Too Many Requests).
- The request fails at the network layer (DNS failure, connection refused, EOF, transport error).

Other `4xx` statuses (`400`, `401`, `403`, `404`, ...) are **not** retried — they surface as the appropriate typed error immediately.

`*TimeoutError` (the client-side timeout error) is itself retryable (its `StatusCode` is `0`, which `IsRetryable` treats as a transport-layer failure).

### Backoff math

- Attempt 1: initial request, no wait beforehand.
- Wait before attempt 2: `retryBackoff` (default 250ms).
- Wait before attempt 3: `retryBackoff * 2`.
- Wait before attempt 4: `retryBackoff * 4`.
- Each wait is **capped at 10 seconds** (`maxBackoff` in `transport.go`).

If the response includes a `Retry-After` header (typical on `429` and `503`), the client honors that value instead of computing its own backoff:

- Delta-seconds form (`Retry-After: 5`) → wait 5 seconds.
- HTTP-date form (`Retry-After: Wed, 21 Oct 2026 07:28:00 GMT`) → wait until that time.
- The honored wait is still capped at 10s.

### Disabling retries

```go
client := tango.NewClient(
    tango.WithAPIKey(os.Getenv("TANGO_API_KEY")),
    tango.WithRetries(0),
)
```

Useful for smoke tests, one-shot scripts, and any caller that wants to drive its own retry policy.

### IsRetryable

`tango.IsRetryable(err) bool` exposes the retry decision so callers building their own retry loops on top of the SDK can reuse the same logic:

```go
if err != nil && tango.IsRetryable(err) {
    // your own retry policy
}
```

## Error model

All HTTP-layer failures surface as one of the typed errors in `errors.go`:

```go
type APIError struct {
    StatusCode   int    // HTTP status, or 0 for transport-level errors
    Message      string // human-readable
    ResponseData any    // decoded JSON body of the error response (when JSON)
}

type AuthError       struct { *APIError }  // 401
type NotFoundError   struct { *APIError }  // 404
type ValidationError struct { *APIError }  // 400 (or client-side rejection)
type RateLimitError  struct {              // 429
    *APIError
    RetryAfter int    // seconds; 0 when Retry-After absent or unparseable
    LimitType  string // X-RateLimit-Type when present
}
type TimeoutError    struct { *APIError }  // client-side timeout; StatusCode == 0
```

Every typed child embeds `*APIError` and implements `Unwrap() error`, so `errors.As` traverses the chain cleanly:

```go
import "errors"

agency, err := client.GetAgency(ctx, "BOGUS")
if err != nil {
    var nf *tango.NotFoundError
    if errors.As(err, &nf) {
        log.Println("agency not found")
        return
    }

    var rle *tango.RateLimitError
    if errors.As(err, &rle) {
        log.Printf("rate limited; retry after %ds (limit: %s)", rle.RetryAfter, rle.LimitType)
        return
    }

    var ve *tango.ValidationError
    if errors.As(err, &ve) {
        log.Printf("bad request: %s (server detail: %v)", ve.Message, ve.ResponseData)
        return
    }

    var te *tango.TimeoutError
    if errors.As(err, &te) {
        log.Printf("timed out: %s", te.Message)
        return
    }

    // Catch-all
    var ae *tango.APIError
    if errors.As(err, &ae) {
        log.Printf("api error: status=%d message=%q", ae.StatusCode, ae.Message)
        return
    }

    return err
}
```

### Validation error message extraction

On `400`, the SDK extracts a useful message from common Tango error-body shapes:

- `{"detail": "..."}` (DRF default)
- `{"message": "..."}` or `{"error": "..."}` (top-level shorthand)
- The first string value in a field-error map (e.g. `{"recipient_uei": "Invalid format."}`)
- The first string in a field-error array

If none of those shapes match, `ValidationError.Message` falls back to `"invalid request parameters"` and the raw decoded body is preserved on `ResponseData`.

### Client-side validation

A handful of methods reject inputs locally before any network call. Examples:

- `GetEntity(ctx, "")` → `*ValidationError{"entity key is required"}`.
- `Resolve(ctx, ResolveInput{Name: ""})` → `*ValidationError{"Resolve: Name is required"}`.
- `CreateWebhookEndpoint(ctx, WebhookEndpointCreateInput{Name: ""})` → `*ValidationError{"CreateWebhookEndpoint: Name is required ..."}`.

These also unwrap to `*APIError` (StatusCode `0`), so the same `errors.As` chain works.

### Breaking change vs v0.1.0: `GetAgency` is typed

`GetAgency` now returns `*AgencyRecord` instead of `Record`. Callers that previously did `agency["name"]` should switch to the named fields:

```go
agency, err := client.GetAgency(ctx, "9700")
if err != nil {
    var nf *tango.NotFoundError
    if errors.As(err, &nf) { /* ... */ }
    return err
}

if agency.Name != nil {
    fmt.Println(*agency.Name)
}

// Forward-compatible fields the server adds that aren't in the typed
// surface yet are preserved on AgencyRecord.Extra.
if dept, ok := agency.Extra["new_field"]; ok { /* ... */ }
```

Pointer fields (`*string`, `*int`, etc.) distinguish "server omitted the field" from "server set it to zero". See `models.go` for the pattern and [`API_REFERENCE.md`](API_REFERENCE.md) for the full list of typed-return methods.

## Rate-limit observability

After every completed request (success or error), the client snapshots the rate-limit headers and exposes them via two accessors:

```go
info := client.RateLimitInfo()
fmt.Printf("%d/%d remaining; resets in %ds; type=%q\n",
    info.Remaining, info.Limit, info.ResetIn, info.LimitType)

if info.RetryAfter > 0 {
    fmt.Printf("server is asking us to back off for %ds\n", info.RetryAfter)
}

// All response headers from the last request (X-Request-Id, X-Tango-Trace-Id, etc.)
headers := client.LastResponseHeaders()
fmt.Println("request id:", headers.Get("X-Request-Id"))
```

`RateLimitInfo` fields are populated from the canonical headers (`X-RateLimit-Remaining`, `X-RateLimit-Limit`, `X-RateLimit-Reset`, `X-RateLimit-Type`, `Retry-After`). When a header is missing or unparseable, the integer fields are `-1` (and `LimitType` is `""`). `client.RateLimitInfo()` returns `nil` before any request has completed.

`client.LastResponseHeaders()` returns a clone of the response headers — modifying the returned map does not affect future requests.

Both accessors are safe to call concurrently with in-flight requests.

## Custom HTTP client

Inject a custom `*http.Client` via `WithHTTPClient`. Common reasons: proxying, tracing, retry instrumentation, swapping `*http.Transport`.

```go
import (
    "net/http"
    "github.com/makegov/tango-go"
)

tr := &http.Transport{
    Proxy: http.ProxyFromEnvironment,
    // ... custom transport config ...
}

logging := &loggingTransport{Wrapped: tr}

client := tango.NewClient(
    tango.WithAPIKey(os.Getenv("TANGO_API_KEY")),
    tango.WithHTTPClient(&http.Client{Transport: logging}),
)
```

> **Note.** Any `Timeout` set on your custom `*http.Client` is ignored — the SDK manages per-attempt timeouts via `context.WithTimeout(ctx, cfg.timeout)`. To override the SDK's timeout, use `WithTimeout`.

## Iterator surface

Every `List*` method that paginates has a sibling `Iterate*` method that returns `*tango.Iterator[Record]`. The iterator walks every page automatically — it auto-detects whether the endpoint uses `?page=` or `?cursor=` pagination from the server's `next` URL.

### Classic loop (works on any Go 1.22+)

```go
it := client.IterateContracts(ctx, &tango.ListContractsOptions{
    AwardingAgency: "9700",
    FiscalYear:     "2025",
})

for it.Next() {
    c := it.Item()
    fmt.Println(c["piid"], c["total_contract_value"])
}
if err := it.Err(); err != nil {
    log.Fatal(err)
}
```

### Range-over-func (Go 1.23+)

`Iterator[T].Seq()` returns `iter.Seq2[T, error]`:

```go
for c, err := range client.IterateContracts(ctx, opts).Seq() {
    if err != nil {
        return err
    }
    fmt.Println(c["piid"])
}
```

Breaking out of the range loop stops the iterator from fetching the next page — useful when you only need the first N matches.

### Semantics

- `Next()` returns `true` when `Item()` is valid; `false` when iteration is finished or an error occurred.
- `Item()` returns the current item; valid only immediately after a `true` return from `Next()`.
- `Err()` returns the error that ended iteration, if any. Nil when the iterator exhausted cleanly.
- `Seq()` yields `(item, nil)` for each result and a single trailing `(zero, err)` if iteration fails. The `Next/Item/Err` surface remains usable — `Seq` is additive sugar.

## Targeting staging or local

```go
client := tango.NewClient(
    tango.WithAPIKey(os.Getenv("TANGO_API_KEY")),
    tango.WithBaseURL("http://localhost:8000"),
)
```

The trailing slash is optional and is normalized away. Use `TANGO_BASE_URL` if you want the same code to pick up the right host based on the environment.
