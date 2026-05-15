package tango

import (
	"context"
	"net/url"
)

// ListIDVsOptions filters /api/idvs/ (Indefinite Delivery Vehicles).
type ListIDVsOptions struct {
	ListOptions

	AwardDate              string
	AwardDateGte           string
	AwardDateLte           string
	AwardingAgency         string
	FundingAgency          string
	ExpiringGte            string
	ExpiringLte            string
	FiscalYear             string
	FiscalYearGte          string
	FiscalYearLte          string
	IDVType                string
	LastDateToOrderGte     string
	LastDateToOrderLte     string
	NAICS                  string
	Ordering               string
	PIID                   string
	PopStartDateGte        string
	PopStartDateLte        string
	PSC                    string
	Recipient              string
	Search                 string
	SetAside               string
	SolicitationIdentifier string
	UEI                    string

	Extra map[string]any
}

func (o *ListIDVsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "expiring_gte", o.ExpiringGte)
	setIfNotEmpty(q, "expiring_lte", o.ExpiringLte)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "idv_type", o.IDVType)
	setIfNotEmpty(q, "last_date_to_order_gte", o.LastDateToOrderGte)
	setIfNotEmpty(q, "last_date_to_order_lte", o.LastDateToOrderLte)
	setIfNotEmpty(q, "naics", o.NAICS)
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "recipient", o.Recipient)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "set_aside", o.SetAside)
	setIfNotEmpty(q, "solicitation_identifier", o.SolicitationIdentifier)
	setIfNotEmpty(q, "uei", o.UEI)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListIDVs queries /api/idvs/.
func (c *Client) ListIDVs(ctx context.Context, opts *ListIDVsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/idvs/", q)
}

// GetIDV fetches a single IDV by key.
func (c *Client) GetIDV(ctx context.Context, key string, opts *GetEntityOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "IDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		setIfNotEmpty(q, "shape", opts.Shape)
		if opts.Flat {
			q.Set("flat", "true")
		}
		if opts.FlatLists {
			q.Set("flat_lists", "true")
		}
	}
	return getGeneric[Record](ctx, c, "/api/idvs/"+pathEscape(key)+"/", q)
}

// IterateIDVs walks every IDV matching opts.
func (c *Client) IterateIDVs(ctx context.Context, opts *ListIDVsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListIDVsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListIDVs(ctx, &next)
		},
	}
}
