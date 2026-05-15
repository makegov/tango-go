# Tango Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/makegov/tango-go.svg)](https://pkg.go.dev/github.com/makegov/tango-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> **In development — v0.1.0.** Not yet at sibling-SDK parity. The public API may shift before v1.0.0. Pin to a specific tag if you depend on this in production.

Official Go SDK for the [Tango API](https://tango.makegov.com) — federal contracts, IDVs, entities, opportunities, grants, vehicles, and more, with dynamic response shaping so you fetch only the fields you need.

The sibling SDKs (`tango-node` and `tango-python`) are at v1.0.0. This Go SDK ships the full transport, error model, retry/rate-limit handling, webhook signing, and a curated subset of resource methods. More endpoints land each 0.x release; see [CHANGELOG.md](CHANGELOG.md) for what's shipped.

## Features

- **Dynamic response shaping** — request exactly the fields you need via 21 built-in shape presets or a custom comma-separated field selector.
- **Typed errors** — `*AuthError`, `*NotFoundError`, `*ValidationError`, `*RateLimitError`, `*TimeoutError`, `*APIError`, all composable via `errors.As` / `errors.Is`.
- **Smart retries** — automatic backoff on 5xx / 408 / 429 / transport errors, honoring the server's `Retry-After` header.
- **Generic pagination** — `PaginatedResponse[T]` envelope + `Iterator[T]` that walks every page for you.
- **Webhook signing** — drop-in HMAC-SHA256 verification, `http.Handler` middleware included.
- **Standard-library only** — no transitive dependencies.

## Installation

```bash
go get github.com/makegov/tango-go
```

Requires Go 1.23 or later (for `Iterator[T].Seq()`'s `iter.Seq2`).

## Quick Start

### Initialize the client

```go
import "github.com/makegov/tango-go"

client := tango.NewClient(
    tango.WithAPIKey("your-api-key"),
)
```

Or, when `TANGO_API_KEY` is set in the environment:

```go
client := tango.NewClient()
```

### List agencies

```go
page, err := client.ListAgencies(ctx, nil)
if err != nil {
    log.Fatal(err)
}
for _, agency := range page.Results {
    fmt.Println(agency["agency_id"], agency["name"])
}
```

### Get a specific agency

`GetAgency` returns a typed `*AgencyRecord` — use the named fields, or `agency.Extra` for anything the server adds that isn't first-classed yet.

```go
agency, err := client.GetAgency(ctx, "9700")
if err != nil {
    log.Fatal(err)
}
fmt.Println(*agency.Name) // "Department of Defense"
```

### Search contracts with a minimal shape

```go
page, err := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions:    tango.ListOptions{Limit: 25, Shape: tango.ShapeContractsMinimal},
    AwardingAgency: "9700",
    AwardDateGte:   "2024-01-01",
    Sort:           "award_date",
    Order:          "desc",
})
```

### Walk every contract matching a filter

```go
iter := client.IterateContracts(ctx, &tango.ListContractsOptions{
    AwardingAgency: "9700",
    FiscalYear:     "2025",
})
for iter.Next() {
    c := iter.Item()
    fmt.Println(c["piid"], c["total_contract_value"])
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

### Get a fully-shaped entity

```go
entity, err := client.GetEntity(ctx, "ABC123DEF456", &tango.GetEntityOptions{
    Shape: tango.ShapeEntitiesComprehensive,
})
```

## Authentication

### With API key

```go
client := tango.NewClient(tango.WithAPIKey("your-key"))
```

### From environment variable (`TANGO_API_KEY`)

```bash
export TANGO_API_KEY=your-key
```

```go
client := tango.NewClient()
```

You can also override the base URL via `TANGO_BASE_URL` or `tango.WithBaseURL("https://staging.tango.makegov.com")` — useful for staging environments.

## Core Concepts

### Dynamic Response Shaping

Every list and get endpoint accepts a `shape` parameter. The SDK ships 21 presets matching the Node/Python SDKs:

```go
// Use a preset
page, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{Shape: tango.ShapeContractsMinimal},
})

// Or roll your own
page, _ = client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{Shape: "key,piid,recipient(uei,display_name)"},
})
```

### Pagination

List endpoints support both page-based (`Page`) and cursor-based (`Cursor`) pagination. `PaginatedResponse[T]` carries both `Next` (the full URL) and `Cursor` (extracted from `Next` for keyset endpoints).

```go
page, _ := client.ListIDVs(ctx, &tango.ListIDVsOptions{
    ListOptions: tango.ListOptions{Limit: 50},
})
// Fetch the next page:
nextPage, _ := client.ListIDVs(ctx, &tango.ListIDVsOptions{
    ListOptions: tango.ListOptions{Limit: 50, Cursor: page.Cursor},
})
```

Or just iterate:

```go
iter := client.IterateIDVs(ctx, nil)
for iter.Next() {
    // iter.Item() is the next IDV
}
```

### Rate limits

After each request, `client.RateLimitInfo()` returns a snapshot of the rate-limit headers:

```go
info := client.RateLimitInfo()
fmt.Printf("%d/%d remaining; resets in %ds\n", info.Remaining, info.Limit, info.ResetIn)
```

The client also automatically retries 429s, honoring `Retry-After`.

### Webhook verification

```go
import "github.com/makegov/tango-go/webhooks"

http.Handle("/tango-webhook", webhooks.Middleware(os.Getenv("WEBHOOK_SECRET"),
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // signature already verified; r.Body is reset and ready to read
        var delivery map[string]any
        json.NewDecoder(r.Body).Decode(&delivery)
        // process delivery...
        w.WriteHeader(http.StatusNoContent)
    }),
))
```

For lower-level use:

```go
ok := webhooks.Verify(rawBody, r.Header.Get(webhooks.SignatureHeader), secret)
```

## Error Handling

Every error returned by the SDK can be unwrapped to one of the typed error types:

```go
_, err := client.GetAgency(ctx, "bogus")
var nf *tango.NotFoundError
if errors.As(err, &nf) {
    // handle 404 specifically
}

var rle *tango.RateLimitError
if errors.As(err, &rle) {
    log.Printf("rate limited; retry after %ds (limit type: %s)", rle.RetryAfter, rle.LimitType)
}

// Catch-all
var apiErr *tango.APIError
if errors.As(err, &apiErr) {
    log.Printf("status=%d data=%v", apiErr.StatusCode, apiErr.ResponseData)
}
```

`tango.IsRetryable(err)` reports whether the SDK considers `err` retryable (it already retries internally — this helper is for callers building their own retry policies).

## API Methods

The SDK exposes ~94 methods on `*Client` covering every endpoint in the sibling Node/Python SDKs. The most-used 15 are listed here; the **full method-by-method reference lives in [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md)**.

| Resource | List | Get | Iterate |
| ---- | ---- | ---- | ---- |
| Agencies | `ListAgencies` | `GetAgency` *(typed: `*AgencyRecord`)* | — |
| Contracts | `ListContracts` | — | `IterateContracts` |
| Entities | `ListEntities` | `GetEntity` | `IterateEntities` |
| IDVs | `ListIDVs` | `GetIDV` | `IterateIDVs` |
| Vehicles | `ListVehicles` | `GetVehicle` | `IterateVehicles` |
| OTAs | `ListOTAs` | `GetOTA` | `IterateOTAs` |
| OTIDVs | `ListOTIDVs` | `GetOTIDV` | `IterateOTIDVs` |
| Opportunities | `ListOpportunities` | — | `IterateOpportunities` |
| Notices | `ListNotices` | — | `IterateNotices` |
| Forecasts | `ListForecasts` | — | `IterateForecasts` |
| Grants | `ListGrants` | — | `IterateGrants` |
| Protests | `ListProtests` | `GetProtest` *(typed: `*ProtestRecord`)* | `IterateProtests` |
| IT Dashboard | `ListItDashboard` | `GetItDashboard` | `IterateItDashboard` |
| NAICS / PSC | `ListNAICS` / `ListPSC` | `GetNAICS` / `GetPSC` | — |
| Webhooks (CRUD) | `ListWebhookEndpoints` / `ListWebhookAlerts` | `Get…` / `Create…` / `Update…` / `Delete…` | — |

Sub-resources and lookups are also covered: `ListEntityContracts` / `IDVs` / `OTAs` / `OTIDVs` / `Subawards` / `Lcats`, `ListIDVAwards` / `ChildIDVs` / `Transactions` / `Lcats`, `ListAgencyAwardingContracts` / `FundingContracts`, `ListVehicleAwardees` / `Orders`, `ListOTIDVAwards`, `ListGsaElibraryContracts`, `ListBusinessTypes`, `ListOffices`, `ListDepartments` *(deprecated)*, `ListMasSins`, `ListAssistanceListings`, `ListLcats`, plus all three metrics getters (`GetNAICSMetrics` / `GetPSCMetrics` / `GetEntityMetrics`) and the dispatcher `ListMetrics`. Meta: `Resolve`, `Validate`, `GetVersion`, `ListAPIKeys`, `SearchOpportunityAttachments`. Webhook signing: `webhooks.Generate` / `Verify` / `Parse` / `VerifyRequest` / `Middleware`.

See [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) for full signatures, filter fields, and quirks per method.

## Project Structure

```
.
├── client.go          # Client + NewClient + accessors
├── options.go         # Functional options
├── transport.go       # HTTP, auth, retries, rate-limit parsing
├── errors.go          # Typed error tree
├── pagination.go      # PaginatedResponse[T] + Iterator[T]
├── shapes.go          # Shape preset constants + DefaultBaseURL
├── internal.go        # ListOptions + generic helpers
├── version.go         # SDK version constant
├── contracts.go       # ListContracts + IterateContracts
├── entities.go        # ListEntities + GetEntity + IterateEntities
├── idvs.go            # ListIDVs + GetIDV + IterateIDVs
├── vehicles.go        # ListVehicles + GetVehicle
├── opportunities.go   # opportunities, notices, forecasts, grants
├── lookups.go         # agencies, organizations, NAICS, PSC, subawards, version
├── resolve.go         # Resolve + Validate
├── client_test.go     # Transport + resource tests
├── webhooks/
│   ├── signing.go     # Generate / Verify / Parse
│   ├── handler.go     # VerifyRequest + Middleware
│   └── signing_test.go
├── examples/          # Runnable example programs
├── CHANGELOG.md
├── LICENSE
└── README.md
```

## Development

```bash
go test ./...            # run unit tests
go test -race ./...      # with the race detector
go vet ./...             # static checks
```

## Requirements

- Go **1.23** or later.

## Documentation

In-repo guides under [`docs/`](docs/):

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — package layout, request lifecycle, design rationale
- [`docs/CLIENT.md`](docs/CLIENT.md) — `NewClient` options, env vars, retry semantics, error model, rate-limit observability, iterator surface
- [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) — method-by-method reference for every public method (~94)
- [`docs/SHAPES.md`](docs/SHAPES.md) — shape grammar, the 21 `Shape*` presets, `Flat` / `FlatLists`, trade-offs
- [`docs/WEBHOOKS.md`](docs/WEBHOOKS.md) — receiving deliveries, CRUD methods, troubleshooting

External:

- SDK reference: <https://pkg.go.dev/github.com/makegov/tango-go>
- API reference: <https://docs.makegov.com>
- Sibling SDKs: [`@makegov/tango-node`](https://github.com/makegov/tango-node) · [`tango-python`](https://github.com/makegov/tango-python)

## License

MIT — see [LICENSE](LICENSE).

## Support

Open an issue at <https://github.com/makegov/tango-go/issues> or email support@makegov.com.

## Contributing

PRs welcome. Run `go test ./...` and `go vet ./...` before opening. For larger work, open an issue first so we can align on the API shape — public methods on `*Client` are a stability surface from 1.0.0 onward.
