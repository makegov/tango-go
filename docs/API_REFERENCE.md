# API Reference

Method-by-method reference for every public function and method on `*tango.Client` plus the supporting types. ~94 methods in total. For client construction options, see [`CLIENT.md`](CLIENT.md); for response shaping, see [`SHAPES.md`](SHAPES.md); for webhook signing + receiving, see [`WEBHOOKS.md`](WEBHOOKS.md).

All methods take `context.Context` first. Options structs are always passed by pointer; `nil` is valid and means "use SDK / server defaults". List methods return `*PaginatedResponse[Record]` (where `Record = map[string]any`); a handful of typed-return methods are flagged inline. Detail methods return `Record` or a typed `*<Resource>Record` struct (also flagged).

```go
import "github.com/makegov/tango-go"

client := tango.NewClient(tango.WithAPIKey(os.Getenv("TANGO_API_KEY")))
```

## Contents

- [Agencies](#agencies)
- [Organizations / Offices / Departments](#organizations--offices--departments)
- [Business types](#business-types)
- [Contracts](#contracts)
- [IDVs](#idvs) (+ sub-resources)
- [OTAs / OTIDVs](#otas--otidvs)
- [Subawards](#subawards)
- [Vehicles](#vehicles) (+ sub-resources)
- [Entities](#entities) (+ sub-resources)
- [Opportunities / Notices / Forecasts / Grants](#opportunities--notices--forecasts--grants)
- [Protests](#protests)
- [IT Dashboard](#it-dashboard)
- [GSA eLibrary](#gsa-elibrary)
- [LCATs](#lcats)
- [Metrics](#metrics)
- [Lookups (NAICS / PSC / MAS SINs / Assistance Listings)](#lookups)
- [Resolve / Validate](#resolve--validate)
- [Webhooks](#webhooks)
- [Meta (Version / API keys)](#meta)

---

## Agencies

### `ListAgencies(ctx, *ListAgenciesOptions) (*PaginatedResponse[Record], error)`

`GET /api/agencies/`. List federal departments and subagencies.

```go
page, err := client.ListAgencies(ctx, &tango.ListAgenciesOptions{
    Page: 1, Limit: 25, Search: "Defense",
})
```

`ListAgenciesOptions` is intentionally minimal: `Page`, `Limit` (max 100; the SDK caps), `Search`.

### `GetAgency(ctx, code string) (*AgencyRecord, error)`

`GET /api/agencies/{code}/`. Fetch a single agency by its CGAC code (e.g. `"9700"` for Defense, `"2000"` for Treasury).

> **Typed return.** Returns `*AgencyRecord`, not `Record`. Pointer fields (`AgencyID`, `Name`, `Abbreviation`, `Code`) distinguish "absent" from "empty"; `Extra map[string]any` preserves forward-compatible fields.

```go
agency, err := client.GetAgency(ctx, "9700")
if agency.Name != nil {
    fmt.Println(*agency.Name)
}
```

### `ListAgencyAwardingContracts(ctx, code string, *AgencyContractsOptions) (*PaginatedResponse[Record], error)`

`GET /api/agencies/{code}/contracts/awarding/`. List contracts where the given agency is the **awarding** agency.

### `ListAgencyFundingContracts(ctx, code string, *AgencyContractsOptions) (*PaginatedResponse[Record], error)`

`GET /api/agencies/{code}/contracts/funding/`. List contracts where the given agency is the **funding** agency.

Both methods accept `AgencyContractsOptions`: embeds `ListOptions` plus `Joiner` (for `Flat: true`), `Ordering`, `Search`, and `Extra` for unknown filters.

---

## Organizations / Offices / Departments

### `ListOrganizations(ctx, *ListOrganizationsOptions) (*PaginatedResponse[Record], error)`

`GET /api/organizations/`. The canonical agency/department/office hierarchy. Use this in preference to the deprecated `ListDepartments`.

Filter fields: `Search`, `Type`, `Level` (`"1"` = department, `"2"` = agency, `"3"` = sub-agency, ...), `CGAC`, `Parent`, `IncludeInactive *bool`.

```go
orgs, _ := client.ListOrganizations(ctx, &tango.ListOrganizationsOptions{
    Level:  "1",
    Search: "Defense",
})
```

### `GetOrganization(ctx, key string) (Record, error)`

`GET /api/organizations/{key}/`.

### `ListOffices(ctx, *ListOfficesOptions) (*PaginatedResponse[Record], error)`

`GET /api/offices/`. Federal contracting offices (FPDS-NG hierarchy). Filter via `Search`.

### `GetOffice(ctx, code string) (Record, error)`

`GET /api/offices/{code}/`. Code is the FPDS-NG office code.

### `ListDepartments(ctx, *ListOptions) (*PaginatedResponse[Record], error)`

> **Deprecated upstream.** `GET /api/departments/`. Retained for parity. Prefer `ListOrganizations` with `Level: "1"`.

### `GetDepartment(ctx, code string) (Record, error)`

`GET /api/departments/{code}/`. Code is typically the CGAC department code.

---

## Business types

### `ListBusinessTypes(ctx, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/business_types/`. SBA / SAM.gov socioeconomic and structural designations (8(a), woman-owned, veteran-owned, non-profit, etc.).

### `GetBusinessType(ctx, code string) (Record, error)`

`GET /api/business_types/{code}/`. Returns `*NotFoundError` when the code is unknown.

---

## Contracts

### `ListContracts(ctx, *ListContractsOptions) (*PaginatedResponse[Record], error)`

`GET /api/contracts/`. Search and list federal contract records.

```go
page, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions:    tango.ListOptions{Shape: tango.ShapeContractsMinimal, Limit: 25},
    AwardingAgency: "9700",
    FiscalYear:     "2025",
    Keyword:        "cloud services",
    Sort:           "award_date",
    Order:          "desc",
})
```

**Filter aliases.** `ListContractsOptions` mirrors the Node and Python SDKs' SDK-friendly aliases:

| SDK-friendly field | Wire param | Notes |
| ------------------ | ---------- | ----- |
| `Keyword` | `search` | |
| `NAICSCode` | `naics` | also `NAICS` accepted |
| `PSCCode` | `psc` | also `PSC` accepted |
| `RecipientName` | `recipient` | also `Recipient` accepted |
| `RecipientUEI` | `uei` | also `UEI` accepted |
| `SetAsideType` | `set_aside` | also `SetAside` accepted |

When both an alias and a canonical field are set, the alias wins (mirrors Node).

**Sorting.** Two ways:

```go
opts.Ordering = "-award_date"        // wire format
// or
opts.Sort = "award_date"
opts.Order = "desc"                  // "asc" (default) or "desc"
```

**Pagination.** Both `?page=` and `?cursor=` are supported. Set `Cursor` for deep pagination; set `Page` for shallow. They're mutually exclusive — `Cursor` wins if both are set.

**Date / FY / dollar fields** are all strings on the wire (e.g. `"2024-01-01"`, `"2024"`). See the godoc on `ListContractsOptions` for the full filter set.

### `IterateContracts(ctx, *ListContractsOptions) *Iterator[Record]`

Walks every contract matching opts. Auto-follows `?page=` or `?cursor=` based on the server's `next` URL.

```go
for c, err := range client.IterateContracts(ctx, opts).Seq() {
    if err != nil { return err }
    fmt.Println(c["piid"])
}
```

---

## IDVs

IDVs (indefinite delivery vehicles) are parent "vehicle award" records that can have child awards/orders under them.

### `ListIDVs(ctx, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/`. Cursor-paginated.

### `GetIDV(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/idvs/{key}/`. Pass `&GetEntityOptions{Shape: tango.ShapeIDVsComprehensive}` for a full-fidelity envelope.

### `IterateIDVs(ctx, *ListIDVsOptions) *Iterator[Record]`

### `ListIDVAwards(ctx, key string, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/awards/`. Lists task-order child awards under a parent IDV. Re-uses `ListIDVsOptions` for filter fidelity.

### `ListIDVChildIDVs(ctx, key string, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/idvs/`. Child IDVs nested under a parent IDV.

### `ListIDVTransactions(ctx, key string, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/transactions/`. Raw transaction history backing an IDV. Only accepts pagination params (no filters).

### `GetIDVSummary(ctx, identifier string) (Record, error)`

> **Deprecated.** `GET /api/idvs/{identifier}/summary/`. The current server returns `404` for this endpoint. Retained for parity with the Node SDK. Migrate to `GetIDV` with a richer `Shape`.

### `ListIDVSummaryAwards(ctx, identifier string, *ListOptions) (*PaginatedResponse[Record], error)`

> **Deprecated.** `GET /api/idvs/{identifier}/summary/awards/`. Server returns `404`. Migrate to `ListIDVAwards`.

### `ListIDVLcats(ctx, key string, *EntityLcatsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/lcats/`. Labor Categories attached to an IDV. Re-uses `EntityLcatsOptions` because the entity and IDV lcats endpoints share a parameter shape.

---

## OTAs / OTIDVs

OTAs (Other Transaction Authority awards) and OTIDVs (umbrella OT agreements with child awards) are FAR-exempt awards used by DoD and others for prototype + research work.

### `ListOTAs(ctx, *ListOTAsOptions) (*PaginatedResponse[Record], error)`

`GET /api/otas/`. Cursor-paginated. Filters: `AwardingAgency`, `FundingAgency`, `PIID`, `Recipient`, `UEI`, `FiscalYear[Gte/Lte]`, `AwardDate[Gte/Lte]`, `ExpiringGte/Lte`, `PopStartDate[Gte/Lte]`, `PopEndDate[Gte/Lte]`, `PSC`, `Search`, `Ordering`, `Joiner` (for `Flat`).

### `GetOTA(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/otas/{key}/`.

### `IterateOTAs(ctx, *ListOTAsOptions) *Iterator[Record]`

### `ListOTIDVs(ctx, *ListOTIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/otidvs/`. Same filter set as `ListOTAsOptions`.

### `GetOTIDV(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/otidvs/{key}/`.

### `IterateOTIDVs(ctx, *ListOTIDVsOptions) *Iterator[Record]`

### `ListOTIDVAwards(ctx, key string, *ListOTIDVAwardsOptions) (*PaginatedResponse[Record], error)`

`GET /api/otidvs/{key}/awards/`. Child awards under an OTIDV parent. Same filter set as `ListOTAsOptions`.

### `IterateOTIDVAwards(ctx, key string, *ListOTIDVAwardsOptions) *Iterator[Record]`

---

## Subawards

### `ListSubawards(ctx, *ListSubawardsOptions) (*PaginatedResponse[Record], error)`

`GET /api/subawards/`. Filters: `AwardKey`, `PrimeUEI`, `SubUEI`, `AwardingAgency`, `FundingAgency`, `FiscalYear[Gte/Lte]`, `Recipient`, `Ordering`.

> **Ordering allowlist.** The server rejects all ordering values except `"last_modified_date"` and `"-last_modified_date"`. Other values return `400` (tango#2254).

> **Shape constraints.** Use `ShapeSubawardsMinimal` — the server rejects `id` and `amount` in subaward shapes.

---

## Vehicles

Vehicles provide a solicitation-centric grouping of related IDVs.

### `ListVehicles(ctx, *ListVehiclesOptions) (*PaginatedResponse[Record], error)`

`GET /api/vehicles/`. Filters: `Search` (full-text), `VehicleType`, `TypeOfIDC`, `ContractType`, `SetAside`, `WhoCanUse`, `NAICSCode`, `PSCCode`, `ProgramAcronym`, `Agency`, `OrganizationID`, dollar/count bounds, fiscal year, award/last-date-to-order date ranges, `Ordering`, `Joiner`.

> **Ordering allowlist.** Server enforces a strict allowlist; other values return `400`.

### `GetVehicle(ctx, uuid string, *GetEntityOptions) (Record, error)`

`GET /api/vehicles/{uuid}/`. On the detail endpoint, `search` filters expanded `awardees(...)` when included in your shape (it does not filter the vehicle itself).

### `IterateVehicles(ctx, *ListVehiclesOptions) *Iterator[Record]`

### `ListVehicleAwardees(ctx, uuid string, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/vehicles/{uuid}/awardees/`. The entities holding child IDVs under a vehicle. Use `ShapeVehicleAwardeesMinimal` for the common preset.

### `ListVehicleOrders(ctx, uuid string, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/vehicles/{uuid}/orders/`. Task orders placed under a vehicle's child IDVs.

> Python-only method on the sibling SDKs — included here for full parity.

---

## Entities

### `ListEntities(ctx, *ListEntitiesOptions) (*PaginatedResponse[Record], error)`

`GET /api/entities/`. Federal vendors / recipients. Filters: `Search`, `CageCode`, `NAICS`, `Name`, `PSC`, `PurposeOfRegistrationCode`, `Socioeconomic`, `State`, `TotalAwardsObligated[Gte/Lte]`, `UEI`, `ZipCode`.

### `GetEntity(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/entities/{key}/`. Key is the UEI or CAGE code. When `Shape` is empty, the server returns its comprehensive default; pass `ShapeEntitiesMinimal` for a slimmer payload.

### `IterateEntities(ctx, *ListEntitiesOptions) *Iterator[Record]`

### Entity sub-resources

All take a UEI plus `*EntitySubresourceOptions` (embeds `ListOptions` + `Joiner` + `Ordering` + `Search` + `Extra`), except `ListEntitySubawards` (uses `EntitySubawardsOptions`) and `ListEntityLcats` (uses `EntityLcatsOptions`).

| Method | Endpoint |
| ------ | -------- |
| `ListEntityContracts(ctx, uei, *EntitySubresourceOptions)` | `GET /api/entities/{uei}/contracts/` |
| `ListEntityIDVs(ctx, uei, *EntitySubresourceOptions)` | `GET /api/entities/{uei}/idvs/` |
| `ListEntityOTAs(ctx, uei, *EntitySubresourceOptions)` | `GET /api/entities/{uei}/otas/` |
| `ListEntityOTIDVs(ctx, uei, *EntitySubresourceOptions)` | `GET /api/entities/{uei}/otidvs/` |
| `ListEntitySubawards(ctx, uei, *EntitySubawardsOptions)` | `GET /api/entities/{uei}/subawards/` |
| `ListEntityLcats(ctx, uei, *EntityLcatsOptions)` | `GET /api/entities/{uei}/lcats/` |

All return `*PaginatedResponse[Record]`. Empty UEI is rejected client-side as `*ValidationError`.

### `GetEntityMetrics(ctx, uei string, months int, periodGrouping string) (Record, error)`

`GET /api/entities/{uei}/metrics/{months}/{periodGrouping}/`. See [Metrics](#metrics) below.

---

## Opportunities / Notices / Forecasts / Grants

### `ListOpportunities(ctx, *ListOpportunitiesOptions) (*PaginatedResponse[Record], error)`

`GET /api/opportunities/`. SAM.gov opportunities. Filters: `Active *bool`, `Agency`, `FirstNoticeDate[After/Before]`, `LastNoticeDate[After/Before]`, `NAICS`, `NoticeType`, `Ordering`, `PlaceOfPerformance`, `PSC`, `ResponseDeadline[After/Before]`, `Search`, `SetAside`, `SolicitationNumber`.

### `IterateOpportunities(ctx, *ListOpportunitiesOptions) *Iterator[Record]`

### `SearchOpportunityAttachments(ctx, SearchOpportunityAttachmentsOptions) (Record, error)`

`GET /api/opportunities/attachment-search/`. Semantic search over the extracted text of opportunity attachments (SOWs, PWSs, J&As, etc.).

```go
res, err := client.SearchOpportunityAttachments(ctx, tango.SearchOpportunityAttachmentsOptions{
    Q:                    "cybersecurity zero trust",
    TopK:                 10,
    IncludeExtractedText: false,
})
```

`Q` is required; empty Q raises `*ValidationError` before any network call. `TopK: 0` means "use the server default".

### `ListNotices(ctx, *ListNoticesOptions) (*PaginatedResponse[Record], error)`

`GET /api/notices/`. Filters: `Active *bool`, `Agency`, `NAICS`, `NoticeType`, `PostedDate[After/Before]`, `PSC`, `ResponseDeadline[After/Before]`, `Search`, `SetAside`, `SolicitationNumber`.

> **No ordering.** The notices viewset rejects every `?ordering=` value, so `ListNoticesOptions` deliberately omits an `Ordering` field (mirrors Python and Node).

### `IterateNotices(ctx, *ListNoticesOptions) *Iterator[Record]`

### `ListForecasts(ctx, *ListForecastsOptions) (*PaginatedResponse[Record], error)`

`GET /api/forecasts/`. Filters: `Agency`, `AwardDate[After/Before]`, `FiscalYear[Gte/Lte]`, `Modified[After/Before]`, `NAICSCode`, `NAICSStartsWith`, `Ordering`, `Search`, `SourceSystem`, `Status`.

### `IterateForecasts(ctx, *ListForecastsOptions) *Iterator[Record]`

### `ListGrants(ctx, *ListGrantsOptions) (*PaginatedResponse[Record], error)`

`GET /api/grants/`. Filters: `Agency`, `ApplicantTypes`, `CFDANumber`, `FundingCategories`, `FundingInstruments`, `OpportunityNumber`, `Ordering`, `PostedDate[After/Before]`, `ResponseDate[After/Before]`, `Search`, `Status`.

### `IterateGrants(ctx, *ListGrantsOptions) *Iterator[Record]`

---

## Protests

### `ListProtests(ctx, *ListProtestsOptions) (*PaginatedResponse[Record], error)`

`GET /api/protests/`. Bid protests from GAO + COFC. Filters: `SourceSystem`, `Outcome`, `CaseType`, `Agency`, `CaseNumber`, `SolicitationNumber`, `Protester`, `Search`, `FiledDate[After/Before]`, `DecisionDate[After/Before]`.

> **No ordering.** The viewset rejects ordering; `ListProtestsOptions` deliberately omits the field.

### `IterateProtests(ctx, *ListProtestsOptions) *Iterator[Record]`

### `GetProtest(ctx, caseNumber string, *GetEntityOptions) (*ProtestRecord, error)`

`GET /api/protests/{caseNumber}/`.

> **Typed return.** Returns `*ProtestRecord` with named fields (`CaseID`, `CaseNumber`, `SourceSystem`, `Outcome`, `CaseType`, `FiledDate`, `DecisionDate`, `Agency`, `Protester`, `ResolvedAgency`, `ResolvedProtester`, `Docket []map[string]any`, `Extra map[string]any`).

Use `shape: "...,docket(...)"` to include the nested docket entries.

---

## IT Dashboard

### `ListItDashboard(ctx, *ListItDashboardOptions) (*PaginatedResponse[Record], error)`

`GET /api/itdashboard/`. Federal IT investments. **Filters are tier-gated** by the API:

- Free: `Search`
- Pro: `AgencyCode`, `TypeOfInvestment`, `UpdatedTime[After/Before]`
- Business+: `AgencyName`, `CIORating`, `CIORatingMax`, `PerformanceRisk`

Hitting a gated filter on a lower tier returns `403` (surfaces as `*APIError` with `StatusCode 403`).

CIO ratings: `1` = High Risk, `2` = Moderately High, `3` = Medium, `4` = Moderately Low, `5` = Low.

### `IterateItDashboard(ctx, *ListItDashboardOptions) *Iterator[Record]`

### `GetItDashboard(ctx, uii string, *GetEntityOptions) (Record, error)`

`GET /api/itdashboard/{uii}/`. UII is the Unique Investment Identifier.

---

## GSA eLibrary

### `ListGsaElibraryContracts(ctx, *ListGsaElibraryContractsOptions) (*PaginatedResponse[Record], error)`

`GET /api/gsa_elibrary_contracts/`. Filters: `Schedule`, `ContractNumber`, `Key`, `PIID`, `UEI`, `SIN`, `Search`, `Ordering`.

### `IterateGsaElibraryContracts(ctx, *ListGsaElibraryContractsOptions) *Iterator[Record]`

### `GetGsaElibraryContract(ctx, uuid string, *GetEntityOptions) (Record, error)`

`GET /api/gsa_elibrary_contracts/{uuid}/`. Python-only on the sibling SDKs; included here for parity.

---

## LCATs

### `ListLcats(ctx, *ListLcatsOptions) (*PaginatedResponse[Record], error)`

Labor Categories. LCATs live under owner resources in the Tango API — there is no top-level `/api/lcats/` endpoint. `ListLcats` dispatches based on which field on `*ListLcatsOptions` is set:

- `UEI` set → `GET /api/entities/{uei}/lcats/`
- `IDVKey` set → `GET /api/idvs/{key}/lcats/`
- Both set → UEI wins (mirrors the Node SDK)
- Neither set → returns `*ValidationError`

The embedded `EntityLcatsOptions` carries `Ordering`, `Search`, plus the standard `ListOptions` (Page / Limit / Cursor / Shape / Flat / FlatLists).

### `IterateLcats(ctx, *ListLcatsOptions) *Iterator[Record]`

Same dispatch rules as `ListLcats`; the iterator walks pages of the dispatched endpoint.

See also: `ListEntityLcats(ctx, uei, *EntityLcatsOptions)` and `ListIDVLcats(ctx, key, *EntityLcatsOptions)` for direct sub-resource calls when you already know the owner type.

---

## Metrics

Rolling-window metrics aggregated by NAICS, PSC, or entity. All three concrete getters take `(code, months, periodGrouping)` where `months > 0` and `periodGrouping` is typically `"month"`, `"quarter"`, or `"year"`.

### `GetNAICSMetrics(ctx, code string, months int, periodGrouping string) (Record, error)`

`GET /api/naics/{code}/metrics/{months}/{periodGrouping}/`.

### `GetPSCMetrics(ctx, code string, months int, periodGrouping string) (Record, error)`

`GET /api/psc/{code}/metrics/{months}/{periodGrouping}/`.

### `GetEntityMetrics(ctx, uei string, months int, periodGrouping string) (Record, error)`

`GET /api/entities/{uei}/metrics/{months}/{periodGrouping}/`.

### `ListMetrics(ctx, ListMetricsOptions) (Record, error)`

Convenience dispatcher that routes to one of the three above based on `opts.OwnerType` (`tango.MetricsOwnerNAICS` / `MetricsOwnerPSC` / `MetricsOwnerEntity`).

```go
m, err := client.ListMetrics(ctx, tango.ListMetricsOptions{
    OwnerType:      tango.MetricsOwnerNAICS,
    OwnerID:        "541511",
    Months:         12,
    PeriodGrouping: "month",
})
```

Empty `OwnerID`, non-positive `Months`, empty `PeriodGrouping`, or unknown `OwnerType` all return `*ValidationError` client-side.

---

## Lookups

### `ListNAICS(ctx, *ListNAICSOptions) (*PaginatedResponse[Record], error)`

`GET /api/naics/`. Filters: `Search`, `RevenueLimit[Gte/Lte]`, `EmployeeLimit[Gte/Lte]`.

### `GetNAICS(ctx, code string) (Record, error)`

`GET /api/naics/{code}/`.

### `ListPSC(ctx, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/psc/`.

### `GetPSC(ctx, code string) (Record, error)`

`GET /api/psc/{code}/`.

### `ListMasSins(ctx, *ListMasSinsOptions) (*PaginatedResponse[Record], error)`

`GET /api/mas_sins/`. MAS (Multiple Award Schedule) Special Item Numbers.

### `GetMasSin(ctx, sin string) (Record, error)`

`GET /api/mas_sins/{sin}/`.

### `ListAssistanceListings(ctx, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/assistance_listings/`. The CFDA (Catalog of Federal Domestic Assistance) program catalog.

### `GetAssistanceListing(ctx, number string) (Record, error)`

`GET /api/assistance_listings/{number}/`. Number is the CFDA number (e.g. `"10.001"`).

---

## Resolve / Validate

### `Resolve(ctx, ResolveInput) (*ResolveResult, error)`

`POST /api/resolve/`. Fuzzy-match a free-text name to ranked entity or organization candidates.

```go
result, err := client.Resolve(ctx, tango.ResolveInput{
    Name:       "Lockheed Martin",
    TargetType: tango.ResolveEntity,    // or tango.ResolveOrganization
    State:      "MD",                    // optional disambiguator
})
for _, c := range result.Candidates {
    fmt.Println(c.Identifier, c.DisplayName, c.MatchTier)
}
```

> **Typed return.** `*ResolveResult` (with `Count int`, `Candidates []ResolveCandidate`). Each candidate has `Identifier`, `DisplayName`, `MatchTier` (Pro+ only — Free responses omit this), and `Extra map[string]any` for forward-compatible fields.

Required fields: `Name`, `TargetType` (`"entity"` | `"organization"`). Both validated client-side.

### `Validate(ctx, ValidateInput) (*ValidateResult, error)`

`POST /api/validate/`. Validate the format of an identifier.

```go
result, err := client.Validate(ctx, tango.ValidateInput{
    Type:  tango.ValidateUEI,           // or ValidatePIID / ValidateSolicitation
    Value: "ABCDEF123456",
})
fmt.Println(result.Result)              // "valid" | "invalid" | "low_confidence"
```

`Value` is required client-side. `Result.Errors` carries structured failures when the result is non-valid.

---

## Webhooks

See [`WEBHOOKS.md`](WEBHOOKS.md) for the full guide. Quick reference:

### Endpoints

| Method | Endpoint |
| ------ | -------- |
| `ListWebhookEventTypes(ctx)` → `*WebhookEventTypesResponse` | `GET /api/webhooks/event-types/` |
| `ListWebhookEndpoints(ctx, *ListOptions)` → `*PaginatedResponse[WebhookEndpoint]` | `GET /api/webhooks/endpoints/` |
| `GetWebhookEndpoint(ctx, id)` → `*WebhookEndpoint` | `GET /api/webhooks/endpoints/{id}/` |
| `CreateWebhookEndpoint(ctx, WebhookEndpointCreateInput)` → `*WebhookEndpoint` | `POST /api/webhooks/endpoints/` |
| `UpdateWebhookEndpoint(ctx, id, WebhookEndpointUpdateInput)` → `*WebhookEndpoint` | `PATCH /api/webhooks/endpoints/{id}/` |
| `DeleteWebhookEndpoint(ctx, id) error` | `DELETE /api/webhooks/endpoints/{id}/` |
| `TestWebhookEndpoint(ctx, endpointID)` → `*WebhookTestDeliveryResult` | `POST /api/webhooks/endpoints/test-delivery/` |
| `GetWebhookSamplePayload(ctx, eventType)` → `*WebhookSamplePayloadResponse` | `GET /api/webhooks/endpoints/sample-payload/?event_type={eventType}` |

`CreateWebhookEndpoint` validates `Name` + `CallbackURL` client-side (both required; `Name` is unique per user server-side).

### Alerts

| Method | Endpoint |
| ------ | -------- |
| `ListWebhookAlerts(ctx, *ListOptions)` → `*PaginatedResponse[WebhookAlert]` | `GET /api/webhooks/alerts/` |
| `GetWebhookAlert(ctx, id)` → `*WebhookAlert` | `GET /api/webhooks/alerts/{id}/` |
| `CreateWebhookAlert(ctx, WebhookAlertCreateInput)` → `*WebhookAlert` | `POST /api/webhooks/alerts/` |
| `UpdateWebhookAlert(ctx, id, WebhookAlertUpdateInput)` → `*WebhookAlert` | `PATCH /api/webhooks/alerts/{id}/` |
| `DeleteWebhookAlert(ctx, id) error` | `DELETE /api/webhooks/alerts/{id}/` |

`CreateWebhookAlert` validates `Name`, `QueryType`, non-empty `Filters` client-side. **`QueryType` is singular** (`"contract"`, not `"contracts"`).

---

## Meta

### `GetVersion(ctx) (Record, error)`

`GET /api/version/`. Returns the server's version metadata (build commit, deployed-at, etc.).

### `ListAPIKeys(ctx) (Record, error)`

`GET /api/api-keys/`. Returns the authenticated user's API keys. Non-paginated — returns a structured `Record` with the caller's keys and metadata.

---

## Client lifecycle + observability

These aren't resource methods but are part of the public surface — useful when wiring up logging, metrics, or per-environment configuration. See [`CLIENT.md`](CLIENT.md) for full details.

| Method | Returns | Purpose |
| ------ | ------- | ------- |
| `Client.BaseURL()` | `string` | The resolved base URL the client is hitting. |
| `Client.RateLimitInfo()` | `*RateLimitInfo` | Snapshot of the rate-limit headers from the most recent response. `nil` before any request. |
| `Client.LastResponseHeaders()` | `http.Header` | Headers from the most recent response (X-Request-Id, etc.). `nil` before any request. |

---

## Notes on conventions

- **`context.Context` is always first.** Even for methods that have no other knobs (`GetVersion`, `ListAPIKeys`), `ctx` is the first arg.
- **Options structs are pointers.** Passing `nil` is valid for any `*Xxx` opts argument and means "use SDK / server defaults". The SDK never panics on a nil opts.
- **Required path segments are validated client-side.** Empty `uei`, `key`, `code`, `id`, etc. return `*ValidationError` with `StatusCode: 0` before any network call.
- **Date fields are strings.** Wire format is `YYYY-MM-DD` for dates, ISO 8601 with timezone for timestamps, integer-as-string for fiscal years (e.g. `"2024"`). No `time.Time` parsing layer in the SDK.
- **`Record` is `map[string]any`.** Use it as you would any `map`; serialize with `encoding/json`; cast to your own struct via `json.Marshal` → `json.Unmarshal` when you want field safety.
- **The full method count is ~94.** This page lists every one. If you find a sibling-SDK method that isn't here, file an issue.
