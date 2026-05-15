package tango

import (
	"context"
	"net/url"
)

// ListVehiclesOptions filters /api/vehicles/ (contracting vehicles).
type ListVehiclesOptions struct {
	ListOptions

	Joiner                string
	Search                string
	VehicleType           string
	TypeOfIDC             string
	ContractType          string
	SetAside              string
	WhoCanUse             string
	NAICSCode             string
	PSCCode               string
	ProgramAcronym        string
	Agency                string
	OrganizationID        string
	TotalObligatedMin     string
	TotalObligatedMax     string
	IDVCountMin           int
	IDVCountMax           int
	OrderCountMin         int
	OrderCountMax         int
	FiscalYear            string
	AwardDateAfter        string
	AwardDateBefore       string
	LastDateToOrderAfter  string
	LastDateToOrderBefore string
	// Server enforces a strict ordering allowlist; other values 400.
	Ordering string

	Extra map[string]any
}

func (o *ListVehiclesOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "joiner", o.Joiner)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "vehicle_type", o.VehicleType)
	setIfNotEmpty(q, "type_of_idc", o.TypeOfIDC)
	setIfNotEmpty(q, "contract_type", o.ContractType)
	setIfNotEmpty(q, "set_aside", o.SetAside)
	setIfNotEmpty(q, "who_can_use", o.WhoCanUse)
	setIfNotEmpty(q, "naics_code", o.NAICSCode)
	setIfNotEmpty(q, "psc_code", o.PSCCode)
	setIfNotEmpty(q, "program_acronym", o.ProgramAcronym)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "organization_id", o.OrganizationID)
	setIfNotEmpty(q, "total_obligated_min", o.TotalObligatedMin)
	setIfNotEmpty(q, "total_obligated_max", o.TotalObligatedMax)
	setIfNonZeroInt(q, "idv_count_min", o.IDVCountMin)
	setIfNonZeroInt(q, "idv_count_max", o.IDVCountMax)
	setIfNonZeroInt(q, "order_count_min", o.OrderCountMin)
	setIfNonZeroInt(q, "order_count_max", o.OrderCountMax)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "award_date_after", o.AwardDateAfter)
	setIfNotEmpty(q, "award_date_before", o.AwardDateBefore)
	setIfNotEmpty(q, "last_date_to_order_after", o.LastDateToOrderAfter)
	setIfNotEmpty(q, "last_date_to_order_before", o.LastDateToOrderBefore)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListVehicles queries /api/vehicles/.
func (c *Client) ListVehicles(ctx context.Context, opts *ListVehiclesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/vehicles/", q)
}

// GetVehicle fetches a single vehicle by UUID.
func (c *Client) GetVehicle(ctx context.Context, uuid string, opts *GetEntityOptions) (Record, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "vehicle uuid is required"}}
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
	return getGeneric[Record](ctx, c, "/api/vehicles/"+pathEscape(uuid)+"/", q)
}

// IterateVehicles walks every vehicle matching opts.
func (c *Client) IterateVehicles(ctx context.Context, opts *ListVehiclesOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListVehiclesOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListVehicles(ctx, &next)
		},
	}
}
