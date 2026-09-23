<!-- markdownlint-disable MD024, MD013 -->
# Changelog

All notable changes to `github.com/makegov/tango-go` will be documented in this file.

This project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.0] - 2026-09-23

Pre-1.0 (SemVer 0.x): the breaking changes below ship without a deprecation cycle, and each is marked **Breaking**.

### Added

- **Budget accounts** (`budget.go`): `ListBudgetAccounts` / `IterateBudgetAccounts` (`GET /api/budget/accounts/`), `GetBudgetAccount`, `GetBudgetAccountQuarters` and `GetBudgetAccountRecipients`, with `ListBudgetAccountsOptions` (including `BureauName`, `AccountTitleContains` and `SubfunctionCode`), `BudgetAccountQuartersOptions` (`TAS`), `BudgetAccountRecipientsOptions` (`FundingOrganizationID`) and the `ShapeBudgetAccountsMinimal` preset. The metric range filters (for example `enacted_ba__gte`) go through `Extra`.
- **`GetEntityBudgetFlows`** (`GET /api/entities/{uei}/budget-flows/`) with `EntityBudgetFlowsOptions` (`Page`, `Limit`, `FiscalYear`): the federal accounts that paid an entity, contract flows only.
- **Singleton detail GETs**: `GetContract`, `GetSubaward`, `GetOpportunity`, `GetNotice`, `GetForecast` and `GetGrant`, each taking `*GetEntityOptions` for shaping.
- **Contract sub-routes**: `ListContractSubawards` (`GET /api/contracts/{key}/subawards/`, taking `*EntitySubawardsOptions`) and `ListContractTransactions` (`GET /api/contracts/{key}/transactions/`, taking `*ListOptions`).
- **Exclusions** (parity with tango-python and tango-node): `ListExclusions` / `IterateExclusions` / `GetExclusion` over `/api/exclusions/`, with every filter the endpoint accepts on `ListExclusionsOptions`. `Active` is derived at query time, so an exclusion reaching its termination date changes it without firing an alert.
- **DIBBS** (parity with tango-python and tango-node): `ListDibbsRfqs` / `ListDibbsRfps` / `ListDibbsAwards`, their `Iterate…` and `Get…` counterparts, and one options struct per resource naming every filter. Award rows are order line items, so the order-level `total_contract_price` does not sum across rows.
- **SBIR/STTR** (parity with tango-python and tango-node): `ListSbirTopics` / `ListSbirSolicitations`, their `Iterate…` and `Get…` counterparts, and options structs naming every filter.
- **New typed filters on existing lists**: `Key` on `ListContractsOptions`, `ListIDVsOptions`, `ListOTAsOptions`, `ListOTIDVsOptions` and `ListOTIDVAwardsOptions`; `OpportunityID` on `ListOpportunitiesOptions`; `NoticeID`, `Department` and `Office` on `ListNoticesOptions`; `ID` on `ListForecastsOptions`; `GrantID` on `ListGrantsOptions`; `PreviousUII` on `ListItDashboardOptions`; `Verbose` on `ListSledOpportunitiesOptions`; `Cage` on `ListEntitiesOptions` (the API's alias for `cage_code`; the server rejects a request that sets both).
- **Subaward filters on `EntitySubawardsOptions`**: `AwardKey`, `PrimeUEI`, `SubUEI`, `AwardingAgency`, `FundingAgency`, `FiscalYear[Gte/Lte]` and `Recipient`, used by both `ListEntitySubawards` and `ListContractSubawards`.
- **`DepartmentRecord`**, the typed return of `GetDepartment` (see Changed).
- **Contract Disputes Act appeal decisions** (Tango API 4.26.0). `ListContractAppeals` / `GetContractAppeal` / `IterateContractAppeals` cover `/api/contract_appeals/` — decisions from the Civilian Board of Contract Appeals (CBCA) and the Armed Services Board of Contract Appeals (ASBCA). `ListContractAppealsOptions` names every filter the endpoint accepts (`Board`, `Docket`, `Appellant`, `Judge`, `DecisionType`, `DecisionDate[After/Before]`, `Listed`, `DocumentID`, `Search`, `Ordering`), and two `Shape*` presets land in `shapes.go`.

  **These are not bid protests.** An appeal disputes a contracting officer's final decision under a contract the government already awarded; a protest challenges the award itself and stays on `ListProtests`. The two resources share no identifiers and no vocabulary, so neither one's filters mean anything against the other's rows.

  `GetContractAppeal` returns a typed `*ContractAppealRecord`, as `GetProtest` does for a protest. Every field on it is a pointer, because an unshaped list row carries only a core subset of the columns and **`decision_text` is served on the Enterprise plan only, absent rather than null below it** — so `nil` has to keep meaning "not served to you" rather than collapsing into an empty string. `TextStatus` and `TextCharCount` describe the extracted text at every plan and read as a pair: a status claiming text alongside a zero character count is a document that has not yielded any.

  Neither `ShapeContractAppealsMinimal` nor `ShapeContractAppealsComprehensive` names `decision_text`, so no suggested shape makes a detail fetch ask for a body most callers cannot read; `TestContractAppealShapesDoNotNameTheEnterpriseDecisionText` pins that. `Listed` is a `*bool` rather than a `bool`, so `false` reaches the server as a filter value instead of vanishing into the zero value.

- **`attachments(extracted_text)` — SLED document bodies on the Small plan and above** (Tango API 4.25.1; parity with tango-python and tango-node). This SDK returns `Record`, so the leaf needs no schema change — what it needed was saying so. Documented on `GetSledOpportunity`, on `ShapeSledOpportunitiesComprehensive` and in `docs/API_REFERENCE.md`: the leaf must be **named** (no `Shape*` preset includes it, and `attachments(*)` does not carry it, because the API resolves the body only for a caller who asked); the **key is absent rather than null** when the text is not being served; and a **contested document never returns text at any plan**. `TestSledShapesDoNotNameThePaidDocumentBody` pins the presets. Searching document text stays ungated on every plan and returns no fragment of it.

- **State, local and education (SLED) procurement** (Tango API 4.25.0; parity with tango-python and tango-node). `ListSledOpportunities` / `GetSledOpportunity`, `ListSledOpportunityRevisions`, `GetSledCoverage`, `ListSledForecasts` / `GetSledForecast`, plus `IterateSledOpportunities` and `IterateSledForecasts`. Three options structs cover every one of the API's 27 solicitation filters and 13 forecast filters as a named field, and five `Shape*` presets land in `shapes.go`.

  Four behaviors are documented on the options structs because each misleads a caller who assumes federal semantics. **Leaving both `Status` and `Active` unset returns open solicitations only** — that default is the API's, and `ListSledOpportunities` deliberately does not synthesize `status=open`, since doing so would make `Active: boolPtr(false)` unreachable (pinned by a test). **`Status` is Tango-derived and refreshed every fifteen minutes**; the portal's own word is served as `source_status`, is frozen at last capture, and is not a liveness filter. **Category scheme tagging is mid-migration**, so `Naics` matches only the small tagged share and `CategoryCode` is the escape hatch. And **`meta.attachment_count` can be lower than `len(attachments)`**, because an auto-generated portal cover sheet is listed and flagged `is_generated_summary` but excluded from the count.

  `Active`, `HasDocuments` and `SourceDeclared` are `*bool` rather than `bool`, so `false` reaches the server as a filter value instead of vanishing into the zero value — the same reason `setIfNotNilBool` exists.

  `ShapeSledRevisionsMinimal` omits `changes` on purpose: the per-field before/after needs a Small plan, so naming it in a suggested shape would 403 a Free caller. `changed_fields` is in the shape and available at every plan.

- **`ListProtestsOptions.NaicsCode`** filters protests by the NAICS code at issue in an SBA OHA size or NAICS appeal. GAO and COFC cases never match it.
- **`ShapeSledOpportunitiesMinimal` and `ShapeSledOpportunitiesComprehensive` now include `delisted_at`**, matching the API's own default SLED shapes, so a caller can see when a portal stopped listing a solicitation.

### Changed

- **Breaking: `ProtestRecord` now matches the fields the API returns.** `Agency` and `Protester` are `string` (they were `map[string]any`), and the docket entries moved from `Docket` (tagged `docket`) to `Dockets` (tagged `dockets`). Code that reads `rec.Agency["..."]`, `rec.Protester["..."]` or `rec.Docket` must be updated. `ProtestRecord` also gains every other field the API serves: `Title`, `SolicitationNumber`, `PostedDate`, `DueDate`, `DocketURL`, `DecisionURL`, `Organization`, `Decisions`, and the opt-in `ChallengedParty`, `NaicsCode`, `SizeStandard`, `OutcomeReason`, `Judge`, `Digest` and `DecisionText`.
- **Breaking: `GetDepartment` returns `*DepartmentRecord`** instead of `Record`. The API serves a department's `code` as an integer (the Department of Defense is `97`), so `DepartmentRecord.Code` is an `*int`. Code that read `rec["code"]` should read `*rec.Code`.

### Removed

- **Breaking: `GetIDVSummary` and `ListIDVSummaryAwards`.** They called `/api/idvs/{id}/summary/` and `/api/idvs/{id}/summary/awards/`, which have never existed in the Tango API. Use `GetIDV` with a richer shape and `ListIDVAwards` instead.
- **Breaking: `SearchOpportunityAttachments` and `SearchOpportunityAttachmentsOptions`.** The API retired `/api/opportunities/attachment-search/`, which now returns 404 for every query. Match attachment text with `Search` on `ListOpportunities`.

### Fixed

- **`GetProtest` failed to decode every response that included `agency` or `protester`**, which is every unshaped call, because the API returns those fields as strings. It now decodes the real payload.
- **`GetProtest` documentation** now says the route takes the case's UUID `case_id`, not a case number such as `B-423274`; look a case up by number with `ListProtests` and `CaseNumber`. The protest docs now name all three sources: GAO, the Court of Federal Claims and the SBA Office of Hearings and Appeals.
- **`ListWebhookAlerts` ignored `Limit`.** The alerts route sizes its pages with `page_size`, not `limit`, so the SDK now sends `Limit` as `page_size`.

### Documentation

- New **Contract appeals** section in `docs/API_REFERENCE.md` covering all three methods, the protest distinction, and the Enterprise gate on the decision body. Alerts on this resource use query type `contract_appeal` and deliver `alerts.contract_appeal.match`.
- New **State & Local (SLED)** section in `docs/API_REFERENCE.md` covering all six methods and both defaults that surprise people.
- `docs/WEBHOOKS.md` troubleshooting gained the date-lapse rule and its one exception. An exclusion or DIBBS solicitation reaching its date fires nothing, because open/closed is derived at query time — but `alerts.sled_opportunity.match` **does** fire on a closing, since SLED liveness is a stored column a fifteen-minute sweep writes.
- New **Exclusions**, **DIBBS** and **SBIR / STTR** sections in `docs/API_REFERENCE.md`, and corrected signatures and descriptions for the budget, contract and entity sub-resource methods.
- The README's resource table now lists every resource with a list method, and its version banner reads v0.2.0.

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
