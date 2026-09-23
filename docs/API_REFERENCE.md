# API Reference

Method-by-method reference for every public function and method on `*tango.Client` plus the supporting types. About 140 methods in total. For client construction options, see [`CLIENT.md`](CLIENT.md); for response shaping, see [`SHAPES.md`](SHAPES.md); for webhook signing + receiving, see [`WEBHOOKS.md`](WEBHOOKS.md).

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
- [Budget](#budget)
- [IDVs](#idvs) (+ sub-resources)
- [OTAs / OTIDVs](#otas--otidvs)
- [Subawards](#subawards)
- [Vehicles](#vehicles) (+ sub-resources)
- [Entities](#entities) (+ sub-resources)
- [Opportunities / Notices / Forecasts / Grants](#opportunities--notices--forecasts--grants)
- [Protests](#protests)
- [Contract appeals](#contract-appeals)
- [State \& Local (SLED)](#state--local-sled)
- [Exclusions](#exclusions)
- [DIBBS](#dibbs)
- [SBIR / STTR](#sbir--sttr)
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

### `GetDepartment(ctx, code string) (*DepartmentRecord, error)`

`GET /api/departments/{code}/`. Returns a typed `*DepartmentRecord` (`Code *int`, `Name`, `Abbreviation`, `Extra`).
The department code is an integer on the wire (the Department of Defense is `97`), so `"97"` and `"097"` both resolve.

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

### `GetContract(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/contracts/{key}/`. Fetches a single contract record. Validates `key` non-empty client-side.

### `ListContractSubawards(ctx, key string, *EntitySubawardsOptions) (*PaginatedResponse[Record], error)`

`GET /api/contracts/{key}/subawards/`. Subawards reported against one prime contract. Takes the same subaward filters as `ListEntitySubawards`.

### `ListContractTransactions(ctx, key string, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/contracts/{key}/transactions/`. The transaction history behind one contract. The route pages with `Page` and `Limit` and takes no filters.

---

## Budget

One row per federal account and fiscal year, covering the budget lifecycle from request through outlay. The record is wide and shape-driven; `ShapeBudgetAccountsMinimal` is a compact starting point. Every dollar and ratio metric also takes exact, `__gte` and `__lte` filters (for example `enacted_ba__gte`), reachable through `Extra`.

### `ListBudgetAccounts(ctx, *ListBudgetAccountsOptions) (*PaginatedResponse[Record], error)`

`GET /api/budget/accounts/`. Lists budget-account rollups. Typed filters: `FederalAccountSymbol`, `FiscalYear` (+ `FiscalYearGte` / `FiscalYearLte`), `AgencyCode`, `BEACategory`, `OnOffBudget`, `BureauName`, `AccountTitleContains`, `SubfunctionCode`, `Search`, `Ordering`.

### `IterateBudgetAccounts(ctx, *ListBudgetAccountsOptions) *Iterator[Record]`

Walks every budget-account rollup matching opts.

### `GetBudgetAccount(ctx, id string, *GetEntityOptions) (Record, error)`

`GET /api/budget/accounts/{id}/`. One account-year by its numeric id. Name `appendix(*)` in the shape to include the appendix object-class breakdown.

### `GetBudgetAccountQuarters(ctx, id string, *BudgetAccountQuartersOptions) (*PaginatedResponse[Record], error)`

`GET /api/budget/accounts/{id}/quarters/`. Quarterly TAS-grain obligation and outlay flow for an account-year: one row per (TAS, quarter). Options: `Page`, `Limit`, `TAS`. Coverage starts at FY2021; an earlier account-year returns an empty page.

### `GetBudgetAccountRecipients(ctx, id string, *BudgetAccountRecipientsOptions) (*PaginatedResponse[Record], error)`

`GET /api/budget/accounts/{id}/recipients/`. Funding-office by recipient contract flows, largest first. Options: `Page`, `Limit`, `FundingOrganizationID`. Each row carries resolved `funding_office` and `recipient` objects and a capped `contracts` list.

---

## IDVs

IDVs (indefinite delivery vehicles) are parent "vehicle award" records that can have child awards/orders under them.

### `ListIDVs(ctx, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/`. Cursor-paginated.

`ListContractsOptions`, `ListIDVsOptions`, `ListOTAsOptions`, `ListOTIDVsOptions` and `ListOTIDVAwardsOptions` all carry `Key`, which matches the award key the detail endpoint takes; join several with `|`.

### `GetIDV(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/idvs/{key}/`. Pass `&GetEntityOptions{Shape: tango.ShapeIDVsComprehensive}` for a full-fidelity envelope.

### `IterateIDVs(ctx, *ListIDVsOptions) *Iterator[Record]`

### `ListIDVAwards(ctx, key string, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/awards/`. Lists task-order child awards under a parent IDV. Re-uses `ListIDVsOptions` for filter fidelity.

### `ListIDVChildIDVs(ctx, key string, *ListIDVsOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/idvs/`. Child IDVs nested under a parent IDV.

### `ListIDVTransactions(ctx, key string, *ListOptions) (*PaginatedResponse[Record], error)`

`GET /api/idvs/{key}/transactions/`. Raw transaction history backing an IDV. Only accepts pagination params (no filters).

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

> **Ordering allowlist.** The server rejects all ordering values except `"last_modified_date"` and `"-last_modified_date"`. Other values return `400`.

> **Shape constraints.** Use `ShapeSubawardsMinimal` — the server rejects `id` and `amount` in subaward shapes.

### `GetSubaward(ctx, key string, *GetEntityOptions) (Record, error)`

`GET /api/subawards/{key}/`. A single subaward record. Validates `key` non-empty client-side.

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

`GET /api/entities/`. Federal vendors / recipients. Filters: `Search`, `CageCode`, `Cage`, `NAICS`, `Name`, `PSC`, `PurposeOfRegistrationCode`, `Socioeconomic`, `State`, `TotalAwardsObligated[Gte/Lte]`, `UEI`, `ZipCode`.

> `Cage` is the API's alias for `CageCode`; both filter the same field, and the server rejects a request that sets both.

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
| `GetEntityBudgetFlows(ctx, uei, *EntityBudgetFlowsOptions)` | `GET /api/entities/{uei}/budget-flows/` |

All return `*PaginatedResponse[Record]`. Empty UEI is rejected client-side as `*ValidationError`.

`EntitySubawardsOptions` carries the subaward filters (`AwardKey`, `PrimeUEI`, `SubUEI`, `AwardingAgency`, `FundingAgency`, `FiscalYear[Gte/Lte]`, `Recipient`) plus `Ordering`. `EntityBudgetFlowsOptions` takes `Page`, `Limit` and `FiscalYear`; budget flows are contract-side only.

### `GetEntityMetrics(ctx, uei string, months int, periodGrouping string) (Record, error)`

`GET /api/entities/{uei}/metrics/{months}/{periodGrouping}/`. See [Metrics](#metrics) below.

---

## Opportunities / Notices / Forecasts / Grants

### `ListOpportunities(ctx, *ListOpportunitiesOptions) (*PaginatedResponse[Record], error)`

`GET /api/opportunities/`. SAM.gov opportunities. Filters: `Active *bool`, `Agency`, `FirstNoticeDate[After/Before]`, `LastNoticeDate[After/Before]`, `NAICS`, `NoticeType`, `Ordering`, `PlaceOfPerformance`, `PSC`, `ResponseDeadline[After/Before]`, `Search`, `SetAside`, `SolicitationNumber`, `OpportunityID`.

### `IterateOpportunities(ctx, *ListOpportunitiesOptions) *Iterator[Record]`

### `GetOpportunity(ctx, opportunityID string, *GetEntityOptions) (Record, error)`

`GET /api/opportunities/{opportunity_id}/`. A single opportunity. Validates `opportunityID` non-empty client-side.

### `ListNotices(ctx, *ListNoticesOptions) (*PaginatedResponse[Record], error)`

`GET /api/notices/`. Filters: `Active *bool`, `Agency`, `Department`, `Office`, `NAICS`, `NoticeType`, `NoticeID`, `PostedDate[After/Before]`, `PSC`, `ResponseDeadline[After/Before]`, `Search`, `SetAside`, `SolicitationNumber`.

> **No ordering.** The notices viewset rejects every `?ordering=` value, so `ListNoticesOptions` deliberately omits an `Ordering` field (mirrors Python and Node).

### `IterateNotices(ctx, *ListNoticesOptions) *Iterator[Record]`

### `GetNotice(ctx, noticeID string, *GetEntityOptions) (Record, error)`

`GET /api/notices/{notice_id}/`. A single notice. Validates `noticeID` non-empty client-side.

### `ListForecasts(ctx, *ListForecastsOptions) (*PaginatedResponse[Record], error)`

`GET /api/forecasts/`. Filters: `Agency`, `AwardDate[After/Before]`, `FiscalYear[Gte/Lte]`, `Modified[After/Before]`, `NAICSCode`, `NAICSStartsWith`, `Ordering`, `Search`, `SourceSystem`, `Status`, `ID`.

### `IterateForecasts(ctx, *ListForecastsOptions) *Iterator[Record]`

### `GetForecast(ctx, id string, *GetEntityOptions) (Record, error)`

`GET /api/forecasts/{id}/`. A single procurement forecast. Validates `id` non-empty client-side.

### `ListGrants(ctx, *ListGrantsOptions) (*PaginatedResponse[Record], error)`

`GET /api/grants/`. Filters: `Agency`, `ApplicantTypes`, `CFDANumber`, `GrantID`, `FundingCategories`, `FundingInstruments`, `OpportunityNumber`, `Ordering`, `PostedDate[After/Before]`, `ResponseDate[After/Before]`, `Search`, `Status`.

### `IterateGrants(ctx, *ListGrantsOptions) *Iterator[Record]`

### `GetGrant(ctx, grantID string, *GetEntityOptions) (Record, error)`

`GET /api/grants/{grant_id}/`. A single grant opportunity. Validates `grantID` non-empty client-side.

---

## Protests

### `ListProtests(ctx, *ListProtestsOptions) (*PaginatedResponse[Record], error)`

`GET /api/protests/`. Bid protests from GAO, the Court of Federal Claims (COFC) and the SBA Office of Hearings and Appeals (SBA OHA). Filters: `SourceSystem`, `Outcome`, `CaseType`, `Agency`, `CaseNumber`, `SolicitationNumber`, `NaicsCode` (SBA OHA size and NAICS appeals only), `Protester`, `Search`, `FiledDate[After/Before]`, `DecisionDate[After/Before]`.

> **No ordering.** The viewset rejects ordering; `ListProtestsOptions` deliberately omits the field.

### `IterateProtests(ctx, *ListProtestsOptions) *Iterator[Record]`

### `GetProtest(ctx, caseID string, *GetEntityOptions) (*ProtestRecord, error)`

`GET /api/protests/{caseID}/`.
`caseID` is the UUID returned as `case_id` by `ListProtests`; the route does not accept a case number such as `B-423274` or `26-292`.
To look a case up by number, call `ListProtests` with `CaseNumber` set and read `case_id` from the result.

> **Typed return.** Returns `*ProtestRecord` with string fields for every scalar the API serves (`CaseID`, `SourceSystem`, `CaseNumber`, `Title`, `Protester`, `Agency`, `SolicitationNumber`, `CaseType`, `Outcome`, the four dates, `DocketURL`, `DecisionURL`, plus the opt-in `ChallengedParty`, `NaicsCode`, `SizeStandard`, `OutcomeReason`, `Judge`, `Digest` and `DecisionText`), `Organization map[string]any`, `Dockets []map[string]any`, `Decisions []map[string]any`, `ResolvedAgency map[string]any` and `ResolvedProtester map[string]any`.

Use `Shape: "...,dockets(*),decisions(*)"` to include the nested docket and decision entries.

---

## Contract appeals

Contract Disputes Act appeal decisions from the two boards of contract appeals — the Civilian Board (CBCA) for civilian agencies, and the Armed Services Board (ASBCA) for defense.

> **These are not bid protests.** An appeal disputes a contracting officer's final decision under a contract the government already awarded — a claim for money, a termination, a default. A protest challenges the award itself and lives on [`/api/protests/`](#protests). The two resources share no identifiers and no vocabulary.

### `ListContractAppeals(ctx, *ListContractAppealsOptions) (*PaginatedResponse[Record], error)`

`GET /api/contract_appeals/`. Filters: `Board`, `Docket`, `Appellant`, `Judge`, `DecisionType`, `DecisionDate[After/Before]`, `Listed`, `DocumentID`, `Search`, `Ordering`.

`Ordering` is one of `decision_date`, `appellant`, `first_listed_at` or `rank`, defaulting to `-decision_date`. `rank` is only meaningful alongside a non-empty `Search`.

`Listed` is a `*bool` so that `false` is a real filter value rather than an absent one. `Appellant` matches the contractor's name as the board published it — there is no entity resolution behind it, so a company that appears under two spellings needs two queries.

> **`Board` is the split that matters.** Docket numbering, decision-type wording and listing practice all differ between CBCA and ASBCA, so a filter tuned against one board's rows can return nothing against the other's.

> **An unshaped row carries only a core subset of the columns.** Pass `ShapeContractAppealsComprehensive` (or your own field list) when you need the rest. Every field on `ContractAppealRecord` is a pointer for exactly this reason: a missing field means "not asked for or not served", never "empty".

### `IterateContractAppeals(ctx, *ListContractAppealsOptions) *Iterator[Record]`

### `GetContractAppeal(ctx, uuid string, *GetEntityOptions) (*ContractAppealRecord, error)`

`GET /api/contract_appeals/{uuid}/`.

> **Typed return.** Returns `*ContractAppealRecord` with named fields (`UUID`, `Board`, `DocketNumbers []string`, `DocketSource`, `DocketRaw`, `DecisionDate`, `DecisionDateRaw`, `DecisionDateRepaired`, `Appellant`, `Judge`, `DecisionType`, `DecisionTypeRaw`, `URL`, `DocumentID`, `ListingURL`, `ListingYear`, `FirstListedAt`, `Listed`, `TextStatus`, `TextCharCount`, `DecisionText`, `Extra map[string]any`).

> **The decision body is `decision_text`, on the Enterprise plan only.** Below that plan the key is **absent rather than null**, so `DecisionText` stays `nil` — never read `nil` as "this decision has no text". `TextStatus` and `TextCharCount` describe the extracted text at every plan and read as a pair: a status claiming text alongside a zero character count is a document that has not yielded any. Neither `Shape*` preset names `decision_text`; ask for it explicitly.

```go
rec, err := client.GetContractAppeal(ctx, uuid, &tango.GetEntityOptions{
    Shape: "uuid,board,decision_date,decision_text",
})
```

A consolidated appeal is decided once and carries several dockets, which is why `DocketNumbers` is a slice; `Docket` matches any one of them. `Listed` reports whether the decision is still on a board listing — a board rebuilds its index in place, so a decision can drop off one without being withdrawn.

Alerts on this resource use query type `contract_appeal` and deliver `alerts.contract_appeal.match`; see [`WEBHOOKS.md`](WEBHOOKS.md).

---

## State & Local (SLED)

**Beta.** State, local and education procurement — solicitations that never appear on SAM.gov because they were never federal. Coverage is partial and grows one jurisdiction at a time.

This data does not join to the federal data: no UEI, no PIID, no agency-hierarchy key and no NAICS/PSC crosswalk. The `organization(*)` expand here is three strings, not the federal 7-key office payload.

### `ListSledOpportunities(ctx, *ListSledOpportunitiesOptions) (*PaginatedResponse[Record], error)`

`GET /api/sled/opportunities/`. Filters: `State`, `Jurisdiction`, `Status`, `Active`, `Agency`, `SolicitationNumber`, `SolicitationType`, `HasDocuments`, `RevisionKind`, `Naics`, `Nigp`, `Unspsc`, `Category`, `CategoryCode`, `Posted[After/Before]`, `ResponseDeadline[After/Before]`, `FirstSeen[After/Before]`, `ChangeSeenAfter`, `Modified[After/Before]`, `Platform`, `NativeID`, `ExternalID`, `Search`, `Ordering`, `Verbose`.

> **Leaving both `Status` and `Active` unset returns open solicitations only.** Only about a fifth of the corpus is open, and a portal drops a closed solicitation rather than restating it, so the API defaults the list to `status=open`. Set `Status` explicitly to page the whole corpus; `Status: "open|unknown"` also reaches the standing rosters and dateless RFIs that `unknown` covers. `GetSledOpportunity` returns a solicitation whatever its status.
>
> This method deliberately does not synthesize `status=open` client-side: doing so would make `Active: boolPtr(false)` unreachable, since `active=false` is the complement of open rather than an independent value.

> **`Status` is Tango's answer, not the portal's.** It is derived from the portal's word, the deadline and the clock, and refreshed every fifteen minutes. The portal's own word is served as `source_status`, is frozen at last capture, and is **not** filterable — most of what it calls open already has a passed deadline.

`Active` and `HasDocuments` are `*bool` so that `false` is a real filter value rather than an absent one. Use `Extra` for anything the struct does not name. `Verbose: true` adds `description` and `contact` to each list row; `description` is otherwise detail-only because its longest values run past 120,000 characters.

`Search` is ranked over title, agency, identifiers, category labels and description, widened by the solicitations whose *attachment text* matched. A row that matched on its description gains a `snippet` with the matching passage; a title-or-agency match carries none. Attachment matching contributes ids only.

`category_codes` scheme tagging is mid-migration, so **`Naics` matches only the small tagged share**. Use `CategoryCode` to match a code under any scheme, including the untagged pre-migration strings.

### `IterateSledOpportunities(ctx, *ListSledOpportunitiesOptions) *Iterator[Record]`

### `GetSledOpportunity(ctx, opportunityID string, *GetEntityOptions) (Record, error)`

`GET /api/sled/opportunities/{opportunityID}/`. Returns the solicitation whatever its status.

> **`meta.attachment_count` can be lower than `len(attachments)`.** Some portals auto-generate a cover sheet alongside the real documents; it is listed and flagged `is_generated_summary`, but excluded from the count and from `has_documents`. The count answers "does this record hold its solicitation package"; the array answers "what files exist".

`size_bytes` and `char_count` only mean something as a pair — 3 MB that yielded no characters is a scan awaiting OCR. `raw(*)` needs a Small plan or above and is explicitly unstable: its shape varies by portal platform.

> **The document body is `attachments(extracted_text)`, on a Small plan or above** (API 4.25.1+). It must be **named** — no `Shape*` preset includes it and `attachments(*)` does not carry it, because the API resolves the body only for a caller who asked. Its key is **absent rather than null** whenever the text is not being served: below Small (withheld and named in `meta.upgrade_hints`), on a contested document, or where it could not be resolved. A **contested document never returns text at any plan**, because its stored bytes disagree with what the record advertised.
>
> Searching document text and reading it are separate: `Search` matches inside attachment text on every plan and returns no fragment of it.

```go
row, err := client.GetSledOpportunity(ctx, id, &tango.GetEntityOptions{
    Shape: "opportunity_id,attachments(name,size_bytes,extracted_text)",
})
```

### `ListSledOpportunityRevisions(ctx, opportunityID string, *ListSledOpportunityRevisionsOptions) (*PaginatedResponse[Record], error)`

`GET /api/sled/opportunities/{opportunityID}/revisions/`. Filters: `Kind`, `SourceDeclared`, `Observed[After/Before]`.

> **`observed_at` is the scrape that saw the change, not the date the agency made it.** No state portal emits amendment notices, so `kind` is Tango's inference from the diff on about 95% of revisions, resolution is that state's crawl cadence, and history starts when Tango began reading the jurisdiction rather than when the solicitation was posted.

Unlike the `revisions(*)` expand, this route serves `enrichment` rows — Tango's own detail fetch filling in coverage rather than an agency amendment. Pass `Kind: "enrichment"` for only those. `changes` (the per-field before and after) needs a Small plan, which is why `ShapeSledRevisionsMinimal` omits it; `changed_fields` is in that shape and available at every plan.

### `GetSledCoverage(ctx) (Record, error)`

`GET /api/sled/opportunities/coverage/`. Takes no parameters; neither shaped nor paginated.

> **Call this before treating a per-state count as market size.** A thin result for a state is at least as likely to be a portal Tango does not read as a quiet market, and that is the ambiguity this endpoint exists to resolve.

Returns corpus totals plus one row per jurisdiction — the total, the count in each of the five statuses, the jurisdiction levels present, and when a solicitation there last changed. Every state row carries all five status buckets whether or not they have rows, so a total and two buckets never invite subtraction.

### `ListSledForecasts(ctx, *ListSledForecastsOptions) (*PaginatedResponse[Record], error)`

`GET /api/sled/forecasts/`. Filters: `State`, `Agency`, `ProcurementCategory`, `ProcurementMethod`, `ContractNumber`, `IncumbentName`, `Advertisement[After/Before]`, `FirstSeen[After/Before]`, `Modified[After/Before]`, `Search`, `Ordering`.

> **No liveness at all.** A forecast has no deadline to have passed, so there is no `Status` field, no `Active` field, and no open-only default. Currency is the caller's call from `estimated_advertisement_date`, which is the **start of the published quarter** rather than a posting date — `estimated_advertisement_raw` keeps the portal's own words, and a large share of rows publish no quarter at all.

`estimated_value(min,max,raw)` is parsed from a free-text award band at serve time. A band naming one number is a floor, so `max` is null — never read a missing `max` as an unbounded ceiling. `incumbent_name` is published text, not a resolved Tango entity.

### `IterateSledForecasts(ctx, *ListSledForecastsOptions) *Iterator[Record]`

### `GetSledForecast(ctx, forecastID string, *GetEntityOptions) (Record, error)`

`GET /api/sled/forecasts/{forecastID}/`.

---

## Exclusions

SAM.gov exclusions: debarments, suspensions and other ineligibility actions.

### `ListExclusions(ctx, *ListExclusionsOptions) (*PaginatedResponse[Record], error)`

`GET /api/exclusions/`. Filters: `Active *bool`, `Delisted *bool`, `ClassificationType`, `ExclusionType`, `ExclusionProgram`, `ExcludingAgencyCode`, `ExcludingAgencyName`, `UEI`, `CageCode`, `NPI`, `EntityUEI`, `ActivateDate[After/Before]`, `TerminationDate[After/Before]`, `UpdateDate[After/Before]`, `Search`, `Ordering`.

> **`Active` is derived at query time.** An exclusion is active when it is not delisted, has activated and has not terminated. Reaching a termination date changes the answer without any write, so it fires no alert. `Delisted` is different: it means SAM lifted or withdrew the exclusion.

Most exclusions name individuals and carry no UEI. `EntityUEI` matches only when an exclusion's UEI resolved to a registered entity.

### `IterateExclusions(ctx, *ListExclusionsOptions) *Iterator[Record]`

### `GetExclusion(ctx, exclusionKey string, *GetEntityOptions) (Record, error)`

`GET /api/exclusions/{exclusion_key}/`.

---

## DIBBS

Defense Logistics Agency solicitations and awards from the DLA Internet Bid Board System. Whether an RFQ or RFP is open is derived at query time from its closing date.

### `ListDibbsRfqs(ctx, *ListDibbsRfqsOptions) (*PaginatedResponse[Record], error)`

`GET /api/dibbs/rfqs/`. Filters: `Open *bool`, `NSN`, `PartNumber`, `Solicitation`, `PurchaseRequest`, `SetAside` (`"Y"` / `"N"`), `StatusCode`, `Organization`, `QuantityMin` / `QuantityMax`, `ReturnByDate[After/Before]`, `IssueDate[After/Before]`, `Search`, `Ordering`.

### `ListDibbsRfps(ctx, *ListDibbsRfpsOptions) (*PaginatedResponse[Record], error)`

`GET /api/dibbs/rfps/`. Filters: `Open *bool`, `NSN`, `PartNumber`, `Solicitation`, `BuyerCode`, `Organization`, `IssuedDate[After/Before]`, `ClosesDate[After/Before]`, `Search`, `Ordering`.

### `ListDibbsAwards(ctx, *ListDibbsAwardsOptions) (*PaginatedResponse[Record], error)`

`GET /api/dibbs/awards/`. One row per award line item. Filters: `NSN`, `PartNumber`, `Solicitation`, `AwardNumber`, `DeliveryOrderNumber`, `PurchaseRequest`, `AwardeeCage`, `Entity`, `Organization`, `AwardDate[After/Before]`, `PostedDate[After/Before]`, `TotalContractPriceMin` / `TotalContractPriceMax`, `Search`, `Ordering`.

> **Order-level money repeats per line.** `total_contract_price` is the whole order's price, carried on every line item of that order, so it does not sum across rows.

### `IterateDibbsRfqs` / `IterateDibbsRfps` / `IterateDibbsAwards`

Walk every row matching the options.

### `GetDibbsRfq` / `GetDibbsRfp` / `GetDibbsAward(ctx, uuid string, *GetEntityOptions) (Record, error)`

`GET /api/dibbs/{rfqs,rfps,awards}/{uuid}/`.

---

## SBIR / STTR

SBIR and STTR topics, and the solicitation cycles they are released under.

### `ListSbirTopics(ctx, *ListSbirTopicsOptions) (*PaginatedResponse[Record], error)`

`GET /api/sbir/topics/`. Filters: `Activity` (`"open"`, `"closed"` or `"unknown"`), `Agency` (partial match on the raw agency text, not organization-resolved), `TopicNumber`, `SolicitationNumber`, `Year`, `DocSource`, `CloseDate[After/Before]`, `OpenDate[After/Before]`, `ReleaseDate[After/Before]`, `Search`, `Ordering`.

### `ListSbirSolicitations(ctx, *ListSbirSolicitationsOptions) (*PaginatedResponse[Record], error)`

`GET /api/sbir/solicitations/`. Filters: `Activity` (`"open"` or `"closed"`), `Program` (`"SBIR"` / `"STTR"`), `SolicitationNumber`, `CycleName`, `SolicitationStatus`, `OutOfCycle *bool`, `Year`, `StartDate[After/Before]`, `EndDate[After/Before]`, `Search`, `Ordering`.

### `IterateSbirTopics` / `IterateSbirSolicitations`

Walk every row matching the options.

### `GetSbirTopic(ctx, topicID string, *GetEntityOptions) (Record, error)` / `GetSbirSolicitation(ctx, solicitationID string, *GetEntityOptions) (Record, error)`

`GET /api/sbir/topics/{topic_id}/` and `GET /api/sbir/solicitations/{solicitation_id}/`.

---

## IT Dashboard

### `ListItDashboard(ctx, *ListItDashboardOptions) (*PaginatedResponse[Record], error)`

`GET /api/itdashboard/`. Federal IT investments. **Filters are tier-gated** by the API:

- Free: `Search`
- Pro: `AgencyCode`, `TypeOfInvestment`, `UpdatedTime[After/Before]`
- Business+: `AgencyName`, `CIORating`, `CIORatingMax`, `PerformanceRisk`

`PreviousUII` finds the investment or investments that superseded a retired UII.

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
| `ListWebhookAlerts(ctx, *ListOptions)` → `*PaginatedResponse[WebhookAlert]` | `GET /api/webhooks/alerts/` (sends `Limit` as `page_size`) |
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
