package tango

import (
	"context"
	"net/url"
)

// ProtestRecord is the typed return model for GetProtest. Mirrors the
// canonical GAO / Court of Federal Claims protest case schema as defined
// in tango-node/src/types.ts:137-151.
//
// All fields except Docket are optional (protests come from multiple
// source systems and not every field is always populated). Extra captures
// any wire fields not first-classed on the struct so callers don't lose
// them — the same pattern used by the request-options structs in this
// package.
//
// Note: this is implementer-B's local typed model; tango-go may
// consolidate typed records into a shared models.go in a later wave.
type ProtestRecord struct {
	CaseID            string           `json:"case_id,omitempty"`
	CaseNumber        string           `json:"case_number,omitempty"`
	SourceSystem      string           `json:"source_system,omitempty"`
	Outcome           string           `json:"outcome,omitempty"`
	CaseType          string           `json:"case_type,omitempty"`
	FiledDate         string           `json:"filed_date,omitempty"`
	DecisionDate      string           `json:"decision_date,omitempty"`
	Agency            map[string]any   `json:"agency,omitempty"`
	Protester         map[string]any   `json:"protester,omitempty"`
	ResolvedAgency    map[string]any   `json:"resolved_agency,omitempty"`
	ResolvedProtester map[string]any   `json:"resolved_protester,omitempty"`
	Docket            []map[string]any `json:"docket,omitempty"`

	// Extra captures any additional fields returned by the API that
	// aren't first-classed above. Populated by callers who need to
	// preserve unknown fields when re-marshaling; not auto-populated
	// from JSON.
	Extra map[string]any `json:"-"`
}

// ListProtestsOptions filters /api/protests/. Mirrors the Node
// `ListProtestsOptions` interface and the Python `list_protests` kwargs.
//
// The protests viewset does not accept ordering (the server rejects it),
// so this struct intentionally omits an Ordering field.
type ListProtestsOptions struct {
	ListOptions

	SourceSystem       string
	Outcome            string
	CaseType           string
	Agency             string
	CaseNumber         string
	SolicitationNumber string
	Protester          string
	Search             string

	// Date bounds (wire format: YYYY-MM-DD). Note the
	// `_after` / `_before` suffix convention — unique to protests; the
	// Python client uses these exact names.
	FiledDateAfter     string
	FiledDateBefore    string
	DecisionDateAfter  string
	DecisionDateBefore string

	Extra map[string]any
}

func (o *ListProtestsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "source_system", o.SourceSystem)
	setIfNotEmpty(q, "outcome", o.Outcome)
	setIfNotEmpty(q, "case_type", o.CaseType)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "case_number", o.CaseNumber)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	setIfNotEmpty(q, "protester", o.Protester)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "filed_date_after", o.FiledDateAfter)
	setIfNotEmpty(q, "filed_date_before", o.FiledDateBefore)
	setIfNotEmpty(q, "decision_date_after", o.DecisionDateAfter)
	setIfNotEmpty(q, "decision_date_before", o.DecisionDateBefore)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListProtests queries /api/protests/ — bid protests from GAO + COFC.
func (c *Client) ListProtests(ctx context.Context, opts *ListProtestsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/protests/", q)
}

// IterateProtests walks every protest matching opts.
func (c *Client) IterateProtests(ctx context.Context, opts *ListProtestsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListProtestsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListProtests(ctx, &next)
		},
	}
}

// GetProtest fetches a single protest by case number / case ID.
//
// Returns a typed *ProtestRecord (one of the few typed-struct return
// values in the SDK, mirroring the Node `ProtestRecord` interface). Use
// shape=...,dockets(...) to include nested docket entries.
func (c *Client) GetProtest(ctx context.Context, caseNumber string, opts *GetEntityOptions) (*ProtestRecord, error) {
	if caseNumber == "" {
		return nil, &ValidationError{&APIError{Message: "protest case number is required"}}
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
	return getGeneric[*ProtestRecord](ctx, c, "/api/protests/"+pathEscape(caseNumber)+"/", q)
}
