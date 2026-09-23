package tango

import (
	"context"
	"net/url"
)

// Contract Disputes Act appeal decisions — /api/contract_appeals/.
//
// An appeal disputes a contracting officer's final decision under a contract the government already awarded: a claim for money, a termination, a default.
// It is not a bid protest, which challenges the award itself and lives on /api/protests/.
// The two resources share no identifiers and no vocabulary — a decision here is a board's ruling between the government and its own contractor, not a bidder's challenge.
//
// Two boards publish the decisions: the Civilian Board of Contract Appeals ("cbca") for civilian agencies, and the Armed Services Board of Contract Appeals ("asbca") for defense.
// Board is the only reliable split between them, because docket numbering, decision-type vocabulary and listing practice all differ by board.

// ContractAppealRecord is the typed return model for GetContractAppeal.
//
// Every field is a pointer (or a slice/map) so the zero value means "absent" rather than "the server sent an empty string".
// That distinction is load-bearing here: an unshaped list response carries only a core subset of the columns, so most of this struct is legitimately missing from a row the caller did not shape, and DecisionText is absent entirely below the Enterprise plan.
// Unknown wire fields land in Extra.
type ContractAppealRecord struct {
	UUID  *string `json:"uuid,omitempty"`
	Board *string `json:"board,omitempty"`

	// DocketNumbers holds every docket the decision was issued under — consolidated appeals are decided once and carry several.
	DocketNumbers []string `json:"docket_numbers,omitempty"`

	// DocketSource names where the dockets were read from, and DocketRaw keeps the board's own unparsed string.
	// Compare the two when a docket looks wrong.
	DocketSource *string `json:"docket_source,omitempty"`
	DocketRaw    *string `json:"docket_raw,omitempty"`

	// DecisionDate is the parsed date (YYYY-MM-DD).
	// DecisionDateRaw keeps the board's published wording, and DecisionDateRepaired is true when the parsed date came from a repair pass rather than straight from the listing.
	DecisionDate         *string `json:"decision_date,omitempty"`
	DecisionDateRaw      *string `json:"decision_date_raw,omitempty"`
	DecisionDateRepaired *bool   `json:"decision_date_repaired,omitempty"`

	// Appellant is the contractor as the board named it, not a resolved Tango entity.
	// There is no UEI on an appeal.
	Appellant *string `json:"appellant,omitempty"`

	Judge *string `json:"judge,omitempty"`

	// DecisionType is the normalized disposition; DecisionTypeRaw keeps the board's own label, which the two boards word differently for the same outcome.
	DecisionType    *string `json:"decision_type,omitempty"`
	DecisionTypeRaw *string `json:"decision_type_raw,omitempty"`

	// URL points at the decision document; ListingURL at the board index page it was found on, and ListingYear at that index's year.
	URL         *string `json:"url,omitempty"`
	DocumentID  *string `json:"document_id,omitempty"`
	ListingURL  *string `json:"listing_url,omitempty"`
	ListingYear *int    `json:"listing_year,omitempty"`

	// FirstListedAt is when the decision was first observed on a board listing, and Listed reports whether it is still on one.
	// A board rebuilds its index in place, so a decision can drop off a listing without being withdrawn.
	FirstListedAt *string `json:"first_listed_at,omitempty"`
	Listed        *bool   `json:"listed,omitempty"`

	// TextStatus and TextCharCount describe the extracted decision text without serving it.
	// They read as a pair: a status that claims text alongside a zero character count is a document that has not yielded any.
	TextStatus    *string `json:"text_status,omitempty"`
	TextCharCount *int    `json:"text_char_count,omitempty"`

	// DecisionText is the full decision body, served on the Enterprise plan only.
	// The key is ABSENT rather than null below that plan, which is why this is a pointer: nil means "not served to you", never "the decision has no text".
	DecisionText *string `json:"decision_text,omitempty"`

	// Extra captures any forward-compatible fields the server adds that aren't in the typed surface yet.
	// Not serialized on the way out.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *ContractAppealRecord) UnmarshalJSON(data []byte) error {
	type alias ContractAppealRecord
	out := (*alias)(a)
	return unmarshalWithExtra(data, out)
}

// ListContractAppealsOptions filters /api/contract_appeals/.
type ListContractAppealsOptions struct {
	ListOptions

	// Board is "cbca" or "asbca".
	// It is the split that matters: docket numbering, decision-type wording and listing practice all differ between the two boards.
	Board string

	// Docket matches a docket number.
	// A consolidated appeal carries several, and matching any one of them returns the decision.
	Docket string

	// Appellant matches the contractor's name as the board published it.
	// There is no entity resolution behind it, so a company that appears under two spellings needs two queries.
	Appellant string

	// Judge matches the deciding judge's name as published.
	Judge string

	// DecisionType matches the normalized disposition.
	// The board's own label is served as decision_type_raw and is not filterable.
	DecisionType string

	// Date bounds on the decision date (wire format: YYYY-MM-DD).
	DecisionDateAfter  string
	DecisionDateBefore string

	// Listed filters on whether the decision is still on a board listing.
	// A pointer so that false is a real filter value rather than an absent one.
	Listed *bool

	// DocumentID matches the board's own document identifier.
	DocumentID string

	// Search is a ranked full-text query over the decision.
	Search string

	// Ordering is one of decision_date, appellant, first_listed_at or rank.
	// The default is -decision_date, newest first.
	// "rank" is only meaningful with a non-empty Search.
	Ordering string

	Extra map[string]any
}

func (o *ListContractAppealsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "board", o.Board)
	setIfNotEmpty(q, "docket", o.Docket)
	setIfNotEmpty(q, "appellant", o.Appellant)
	setIfNotEmpty(q, "judge", o.Judge)
	setIfNotEmpty(q, "decision_type", o.DecisionType)
	setIfNotEmpty(q, "decision_date_after", o.DecisionDateAfter)
	setIfNotEmpty(q, "decision_date_before", o.DecisionDateBefore)
	setIfNotNilBool(q, "listed", o.Listed)
	setIfNotEmpty(q, "document_id", o.DocumentID)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListContractAppeals queries /api/contract_appeals/ — Contract Disputes Act appeal decisions from the CBCA and ASBCA.
//
// An unshaped row carries only a core subset of the columns.
// Pass ShapeContractAppealsComprehensive (or your own field list) when you need the rest.
func (c *Client) ListContractAppeals(ctx context.Context, opts *ListContractAppealsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/contract_appeals/", q)
}

// IterateContractAppeals walks every appeal decision matching opts.
func (c *Client) IterateContractAppeals(ctx context.Context, opts *ListContractAppealsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListContractAppealsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListContractAppeals(ctx, &next)
		},
	}
}

// GetContractAppeal fetches one appeal decision by uuid.
//
// Returns a typed *ContractAppealRecord, like GetProtest does for a protest.
// The full decision body is the decision_text field, served on the Enterprise plan only; below that plan the key is absent rather than null, so DecisionText stays nil.
func (c *Client) GetContractAppeal(ctx context.Context, uuid string, opts *GetEntityOptions) (*ContractAppealRecord, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "contract appeal uuid is required"}}
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
	return getGeneric[*ContractAppealRecord](ctx, c, "/api/contract_appeals/"+pathEscape(uuid)+"/", q)
}
