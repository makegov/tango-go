# Webhooks

This guide covers everything `github.com/makegov/tango-go` provides for **building, testing, and operating webhook integrations against the Tango API**: signing, signature verification, the `http.Handler` middleware, and the client-side CRUD methods for managing endpoints and alerts.

The SDK has two distinct surfaces for webhooks:

| Concern | Where | Imports |
| ------- | ----- | ------- |
| Sign / verify / receive deliveries | `github.com/makegov/tango-go/webhooks` | stdlib only |
| Create / list / update / delete endpoints + alerts (CRUD over the API) | `github.com/makegov/tango-go` (`*Client` methods) | full API client |

A webhook receiver that only needs to verify deliveries can import `webhooks/` alone — it doesn't pull the HTTP client, the retry loop, or any resource methods. The two surfaces are wire-compatible: signatures produced by `webhooks.Generate` verify correctly against Tango deliveries and vice versa.

For the API-level contract (signing scheme, event taxonomy, retry behavior), see the [Tango Webhooks Partner Guide](https://docs.makegov.com/webhooks-user-guide/).

## Concepts

Tango webhooks have three pieces of state:

| Concept | What it is | Tango term |
| ------- | ---------- | ---------- |
| **Endpoint** | A URL Tango POSTs to, plus a generated signing secret | `WebhookEndpoint` |
| **Alert** | A saved query-filter that fires deliveries when matching records appear | `WebhookAlert` |
| **Delivery** | A single signed POST Tango makes when a matching event fires | (the request itself) |

A typical setup:

1. **Create an endpoint** with the public URL of your handler. The API returns a `secret` on creation — save it; it's used to sign every delivery.
2. **Create one or more alerts** describing the records your handler cares about (e.g. new IT-services contracts).
3. **Tango POSTs to your endpoint** when matching records appear. The body is JSON; the header `X-Tango-Signature: sha256=<hex>` is the HMAC-SHA256 of the raw body bytes keyed by your endpoint's secret.
4. **Your handler verifies the signature**, parses the body, and acts on it.

## Receiving deliveries

### The `webhooks` subpackage

```go
import "github.com/makegov/tango-go/webhooks"
```

Surface:

| Symbol | Purpose |
| ------ | ------- |
| `SignatureHeader` (const) | `"X-Tango-Signature"` |
| `SignaturePrefix` (const) | `"sha256="` |
| `Generate(body []byte, secret string) string` | Produce a wire-format `sha256=<hex>` signature for tests. |
| `Verify(body []byte, header, secret string) bool` | Constant-time check; never panics. |
| `Parse(header string) (ParsedSignature, bool)` | Decompose a header value for debugging. |
| `VerifyRequest(r *http.Request, secret string) ([]byte, error)` | Read + verify; resets `r.Body` so handlers can re-read. |
| `Middleware(secret string, next http.Handler) http.Handler` | Drop-in `http.Handler` wrapper. |
| `ErrInvalidSignature` | Returned by `VerifyRequest` on failure. |

### Drop-in middleware

The fastest way to wire up a verified handler is `webhooks.Middleware`:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/makegov/tango-go/webhooks"
)

func main() {
    secret := os.Getenv("TANGO_WEBHOOK_SECRET")

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // signature already verified; r.Body has been reset and is readable
        var delivery map[string]any
        if err := json.NewDecoder(r.Body).Decode(&delivery); err != nil {
            http.Error(w, "bad json", http.StatusBadRequest)
            return
        }
        log.Printf("received: %v", delivery["event_type"])
        w.WriteHeader(http.StatusNoContent)
    })

    http.Handle("/tango-webhook", webhooks.Middleware(secret, handler))

    log.Fatal(http.ListenAndServe(":3000", nil))
}
```

`Middleware` reads the request body, verifies the signature, and either returns `401 Unauthorized` immediately or passes through to `next` with the body re-set so the handler can read it.

### Verifying manually

For frameworks where the middleware pattern doesn't fit (gorilla, chi, fiber, custom routers), call `webhooks.VerifyRequest` directly:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, err := webhooks.VerifyRequest(r, secret)
    if err != nil {
        // err is webhooks.ErrInvalidSignature on a bad / missing signature
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    // body is the raw bytes; r.Body has been re-set to the same bytes
    var delivery map[string]any
    _ = json.Unmarshal(body, &delivery)
    // ...
}
```

Or, when you already have the raw body bytes from somewhere else:

```go
ok := webhooks.Verify(rawBody, r.Header.Get(webhooks.SignatureHeader), secret)
if !ok {
    http.Error(w, "invalid signature", http.StatusUnauthorized)
    return
}
```

### Signature format

```text
X-Tango-Signature: sha256=<lowercase hex HMAC-SHA256 of raw body>
```

- HMAC-SHA256 keyed by the endpoint's secret.
- Computed over the **raw request body bytes** — re-serializing a parsed JSON document will produce a different signature because of whitespace, key ordering, and float formatting differences.
- Comparison is constant-time (`crypto/hmac.Equal`) — `Verify` never short-circuits on a per-byte mismatch.
- `Parse` accepts both the canonical `sha256=<hex>` form and a bare hex string (legacy compatibility, mirroring Node and Python).

### Generating signatures for tests

`webhooks.Generate` is useful for unit-testing your handler without a live Tango account:

```go
func TestHandler(t *testing.T) {
    body := []byte(`{"event_type":"alerts.contract.match","matches":{}}`)
    sig := webhooks.Generate(body, "test_secret")
    // sig = "sha256=<hex>"

    req := httptest.NewRequest("POST", "/tango-webhook", bytes.NewReader(body))
    req.Header.Set(webhooks.SignatureHeader, sig)

    rec := httptest.NewRecorder()
    handler(rec, req)

    if rec.Code != http.StatusNoContent {
        t.Fatalf("expected 204, got %d", rec.Code)
    }
}
```

## Managing endpoints + alerts via the API

All CRUD methods are on `*tango.Client`. They require **Large / Enterprise** tier access on the Tango side.

```go
import "github.com/makegov/tango-go"

client := tango.NewClient(tango.WithAPIKey(os.Getenv("TANGO_API_KEY")))
```

### Endpoints

An endpoint is the URL Tango POSTs to, paired with a signing secret.

| Method | Endpoint | Notes |
| ------ | -------- | ----- |
| `ListWebhookEndpoints(ctx, *ListOptions)` | `GET /api/webhooks/endpoints/` | Returns `*PaginatedResponse[WebhookEndpoint]`. |
| `GetWebhookEndpoint(ctx, id)` | `GET /api/webhooks/endpoints/{id}/` | |
| `CreateWebhookEndpoint(ctx, WebhookEndpointCreateInput)` | `POST /api/webhooks/endpoints/` | `Name` + `CallbackURL` required client-side. |
| `UpdateWebhookEndpoint(ctx, id, WebhookEndpointUpdateInput)` | `PATCH /api/webhooks/endpoints/{id}/` | Pointer fields — nil means "don't change". |
| `DeleteWebhookEndpoint(ctx, id)` | `DELETE /api/webhooks/endpoints/{id}/` | |
| `TestWebhookEndpoint(ctx, endpointID)` | `POST /api/webhooks/endpoints/test-delivery/` | Forces a test delivery; returns the result. |

```go
// Create — Name is unique per user; save the returned Secret immediately.
created, err := client.CreateWebhookEndpoint(ctx, tango.WebhookEndpointCreateInput{
    Name:        "production-handler",
    CallbackURL: "https://example.com/tango-webhook",
    // IsActive: defaults to true server-side when omitted
})
if err != nil {
    return err
}
secret := *created.Secret  // pointer field — present on create
fmt.Println("secret:", secret)

// List
list, _ := client.ListWebhookEndpoints(ctx, &tango.ListOptions{Limit: 25})
for _, ep := range list.Results {
    fmt.Println(*ep.ID, *ep.Name, *ep.CallbackURL, *ep.IsActive)
}

// Update
isActive := false
_, _ = client.UpdateWebhookEndpoint(ctx, *created.ID, tango.WebhookEndpointUpdateInput{
    IsActive: &isActive,
})

// Delete
_ = client.DeleteWebhookEndpoint(ctx, *created.ID)
```

`WebhookEndpoint` fields are all pointers (`*string`, `*bool`) so the zero value distinguishes "absent" from "explicit empty". Forward-compatible fields the server adds end up in `WebhookEndpoint.Extra` (a `map[string]any`).

### Alerts

Alerts are the filter-based subscription API. They replaced subject-based subscriptions in v0.4.0 of the canonical SDKs — alerts are the only way to subscribe.

| Method | Endpoint | Notes |
| ------ | -------- | ----- |
| `ListWebhookAlerts(ctx, *ListOptions)` | `GET /api/webhooks/alerts/` | Returns `*PaginatedResponse[WebhookAlert]`. |
| `GetWebhookAlert(ctx, id)` | `GET /api/webhooks/alerts/{id}/` | |
| `CreateWebhookAlert(ctx, WebhookAlertCreateInput)` | `POST /api/webhooks/alerts/` | `Name`, `QueryType`, `Filters` required. |
| `UpdateWebhookAlert(ctx, id, WebhookAlertUpdateInput)` | `PATCH /api/webhooks/alerts/{id}/` | Only `Name`, `Frequency`, `CronExpression`, `IsActive` are writable. |
| `DeleteWebhookAlert(ctx, id)` | `DELETE /api/webhooks/alerts/{id}/` | |

```go
alert, err := client.CreateWebhookAlert(ctx, tango.WebhookAlertCreateInput{
    Name:      "new-it-cloud-contracts",
    QueryType: "contract",  // singular — required
    Filters: map[string]any{
        "naics":           "541511",
        "awarding_agency": "9700",
    },
    // Frequency / CronExpression / Endpoint are optional pointer fields
})
```

> **Gotcha.** `QueryType` is singular — `"contract"`, not `"contracts"`. The SDK does not validate this client-side; the server returns 400 with a clear message if you pass the plural form.

### Event types and sample payloads

```go
// What event types can I subscribe to?
info, _ := client.ListWebhookEventTypes(ctx)
for _, et := range info.EventTypes {
    fmt.Println(*et.EventType, "-", *et.Description)
}

// What does a payload look like?
sample, _ := client.GetWebhookSamplePayload(ctx, "alerts.contract.match")
fmt.Println(*sample.EventType)
fmt.Println(*sample.SignatureHeader)  // example X-Tango-Signature
fmt.Println(sample.SampleDelivery.Events)

// Get every event type's sample at once
all, _ := client.GetWebhookSamplePayload(ctx, "")
for evType, body := range all.Samples {
    fmt.Println(evType, body.SampleDelivery.Events)
}
```

`GetWebhookSamplePayload(ctx, "")` (empty `eventType`) returns the all-types variant; non-empty returns the single-event-type variant. Both responses come back as a `*WebhookSamplePayloadResponse` — branch on which fields are populated. The `*Note` field is a server-supplied helper string.

### Test delivery

Force the API to POST a real test delivery to a registered endpoint:

```go
result, err := client.TestWebhookEndpoint(ctx, "ENDPOINT_UUID")
if err != nil {
    return err
}

fmt.Println("success:", *result.Success)
if result.StatusCode != nil {
    fmt.Println("status:", *result.StatusCode)
}
if result.ResponseTimeMs != nil {
    fmt.Println("rtt:", *result.ResponseTimeMs, "ms")
}
if result.Error != nil {
    fmt.Println("error:", *result.Error)
}
```

`success: false` with a populated `StatusCode` + `ResponseBody` means Tango reached your endpoint but got a non-2xx response back. Check your handler's logs.

## Common workflows

### "Set me up to receive contract-match alerts from scratch"

```go
client := tango.NewClient(tango.WithAPIKey(os.Getenv("TANGO_API_KEY")))

// 1. See what's available
info, _ := client.ListWebhookEventTypes(ctx)
for _, et := range info.EventTypes {
    fmt.Println(*et.EventType)
}

// 2. Create endpoint (expose your handler via ngrok or cloudflared first)
endpoint, err := client.CreateWebhookEndpoint(ctx, tango.WebhookEndpointCreateInput{
    Name:        "my-handler",
    CallbackURL: "https://<your-tunnel>.ngrok.io/tango-webhook",
})
if err != nil { return err }
// SAVE the secret — you need it to verify incoming deliveries
fmt.Println("secret:", *endpoint.Secret)

// 3. Create an alert
_, err = client.CreateWebhookAlert(ctx, tango.WebhookAlertCreateInput{
    Name:      "new-it-cloud-contracts",
    QueryType: "contract",
    Filters:   map[string]any{"naics": "541511"},
})
if err != nil { return err }

// 4. Force a test delivery to verify your handler is reachable
result, _ := client.TestWebhookEndpoint(ctx, *endpoint.ID)
fmt.Println("test delivery:", *result.Success)
```

### "Verify a delivery in any HTTP framework"

```go
// The pattern is the same in net/http, gorilla, chi, fiber, etc.:
//   1. Get the raw body BEFORE any JSON parsing middleware
//   2. Get the X-Tango-Signature header
//   3. Call webhooks.Verify(rawBody, header, secret)

func handleTangoWebhook(rawBody []byte, signatureHeader string) (map[string]any, error) {
    if !webhooks.Verify(rawBody, signatureHeader, secret) {
        return nil, webhooks.ErrInvalidSignature
    }
    var delivery map[string]any
    if err := json.Unmarshal(rawBody, &delivery); err != nil {
        return nil, err
    }
    return delivery, nil
}
```

### "Test my handler in a unit test (no network)"

```go
import (
    "bytes"
    "net/http/httptest"
    "testing"

    "github.com/makegov/tango-go/webhooks"
)

func TestHandler(t *testing.T) {
    body := []byte(`{"event_type":"alerts.contract.match","matches":{}}`)
    sig := webhooks.Generate(body, "test_secret")

    req := httptest.NewRequest("POST", "/tango-webhook", bytes.NewReader(body))
    req.Header.Set(webhooks.SignatureHeader, sig)
    rec := httptest.NewRecorder()

    myHandler.ServeHTTP(rec, req)

    if rec.Code != http.StatusNoContent {
        t.Fatalf("got %d, want 204", rec.Code)
    }
}
```

## Troubleshooting

**`Verify` always returns `false`.** Verify on raw bytes, not on a re-serialized parsed body. Most frameworks expose the raw body separately from a parsed JSON shortcut — use the raw one. Also confirm you're passing the right secret (the one returned by `CreateWebhookEndpoint`, not your API key).

**`VerifyRequest` returns `ErrInvalidSignature`.** Two likely causes: (1) the body has already been read by some other middleware and `r.Body` is empty by the time you reach the verify call, or (2) the signature header is missing. Try logging `r.Header.Get(webhooks.SignatureHeader)` and `len(rawBody)` to confirm.

**`CreateWebhookEndpoint` returns 400 / `*ValidationError`.** Endpoint names are unique per user — if you've already created one with that `Name`, pick a different name or call `ListWebhookEndpoints` to find the existing one and reuse its ID.

**`CreateWebhookAlert` returns `*ValidationError` with "query_type is required".** `QueryType` is singular — `"contract"`, not `"contracts"`. The SDK passes the value through; the server enforces.

**`TestWebhookEndpoint` returns `result.Success: false`.** Tango reached your endpoint but got a non-2xx back. Inspect `*result.StatusCode` and `*result.ResponseBody` — your handler probably 500'd. Don't forget that `Middleware` returns 401 if the signature doesn't verify, which `TestWebhookEndpoint` will surface as `Success: false` with `StatusCode: 401`.

**`GetWebhookSamplePayload` returns 401 / `*AuthError`.** The sample-payload endpoint requires authentication. Set `TANGO_API_KEY` or pass `tango.WithAPIKey(...)` to `NewClient`.

**`ListWebhookAlerts` returns an empty array unexpectedly.** Alerts are scoped to the authenticated user. Also confirm the endpoint UUID associated with your alerts is still alive — alerts pointing at a deleted endpoint aren't automatically cleaned up.
