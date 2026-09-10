<!-- markdownlint-disable MD024, MD013 -->
# Changelog

All notable changes to `github.com/makegov/tango-go` will be documented in this file.

This project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- **`attachments(extracted_text)` — SLED document bodies on the Small plan and above** (Tango API 4.25.1; parity with tango-python and tango-node). This SDK returns `Record`, so the leaf needs no schema change — what it needed was saying so. Documented on `GetSledOpportunity`, on `ShapeSledOpportunitiesComprehensive` and in `docs/API_REFERENCE.md`: the leaf must be **named** (no `Shape*` preset includes it, and `attachments(*)` does not carry it, because the API resolves the body only for a caller who asked); the **key is absent rather than null** when the text is not being served; and a **contested document never returns text at any plan**. `TestSledShapesDoNotNameThePaidDocumentBody` pins the presets. Searching document text stays ungated on every plan and returns no fragment of it.

- **State, local and education (SLED) procurement** (Tango API 4.25.0; parity with tango-python and tango-node). `ListSledOpportunities` / `GetSledOpportunity`, `ListSledOpportunityRevisions`, `GetSledCoverage`, `ListSledForecasts` / `GetSledForecast`, plus `IterateSledOpportunities` and `IterateSledForecasts`. Three options structs cover every one of the API's 27 solicitation filters and 13 forecast filters as a named field, and five `Shape*` presets land in `shapes.go`.

  Four behaviors are documented on the options structs because each misleads a caller who assumes federal semantics. **Leaving both `Status` and `Active` unset returns open solicitations only** — that default is the API's, and `ListSledOpportunities` deliberately does not synthesize `status=open`, since doing so would make `Active: boolPtr(false)` unreachable (pinned by a test). **`Status` is Tango-derived and refreshed every fifteen minutes**; the portal's own word is served as `source_status`, is frozen at last capture, and is not a liveness filter. **Category scheme tagging is mid-migration**, so `Naics` matches only the small tagged share and `CategoryCode` is the escape hatch. And **`meta.attachment_count` can be lower than `len(attachments)`**, because an auto-generated portal cover sheet is listed and flagged `is_generated_summary` but excluded from the count.

  `Active`, `HasDocuments` and `SourceDeclared` are `*bool` rather than `bool`, so `false` reaches the server as a filter value instead of vanishing into the zero value — the same reason `setIfNotNilBool` exists.

  `ShapeSledRevisionsMinimal` omits `changes` on purpose: the per-field before/after needs a Small plan, so naming it in a suggested shape would 403 a Free caller. `changed_fields` is in the shape and available at every plan.

### Documentation

- New **State & Local (SLED)** section in `docs/API_REFERENCE.md` covering all six methods and both defaults that surprise people.
- `docs/WEBHOOKS.md` troubleshooting gained the date-lapse rule and its one exception. An exclusion or DIBBS solicitation reaching its date fires nothing, because open/closed is derived at query time — but `alerts.sled_opportunity.match` **does** fire on a closing, since SLED liveness is a stored column a fifteen-minute sweep writes.

## [0.1.0] - 2026-05-15

First public release of the Tango Go SDK.

This is an initial **v0.1.0** rather than v1.0.0 (which `tango-node` and `tango-python` are at) — the surface here is the core of the API, not the full surface. The transport, error model, retry/rate-limit handling, and webhook signing are at sibling-SDK quality and are not expected to change. The list of resource methods is intentionally a subset; more endpoints will land in 0.x releases before tagging 1.0.0.

### Added

#### Client core

- `tango.NewClient(...Option)` constructor with functional options:
  `WithAPIKey`, `WithBaseURL`, `WithTimeout`, `WithRetries`, `WithRetryBackoff`,
  `WithHTTPClient`, `WithUserAgent`. `TANGO_API_KEY` and `TANGO_BASE_URL`
  environment variables are honored as fallbacks.
- `X-API-KEY` authentication (matches `tango-node` / `tango-python`).
- Typed error tree: `*APIError` (base), `*AuthError` (401), `*NotFoundError`
  (404), `*ValidationError` (400), `*RateLimitError` (429, with `RetryAfter`
  + `LimitType`), `*TimeoutError`. All compose via `errors.As`/`errors.Is`.
  `IsRetryable(err)` exposes the SDK's retry decision.
- Automatic retry on 5xx / 408 / 429 / network errors with exponential
  backoff (default 250ms base, doubling, capped at 10s). Server
  `Retry-After` header overrides backoff.
- `Client.RateLimitInfo()` and `Client.LastResponseHeaders()` for
  observability (parity with the Python `rate_limit_info` /
  `last_response_headers` properties).
- Generic `PaginatedResponse[T]` envelope; cursor extracted from `next`
  URL for keyset endpoints.
- `Iterator[T]` (returned by every `IterateXxx` method) walks every
  result of a paginated endpoint, following either `?page=` or
  `?cursor=`. Go 1.23+ `Iterator[T].Seq()` returns an `iter.Seq2` for
  range-over-func use.
- All 21 shape presets exported as `ShapeXxx` constants — same values
  as the JS/Python `ShapeConfig.*` enums.

#### Typed models

- Typed response models in `models.go` mirroring the named interfaces in
  `tango-node/src/types.ts` and `tango-node/src/models/Webhooks.ts`:
  `AgencyRecord`, `WebhookEndpoint`, `WebhookEventType`,
  `WebhookEventTypesResponse`, `WebhookSampleDelivery`,
  `WebhookSamplePayloadResponse`, `WebhookTestDeliveryResult`,
  `WebhookAlert`. Each carries an `Extra map[string]any` for
  forward-compatible fields the server adds, populated via a custom
  `UnmarshalJSON`. Pointer-typed optional fields distinguish "absent"
  from server-supplied zero values.
- Input types: `WebhookEndpointCreateInput`,
  `WebhookEndpointUpdateInput`, `WebhookAlertCreateInput`,
  `WebhookAlertUpdateInput`.
- `Client.GetAgency` returns a typed `*AgencyRecord` — use the named
  fields (`AgencyID`, `Name`, `Abbreviation`, `Code`, `Department`) or
  `AgencyRecord.Extra` for forward-compatible fields not in the typed
  surface.
- `Client.GetProtest` returns a typed `*ProtestRecord` (list results
  stay as `Record`).

#### Resource methods

Core resources:

- `ListAgencies`, `GetAgency`, `ListContracts` / `IterateContracts`,
  `ListEntities` / `GetEntity` / `IterateEntities`, `ListIDVs` /
  `GetIDV` / `IterateIDVs`, `ListVehicles` / `GetVehicle` /
  `IterateVehicles`, `ListOpportunities` / `IterateOpportunities`,
  `ListNotices` / `IterateNotices`, `ListForecasts` /
  `IterateForecasts`, `ListGrants` / `IterateGrants`,
  `ListOrganizations` / `GetOrganization`, `ListNAICS` / `GetNAICS`,
  `ListPSC` / `GetPSC`, `ListSubawards`, `Resolve`, `Validate`,
  `GetVersion`.

Sub-resources:

- Entities: `ListEntityContracts`, `ListEntityIDVs`, `ListEntityOTAs`,
  `ListEntityOTIDVs`, `ListEntitySubawards`, `ListEntityLcats`.
- IDVs: `ListIDVAwards`, `ListIDVChildIDVs`, `ListIDVTransactions`,
  `ListIDVLcats`.
- Agencies: `ListAgencyAwardingContracts`,
  `ListAgencyFundingContracts`.
- Vehicles: `ListVehicleAwardees`, `ListVehicleOrders`.

OTAs / OTIDVs / Protests / IT Dashboard / GSA eLibrary / LCATs:

- `ListOTAs`, `IterateOTAs`, `GetOTA`.
- `ListOTIDVs`, `IterateOTIDVs`, `GetOTIDV`, plus
  `ListOTIDVAwards` / `IterateOTIDVAwards`.
- `ListProtests`, `IterateProtests`, `GetProtest`.
- `ListItDashboard`, `IterateItDashboard`, `GetItDashboard`. Filters
  are tier-gated by the API; see godoc on `ListItDashboardOptions` for
  the free / pro / business+ split.
- `ListGsaElibraryContracts`, `IterateGsaElibraryContracts`,
  `GetGsaElibraryContract`.
- `ListLcats`, `IterateLcats` — router: dispatches to
  `/api/entities/{uei}/lcats/` when `UEI` is set, or
  `/api/idvs/{key}/lcats/` when `IDVKey` is set. Returns
  `*ValidationError` if neither is set.

Lookups, reference data, and metrics:

- `ListBusinessTypes`, `GetBusinessType`.
- `ListOffices` (with `Search`), `GetOffice`.
- `ListDepartments`, `GetDepartment` — `ListDepartments` is marked
  `Deprecated:` in godoc; prefer `ListOrganizations` with `Level=1`.
- `ListMasSins` (with `Search`), `GetMasSin`.
- `ListAssistanceListings`, `GetAssistanceListing`.
- `GetNAICSMetrics`, `GetPSCMetrics`, `GetEntityMetrics`, plus
  `ListMetrics(ctx, ListMetricsOptions)` — a dispatcher that routes by
  `OwnerType` (`naics` | `psc` | `entity`). Constants
  `MetricsOwnerNAICS`, `MetricsOwnerPSC`, `MetricsOwnerEntity`.
- `SearchOpportunityAttachments` — required `Q`, optional `TopK` +
  `IncludeExtractedText`.
- `ListAPIKeys` (non-paginated).

#### Webhooks

- `tango/webhooks` signing subpackage: `Generate`, `Verify`, `Parse`,
  plus `VerifyRequest(*http.Request, secret)` and an `http.Handler`
  `Middleware(secret, next)` helper. HMAC-SHA256, raw-body,
  constant-time comparison — wire-compatible with the JS/Python SDKs.
- Client-side webhook CRUD in `webhooks_api.go` (root `tango` package,
  distinct from the `webhooks/` signing subpackage):
  `ListWebhookEventTypes`, `ListWebhookEndpoints`, `GetWebhookEndpoint`,
  `CreateWebhookEndpoint`, `UpdateWebhookEndpoint`,
  `DeleteWebhookEndpoint`, `TestWebhookEndpoint`,
  `GetWebhookSamplePayload`, `ListWebhookAlerts`, `GetWebhookAlert`,
  `CreateWebhookAlert`, `UpdateWebhookAlert`, `DeleteWebhookAlert`.
  Client-side validation rejects empty `Name`/`CallbackURL` on
  `CreateWebhookEndpoint` and empty `Filters` on `CreateWebhookAlert`,
  matching the `tango-node` 1.0.0 behavior.
