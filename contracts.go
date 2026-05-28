package tango

import (
	"context"
	"net/url"
	"strconv"
)

// ListContractsOptions are the filters and pagination params accepted by
// ListContracts. All fields are optional; embed ListOptions for
// pagination + shape control. SDK-friendly aliases (KeywordSearch,
// RecipientName, NAICSCode, PSCCode, RecipientUEI, SetAsideType) map to
// the canonical API names — passing both prefers the SDK alias to mirror
// the Node/Python SDKs.
type ListContractsOptions struct {
	ListOptions

	// Date / FY / dollar bounds
	AwardDate     string
	AwardDateGte  string
	AwardDateLte  string
	AwardType     string
	FiscalYear    string // accept as string to allow "2024" / "FY24" / range expressions
	FiscalYearGte string
	FiscalYearLte string
	ObligatedGte  string
	ObligatedLte  string

	// Period of performance / expiring
	PopStartDateGte string
	PopStartDateLte string
	PopEndDateGte   string
	PopEndDateLte   string
	ExpiringGte     string
	ExpiringLte     string

	// Agencies / identifiers
	AwardingAgency         string
	FundingAgency          string
	PIID                   string
	SolicitationIdentifier string
	NAICS                  string
	PSC                    string
	Recipient              string
	UEI                    string
	SetAside               string

	// SDK-friendly aliases (mirror Node/Python)
	NAICSCode     string // -> naics
	PSCCode       string // -> psc
	RecipientName string // -> recipient
	RecipientUEI  string // -> uei
	SetAsideType  string // -> set_aside

	// Search + ordering
	Search   string
	Keyword  string // -> search
	Ordering string
	Sort     string // combined with Order -> ordering
	Order    string // "asc" | "desc"

	// Extra is an escape hatch for filter keys not yet first-classed on
	// this struct. Values are stringified with fmt.Sprint.
	Extra map[string]any
}

func (o *ListContractsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)

	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "award_type", o.AwardType)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "obligated_gte", o.ObligatedGte)
	setIfNotEmpty(q, "obligated_lte", o.ObligatedLte)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "pop_end_date_gte", o.PopEndDateGte)
	setIfNotEmpty(q, "pop_end_date_lte", o.PopEndDateLte)
	setIfNotEmpty(q, "expiring_gte", o.ExpiringGte)
	setIfNotEmpty(q, "expiring_lte", o.ExpiringLte)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "solicitation_identifier", o.SolicitationIdentifier)
	setIfNotEmpty(q, "naics", firstNonEmpty(o.NAICSCode, o.NAICS))
	setIfNotEmpty(q, "psc", firstNonEmpty(o.PSCCode, o.PSC))
	setIfNotEmpty(q, "recipient", firstNonEmpty(o.RecipientName, o.Recipient))
	setIfNotEmpty(q, "uei", firstNonEmpty(o.RecipientUEI, o.UEI))
	setIfNotEmpty(q, "set_aside", firstNonEmpty(o.SetAsideType, o.SetAside))
	setIfNotEmpty(q, "search", firstNonEmpty(o.Keyword, o.Search))

	if o.Sort != "" {
		prefix := ""
		if o.Order == "desc" {
			prefix = "-"
		}
		q.Set("ordering", prefix+o.Sort)
	} else {
		setIfNotEmpty(q, "ordering", o.Ordering)
	}

	for k, v := range o.Extra {
		switch t := v.(type) {
		case string:
			if t != "" {
				q.Set(k, t)
			}
		case int:
			q.Set(k, strconv.Itoa(t))
		case bool:
			q.Set(k, strconv.FormatBool(t))
		default:
			// best-effort stringify
			q.Set(k, valueToString(v))
		}
	}
	return q
}

// ListContracts queries /api/contracts/.
func (c *Client) ListContracts(ctx context.Context, opts *ListContractsOptions) (*PaginatedResponse[Record], error) {
	var q url.Values
	if opts != nil {
		q = opts.toQuery()
	} else {
		q = url.Values{}
	}
	return listGeneric[Record](ctx, c, "/api/contracts/", q)
}

// IterateContracts returns an Iterator that walks every contract matching
// the given options. The iterator follows ?page= or ?cursor= on the
// server's next URL automatically.
//
// Classic loop:
//
//	it := client.IterateContracts(ctx, &tango.ListContractsOptions{AwardingAgency: "9700"})
//	for it.Next() {
//	    fmt.Println(it.Item()["piid"])
//	}
//	if err := it.Err(); err != nil { return err }
//
// Go 1.23+ range-over-func:
//
//	for c, err := range client.IterateContracts(ctx, opts).Seq() {
//	    if err != nil { return err }
//	    fmt.Println(c["piid"])
//	}
func (c *Client) IterateContracts(ctx context.Context, opts *ListContractsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListContractsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListContracts(ctx, &next)
		},
	}
}

// GetContract fetches a single federal contract record by its key
// (/api/contracts/{key}/).
func (c *Client) GetContract(ctx context.Context, key string, opts *ListOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "contract key is required"}}
	}
	q := url.Values{}
	opts.applyTo(q)
	return getGeneric[Record](ctx, c, "/api/contracts/"+pathEscape(key)+"/", q)
}

// ListContractSubawards lists subawards reported against a single prime
// contract (/api/contracts/{key}/subawards/).
func (c *Client) ListContractSubawards(ctx context.Context, key string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "contract key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/contracts/"+pathEscape(key)+"/subawards/", q)
}

// ListContractTransactions lists the raw transaction history backing a single
// contract (/api/contracts/{key}/transactions/).
func (c *Client) ListContractTransactions(ctx context.Context, key string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "contract key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/contracts/"+pathEscape(key)+"/transactions/", q)
}

func valueToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}
