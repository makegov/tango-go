# Response Shaping

Dynamic response shaping is the Tango API's signature feature: instead of always receiving every field on a resource, you tell the server exactly which fields you want, and it returns only those. Payloads stay small, responses stay fast, and the SDK doesn't have to chase schema drift.

This guide explains the shape grammar, the 21 built-in `Shape*` constants, the `Flat` / `FlatLists` modifiers, and the trade-offs to think about when picking a shape.

## What is a shape?

A shape is a comma-separated list of fields to return. The simplest shape is a flat field list:

```text
key,piid,award_date
```

You can nest fields with `parent(child1, child2)`:

```text
key,piid,recipient(display_name,uei),total_contract_value
```

You can request every field at a given level with `*`:

```text
key,piid,recipient(*)
```

You can alias fields with `::`:

```text
recipient::vendor(display_name,uei)
```

The full grammar:

```text
shape       := field_list
field_list  := field ("," field)*
field       := field_name [alias] [nested]
field_name  := identifier | "*"
alias       := "::" identifier
nested      := "(" field_list ")"
identifier  := [a-zA-Z_][a-zA-Z0-9_]*
```

> The SDK does not parse or validate shape strings client-side — it passes the value through to the server as the `shape` query parameter. The server returns `400` with a parse error in the body if the shape is malformed (surfaces as `*tango.ValidationError`).

## Using a shape

Every list and get options struct has a `Shape` field (either directly, or via embedded `ListOptions`):

```go
page, err := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{
        Shape: "key,piid,award_date,recipient(display_name)",
        Limit: 25,
    },
    AwardingAgency: "9700",
})
```

For convenience, the SDK ships 21 preset constants. Use them when you don't need a custom selector:

```go
page, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{Shape: tango.ShapeContractsMinimal},
})
```

## Shape preset constants

All 21 constants live in `shapes.go`. They mirror the `ShapeConfig.*` enums in `tango-node` and `tango-python` exactly — same names, same field selectors, same intent.

| Constant | Intended use | Notes |
| -------- | ------------ | ----- |
| `ShapeContractsMinimal` | `ListContracts` | key, piid, award_date, recipient (name), description, total_contract_value |
| `ShapeEntitiesMinimal` | `ListEntities` | uei, legal_business_name, cage_code, business_types |
| `ShapeEntitiesComprehensive` | `GetEntity` | UEI + names + NAICS/PSC + addresses + federal_obligations + congressional district |
| `ShapeForecastsMinimal` | `ListForecasts` | id, title, anticipated_award_date, fiscal_year, naics_code, status |
| `ShapeOpportunitiesMinimal` | `ListOpportunities` | opportunity_id, title, solicitation_number, response_deadline, active |
| `ShapeNoticesMinimal` | `ListNotices` | notice_id, title, solicitation_number, posted_date |
| `ShapeProtestsMinimal` | `ListProtests` | case_id, case_number, title, source_system, outcome, filed_date |
| `ShapeGrantsMinimal` | `ListGrants` | grant_id, opportunity_number, title, status, agency_code |
| `ShapeIDVsMinimal` | `ListIDVs` | key, piid, award_date, recipient (name+UEI), value, obligated, idv_type |
| `ShapeIDVsComprehensive` | `GetIDV` | full IDV envelope incl. competition, legislative_mandates, transactions, subawards summary |
| `ShapeVehiclesMinimal` | `ListVehicles` | uuid, solicitation_identifier, program_acronym, organization, vehicle_type, totals |
| `ShapeVehiclesComprehensive` | `GetVehicle` | full vehicle envelope incl. metrics |
| `ShapeVehicleAwardeesMinimal` | `ListVehicleAwardees` | uuid, key, piid, award_date, order_count, obligations, recipient (name+UEI) |
| `ShapeVehicleOrdersMinimal` | `ListVehicleOrders` | key, piid, award_date, obligated, value, description, recipient (name+UEI) |
| `ShapeOrganizationsMinimal` | `ListOrganizations` | key, fh_key, name, level, type, short_name |
| `ShapeOTAsMinimal` | `ListOTAs` | key, piid, award_date, recipient, description, value, obligated |
| `ShapeOTIDVsMinimal` | `ListOTIDVs` | key, piid, award_date, recipient, description, value, obligated, idv_type |
| `ShapeSubawardsMinimal` | `ListSubawards` | award_key, prime_recipient (uei+name), subaward_recipient (uei+name) — server rejects `id` / `amount` here |
| `ShapeGsaElibraryContractsMinimal` | `ListGsaElibraryContracts` | uuid, contract_number, schedule, recipient, idv |
| `ShapeItdashboardInvestmentsMinimal` | `ListItDashboard` | uii, agency_name, bureau_name, investment_title, type_of_investment, part_of_it_portfolio, updated_time, url |
| `ShapeItdashboardInvestmentsComprehensive` | `GetItDashboard` | adds agency_code + bureau_code |

The literal field selectors are in `shapes.go` if you want to copy-modify one — that's a fine pattern for building a custom shape that starts from a preset.

## Custom shapes

### Narrow

```go
page, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{Shape: "key,piid"},
})
```

Fastest, smallest. Best when you only need identifiers — e.g. you're collecting keys for a follow-up fetch.

### Wide nested

```go
shape := "key,piid,award_date," +
    "recipient(display_name,uei,cage_code,business_types(*))," +
    "place_of_performance(*)," +
    "awarding_office(*)"

contracts, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{Shape: shape},
})
```

Heavier than the presets, but lets you grab nested objects in one round trip.

### Wildcard

```go
shape := "*"               // every top-level field
shape := "recipient(*)"     // every field on recipient, ignore everything else
```

`*` is convenient for exploration. Don't ship it to production — it tends to return huge payloads.

### Start from a preset, then extend

```go
shape := tango.ShapeContractsMinimal + ",awarding_agency,obligated"
```

Strings are concatenable. The server applies the union of the field paths.

## Flat responses

When `Flat: true`, the server returns dotted key names instead of nested objects:

```go
page, _ := client.ListContracts(ctx, &tango.ListContractsOptions{
    ListOptions: tango.ListOptions{
        Shape: tango.ShapeContractsMinimal,
        Flat:  true,
    },
})

c := page.Results[0]
// c is a Record (map[string]any) with keys like:
//   "key", "piid", "award_date", "recipient.display_name", "description", "total_contract_value"
fmt.Println(c["recipient.display_name"])
```

The default separator is `.`. The Tango API accepts a `joiner` parameter to override it — exposed on the SDK as the `Joiner` field on certain options structs (e.g. `ListVehiclesOptions`, `AgencyContractsOptions`, `EntitySubresourceOptions`, the OTA/OTIDV options):

```go
client.ListAgencyAwardingContracts(ctx, "9700", &tango.AgencyContractsOptions{
    ListOptions: tango.ListOptions{
        Shape: "key,piid,recipient(uei,display_name)",
        Flat:  true,
    },
    Joiner: "__",
})
// Returned keys: "recipient__uei", "recipient__display_name"
```

> **Note.** Unlike the Node/Python SDKs, tango-go does not unflatten responses on the client side. `Flat: true` is a request-side option — the response is a `Record` with the keys the server returned. If you want nested objects, leave `Flat` as `false` (the default).

### Flat lists

`FlatLists: true` extends `Flat` to list-valued nested fields. Only meaningful when `Flat: true`. Use when you need a fully tabular result — e.g. when feeding the response into a CSV or DataFrame.

## Trade-offs

| Want | Use |
| ---- | --- |
| Smallest payload + fastest response | `Shape<Resource>Minimal` |
| Everything in one call | `Shape<Resource>Comprehensive` (on `Get*`) or `*` |
| Tabular output (CSV / DataFrame) | `Shape*Minimal` + `Flat: true` (+ `FlatLists: true` if needed) |
| Only identifiers (for follow-up fetches) | Custom: `"key,piid"` or `"uei,legal_business_name"` |
| Stable schema you don't have to chase | Custom shape — pick the fields, the server won't add unexpected ones |

The `Shape*Minimal` presets are tuned for the common list view (results table, filters, pagination). The `Shape*Comprehensive` presets are tuned for the detail view (one record, every relevant field). Both leave headroom — they intentionally don't ask for every field on every nested object, because that's where payload size blows up.

## Server-side validation

Shape errors surface as `*tango.ValidationError` (HTTP 400) with the server's parse error on `ResponseData`. Common gotchas:

- **Field doesn't exist on the resource.** Misspelled `recipiient` → 400. The server lists valid fields in the error body.
- **Nested syntax on a non-nested field.** `award_date(year)` → 400; `award_date` isn't a parent.
- **Listing a server-rejected field.** `ListSubawards` rejects `id` and `amount` in shapes (see `ShapeSubawardsMinimal`'s godoc). The server returns 400 with an explanation.
- **Empty shape string.** Empty means "server default" — that's fine and the SDK passes it through. Only an explicit malformed value is rejected.

When in doubt, drop down to the wildcard once to see what the server returns, then trim from there.
