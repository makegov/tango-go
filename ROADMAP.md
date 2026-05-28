# ROADMAP

This roadmap tracks the Go SDK only. The goal is to stay closely aligned with the Tango Python and Node SDKs while keeping a first-class Go experience (idiomatic options, typed errors, generics where they pay rent).

## Now

- [X] Port the full Tango Python/Node SDK client surface to Go.
- [X] Implement dynamic response shaping (shape, flat, flat_lists) with full parity.
- [X] Typed error tree with `errors.Is` / `errors.As` ergonomics.
- [X] Generic `PaginatedResponse[T]` and `Iterator[T]` for cursor + page-numbered pagination.
- [X] Webhook signing + verification (`tango/webhooks` subpackage) wire-compatible with sibling SDKs.
- [X] Webhook CRUD via the API client (endpoints, alerts, event types, test delivery, sample payloads).
- [X] Sub-resource walks for IDVs, entities, agencies, vehicles.
- [X] OTAs / OTIDVs, GSA eLibrary, IT Dashboard, protests, LCATs.

## 0.2 (API sync to v4.6.9)

- [X] Budget surface: `ListBudgetAccounts` / `IterateBudgetAccounts`, `GetBudgetAccount`, `GetBudgetAccountQuarters`, `GetBudgetAccountRecipients` (`/api/budget/accounts/`).
- [X] Singleton detail GETs: `GetContract`, `GetOpportunity`, `GetNotice`, `GetForecast`, `GetGrant`, `GetSubaward`.
- [X] Contract sub-routes: `ListContractSubawards`, `ListContractTransactions`.
- [X] `GetEntityBudgetFlows` (`/api/entities/{uei}/budget-flows/`).
- [X] `GrantID` filter on grants; `Cage` filter on entities.
- [X] Removed fabricated `GetIDVSummary` / `ListIDVSummaryAwards` (paths never existed upstream).

## Next

- [ ] Comprehensive integration tests against the live Tango API.
- [ ] Typed structs for remaining lookup/metric endpoints (NAICS, PSC, entity metrics, attachment search).
- [ ] More targeted examples and recipes in `docs/` and `examples/`.
- [ ] `WithLogger` option for structured logging hooks.

## Later

- [ ] Generated method coverage from the Tango OpenAPI spec to keep parity drift to zero.
- [ ] Optional adapters for alternative HTTP transports (e.g. `*http.Client` wrappers with tracing).
- [ ] Browser / WASM-friendly build (if there's demand).

## Maybe Someday

- [ ] CLI utilities built on top of the Go SDK.
- [ ] First-class OpenTelemetry spans for all HTTP calls.
