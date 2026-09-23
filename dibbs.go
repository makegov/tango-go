package tango

import (
	"context"
	"net/url"
)

// DLA Internet Bid Board System (DIBBS) — /api/dibbs/rfqs/, /api/dibbs/rfps/ and /api/dibbs/awards/.
//
// Every DIBBS record belongs to the Defense Logistics Agency.
// Whether an RFQ or RFP is open is derived at query time from its closing date, so a solicitation reaching that date changes Open without any write and fires no alert.

// ListDibbsRfqsOptions filters /api/dibbs/rfqs/ — DLA requests for quotation.
type ListDibbsRfqsOptions struct {
	ListOptions

	// Open filters on whether the RFQ still accepts quotes today (its return_by_date has not passed).
	// A pointer so that false is a real filter value rather than an absent one.
	Open *bool

	// NSN, PartNumber, PurchaseRequest and StatusCode accept several values joined with "|".
	NSN             string
	PartNumber      string
	Solicitation    string
	PurchaseRequest string
	// SetAside is "Y" or "N".
	SetAside   string
	StatusCode string
	// Organization matches the resolved organization's federal hierarchy key.
	Organization string

	QuantityMin int
	QuantityMax int

	// Date bounds (wire format: YYYY-MM-DD).
	ReturnByDateAfter  string
	ReturnByDateBefore string
	IssueDateAfter     string
	IssueDateBefore    string

	Search string
	// Ordering is one of issue_date, modified, quantity or return_by_date; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListDibbsRfqsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotNilBool(q, "open", o.Open)
	setIfNotEmpty(q, "nsn", o.NSN)
	setIfNotEmpty(q, "part_number", o.PartNumber)
	setIfNotEmpty(q, "solicitation", o.Solicitation)
	setIfNotEmpty(q, "purchase_request", o.PurchaseRequest)
	setIfNotEmpty(q, "set_aside", o.SetAside)
	setIfNotEmpty(q, "status_code", o.StatusCode)
	setIfNotEmpty(q, "organization", o.Organization)
	setIfNonZeroInt(q, "quantity_min", o.QuantityMin)
	setIfNonZeroInt(q, "quantity_max", o.QuantityMax)
	setIfNotEmpty(q, "return_by_date_after", o.ReturnByDateAfter)
	setIfNotEmpty(q, "return_by_date_before", o.ReturnByDateBefore)
	setIfNotEmpty(q, "issue_date_after", o.IssueDateAfter)
	setIfNotEmpty(q, "issue_date_before", o.IssueDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListDibbsRfqs queries /api/dibbs/rfqs/.
func (c *Client) ListDibbsRfqs(ctx context.Context, opts *ListDibbsRfqsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/dibbs/rfqs/", opts.toQuery())
}

// IterateDibbsRfqs walks every RFQ matching opts.
func (c *Client) IterateDibbsRfqs(ctx context.Context, opts *ListDibbsRfqsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListDibbsRfqsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListDibbsRfqs(ctx, &next)
		},
	}
}

// GetDibbsRfq fetches one RFQ by uuid (/api/dibbs/rfqs/{uuid}/).
func (c *Client) GetDibbsRfq(ctx context.Context, uuid string, opts *GetEntityOptions) (Record, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "DIBBS RFQ uuid is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/dibbs/rfqs/"+pathEscape(uuid)+"/", opts.toQuery())
}

// ListDibbsRfpsOptions filters /api/dibbs/rfps/ — DLA requests for proposal.
type ListDibbsRfpsOptions struct {
	ListOptions

	// Open filters on whether the solicitation still accepts offers today (its closes_date has not passed).
	Open *bool

	// NSN, PartNumber and BuyerCode accept several values joined with "|".
	NSN          string
	PartNumber   string
	Solicitation string
	BuyerCode    string
	Organization string

	// Date bounds (wire format: YYYY-MM-DD).
	IssuedDateAfter  string
	IssuedDateBefore string
	ClosesDateAfter  string
	ClosesDateBefore string

	Search string
	// Ordering is one of closes_date, issued_date or modified; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListDibbsRfpsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotNilBool(q, "open", o.Open)
	setIfNotEmpty(q, "nsn", o.NSN)
	setIfNotEmpty(q, "part_number", o.PartNumber)
	setIfNotEmpty(q, "solicitation", o.Solicitation)
	setIfNotEmpty(q, "buyer_code", o.BuyerCode)
	setIfNotEmpty(q, "organization", o.Organization)
	setIfNotEmpty(q, "issued_date_after", o.IssuedDateAfter)
	setIfNotEmpty(q, "issued_date_before", o.IssuedDateBefore)
	setIfNotEmpty(q, "closes_date_after", o.ClosesDateAfter)
	setIfNotEmpty(q, "closes_date_before", o.ClosesDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListDibbsRfps queries /api/dibbs/rfps/.
func (c *Client) ListDibbsRfps(ctx context.Context, opts *ListDibbsRfpsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/dibbs/rfps/", opts.toQuery())
}

// IterateDibbsRfps walks every RFP matching opts.
func (c *Client) IterateDibbsRfps(ctx context.Context, opts *ListDibbsRfpsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListDibbsRfpsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListDibbsRfps(ctx, &next)
		},
	}
}

// GetDibbsRfp fetches one RFP by uuid (/api/dibbs/rfps/{uuid}/).
func (c *Client) GetDibbsRfp(ctx context.Context, uuid string, opts *GetEntityOptions) (Record, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "DIBBS RFP uuid is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/dibbs/rfps/"+pathEscape(uuid)+"/", opts.toQuery())
}

// ListDibbsAwardsOptions filters /api/dibbs/awards/ — DLA award line items.
//
// Each row is one line item of an order, so order-level money does not sum across rows.
type ListDibbsAwardsOptions struct {
	ListOptions

	NSN          string
	PartNumber   string
	Solicitation string
	// AwardNumber is order-scoped: one award number spans every line item of the order.
	AwardNumber         string
	DeliveryOrderNumber string
	PurchaseRequest     string
	// AwardeeCage is set on every row, whether or not the awardee resolved to a Tango entity.
	AwardeeCage string
	// Entity matches the UEI of the resolved Tango entity, set only when the awardee's CAGE code matched a registered entity.
	Entity       string
	Organization string

	// Date bounds (wire format: YYYY-MM-DD).
	AwardDateAfter   string
	AwardDateBefore  string
	PostedDateAfter  string
	PostedDateBefore string

	// TotalContractPriceMin and TotalContractPriceMax bound the order's total contract price, which repeats on every line item of that order.
	TotalContractPriceMin string
	TotalContractPriceMax string

	Search string
	// Ordering is one of award_date, modified, posted_date or total_contract_price; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListDibbsAwardsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "nsn", o.NSN)
	setIfNotEmpty(q, "part_number", o.PartNumber)
	setIfNotEmpty(q, "solicitation", o.Solicitation)
	setIfNotEmpty(q, "award_number", o.AwardNumber)
	setIfNotEmpty(q, "delivery_order_number", o.DeliveryOrderNumber)
	setIfNotEmpty(q, "purchase_request", o.PurchaseRequest)
	setIfNotEmpty(q, "awardee_cage", o.AwardeeCage)
	setIfNotEmpty(q, "entity", o.Entity)
	setIfNotEmpty(q, "organization", o.Organization)
	setIfNotEmpty(q, "award_date_after", o.AwardDateAfter)
	setIfNotEmpty(q, "award_date_before", o.AwardDateBefore)
	setIfNotEmpty(q, "posted_date_after", o.PostedDateAfter)
	setIfNotEmpty(q, "posted_date_before", o.PostedDateBefore)
	setIfNotEmpty(q, "total_contract_price_min", o.TotalContractPriceMin)
	setIfNotEmpty(q, "total_contract_price_max", o.TotalContractPriceMax)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListDibbsAwards queries /api/dibbs/awards/.
func (c *Client) ListDibbsAwards(ctx context.Context, opts *ListDibbsAwardsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/dibbs/awards/", opts.toQuery())
}

// IterateDibbsAwards walks every award line item matching opts.
func (c *Client) IterateDibbsAwards(ctx context.Context, opts *ListDibbsAwardsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListDibbsAwardsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListDibbsAwards(ctx, &next)
		},
	}
}

// GetDibbsAward fetches one award line item by uuid (/api/dibbs/awards/{uuid}/).
func (c *Client) GetDibbsAward(ctx context.Context, uuid string, opts *GetEntityOptions) (Record, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "DIBBS award uuid is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/dibbs/awards/"+pathEscape(uuid)+"/", opts.toQuery())
}
