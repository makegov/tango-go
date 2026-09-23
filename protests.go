package tango

import (
	"context"
	"net/url"
)

// ProtestRecord is the typed return model for GetProtest: one protest case from GAO, the Court of Federal Claims (COFC) or the SBA Office of Hearings and Appeals (SBA OHA).
//
// Every field is optional, because which fields are populated depends on the source system and on the shape requested.
// A JSON null decodes to the field's zero value.
// Dates are YYYY-MM-DD strings.
//
// The SBA OHA fields (ChallengedParty, NaicsCode, SizeStandard, OutcomeReason, Judge) are empty on GAO and COFC cases and are returned only when named in the shape.
// Digest and DecisionText are also returned only when named; DecisionText needs an Enterprise plan.
// Dockets, Decisions, ResolvedAgency and ResolvedProtester are expands and are returned only when named in the shape (for example `dockets(*)`); the resolved_* expands need a Pro plan or above.
type ProtestRecord struct {
	CaseID             string `json:"case_id,omitempty"`
	SourceSystem       string `json:"source_system,omitempty"`
	CaseNumber         string `json:"case_number,omitempty"`
	Title              string `json:"title,omitempty"`
	Protester          string `json:"protester,omitempty"`
	ChallengedParty    string `json:"challenged_party,omitempty"`
	Agency             string `json:"agency,omitempty"`
	SolicitationNumber string `json:"solicitation_number,omitempty"`
	NaicsCode          string `json:"naics_code,omitempty"`
	SizeStandard       string `json:"size_standard,omitempty"`
	CaseType           string `json:"case_type,omitempty"`
	Outcome            string `json:"outcome,omitempty"`
	OutcomeReason      string `json:"outcome_reason,omitempty"`
	Judge              string `json:"judge,omitempty"`
	FiledDate          string `json:"filed_date,omitempty"`
	PostedDate         string `json:"posted_date,omitempty"`
	DecisionDate       string `json:"decision_date,omitempty"`
	DueDate            string `json:"due_date,omitempty"`
	DocketURL          string `json:"docket_url,omitempty"`
	DecisionURL        string `json:"decision_url,omitempty"`
	Digest             string `json:"digest,omitempty"`
	DecisionText       string `json:"decision_text,omitempty"`

	// Organization is the resolved department / agency / office for the protested agency, or nil when it could not be resolved.
	Organization map[string]any `json:"organization,omitempty"`

	// Dockets holds one entry per docket filed under the case (keys such as docket_number, filed_date, outcome).
	Dockets []map[string]any `json:"dockets,omitempty"`

	// Decisions holds the opinions and orders issued on the case (keys such as title, decision_date, outcome, judges, document_url).
	Decisions []map[string]any `json:"decisions,omitempty"`

	// ResolvedAgency is the best-guess organization match (key, name, match_confidence, rationale).
	ResolvedAgency map[string]any `json:"resolved_agency,omitempty"`

	// ResolvedProtester is the best-guess entity match (uei, name, match_confidence, rationale).
	ResolvedProtester map[string]any `json:"resolved_protester,omitempty"`

	// Extra is never populated from JSON; callers may use it to carry additional fields when re-marshaling.
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
	// NaicsCode matches the NAICS code at issue in an SBA OHA size or NAICS appeal; GAO and COFC cases never match it.
	NaicsCode string
	Protester string
	Search    string

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
	setIfNotEmpty(q, "naics_code", o.NaicsCode)
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

// ListProtests queries /api/protests/ — bid protests from GAO, COFC and SBA OHA.
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

// GetProtest fetches a single protest case by its case_id, the UUID returned as `case_id` by ListProtests.
//
// The route does not accept a case number such as "B-423274" or "26-292"; to find a case by number, call ListProtests with CaseNumber set and read `case_id` from the result.
// Returns a typed *ProtestRecord.
// Use a shape such as "case_id,title,dockets(*),decisions(*)" to include the nested docket and decision entries.
func (c *Client) GetProtest(ctx context.Context, caseID string, opts *GetEntityOptions) (*ProtestRecord, error) {
	if caseID == "" {
		return nil, &ValidationError{&APIError{Message: "protest case_id is required"}}
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
	return getGeneric[*ProtestRecord](ctx, c, "/api/protests/"+pathEscape(caseID)+"/", q)
}
