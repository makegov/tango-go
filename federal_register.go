package tango

import (
	"context"
	"net/url"
)

// Federal Register documents — /api/federal_register/ (Tango API 5.3.0).
//
// These are the rules, proposed rules, notices and presidential documents published in the Federal Register since 1994.
// The detail route takes a document's uuid, not its document_number: the Federal Register reused some document numbers before 2016, so a number is not unique on its own.
// Look a document up by number with ListFederalRegisterDocuments and DocumentNumber.
//
// full_text is served only when named in the shape, and neither ShapeFederalRegisterMinimal nor ShapeFederalRegisterComprehensive includes it, since it can run to several MB.
// Alerts on this resource use query type "federal_register" and deliver alerts.federal_register.match.

// ListFederalRegisterDocumentsOptions filters /api/federal_register/.
type ListFederalRegisterDocumentsOptions struct {
	ListOptions

	// Search is ranked full-text search over the title, abstract and action; wrap it in double quotes for a phrase.
	Search string

	// DocumentNumber is the exact FR document number (for example "2016-31922"); join several with "|".
	// A number the Federal Register reused before 2016 matches every document that carries it.
	DocumentNumber string

	// Type is "Notice", "Rule", "Proposed Rule", "Presidential Document", "Correction", "Sunshine Act Document" or "Uncategorized Document", case-insensitive; join several with "|".
	// Documents published before 2008 are mostly "Uncategorized Document".
	Type string

	// Agency is a Tango organization: a name, abbreviation, code or organization key (for example "EPA").
	// It matches a document when any agency it lists falls within that organization, so a department includes its sub-agencies; join several with "|".
	Agency string
	// FRAgency is the Federal Register's own agency slug (for example "environmental-protection-agency"); join several with "|".
	FRAgency string

	// Date bounds, inclusive (wire format: YYYY-MM-DD).
	PublicationDateAfter  string
	PublicationDateBefore string
	EffectiveOnAfter      string
	EffectiveOnBefore     string
	CommentsCloseOnAfter  string
	CommentsCloseOnBefore string

	// CommentsOpen keeps documents whose comment period closes today or later (true) or has already closed (false).
	// Documents with no comment deadline match neither. It is a pointer so that false reaches the server as a filter value.
	CommentsOpen *bool

	// CFRTitle is a CFR title number (for example "40").
	CFRTitle string
	// CFRPart is a CFR part number (for example "52"). It requires CFRTitle, and both must match the same CFR reference.
	CFRPart string

	// Significant keeps documents that are (true) or are not (false) significant under Executive Order 12866.
	Significant *bool

	// RIN is a Regulation Identifier Number (for example "2060-AV16"); join several with "|".
	RIN string
	// ExecutiveOrderNumber is the exact executive order number.
	ExecutiveOrderNumber string

	// Ordering is one of publication_date, effective_on, comments_close_on, document_number or rank; prefix "-" for descending.
	// The API default is "-publication_date". "rank" requires a non-empty Search.
	Ordering string

	Extra map[string]any
}

func (o *ListFederalRegisterDocumentsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "document_number", o.DocumentNumber)
	setIfNotEmpty(q, "type", o.Type)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "fr_agency", o.FRAgency)
	setIfNotEmpty(q, "publication_date_after", o.PublicationDateAfter)
	setIfNotEmpty(q, "publication_date_before", o.PublicationDateBefore)
	setIfNotEmpty(q, "effective_on_after", o.EffectiveOnAfter)
	setIfNotEmpty(q, "effective_on_before", o.EffectiveOnBefore)
	setIfNotEmpty(q, "comments_close_on_after", o.CommentsCloseOnAfter)
	setIfNotEmpty(q, "comments_close_on_before", o.CommentsCloseOnBefore)
	setIfNotNilBool(q, "comments_open", o.CommentsOpen)
	setIfNotEmpty(q, "cfr_title", o.CFRTitle)
	setIfNotEmpty(q, "cfr_part", o.CFRPart)
	setIfNotNilBool(q, "significant", o.Significant)
	setIfNotEmpty(q, "rin", o.RIN)
	setIfNotEmpty(q, "executive_order_number", o.ExecutiveOrderNumber)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListFederalRegisterDocuments queries /api/federal_register/.
func (c *Client) ListFederalRegisterDocuments(ctx context.Context, opts *ListFederalRegisterDocumentsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/federal_register/", opts.toQuery())
}

// IterateFederalRegisterDocuments walks every Federal Register document matching opts.
func (c *Client) IterateFederalRegisterDocuments(ctx context.Context, opts *ListFederalRegisterDocumentsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListFederalRegisterDocumentsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListFederalRegisterDocuments(ctx, &next)
		},
	}
}

// GetFederalRegisterDocument fetches one Federal Register document by its uuid (/api/federal_register/{uuid}/).
//
// It takes the uuid, not the document_number, which is not unique on its own; look a document up by number with ListFederalRegisterDocuments and DocumentNumber.
// Name full_text in the shape to get the document's plain text, which can run to several MB:
//
//	opts := &GetEntityOptions{Shape: "uuid,document_number,title,full_text"}
func (c *Client) GetFederalRegisterDocument(ctx context.Context, documentUUID string, opts *GetEntityOptions) (Record, error) {
	if documentUUID == "" {
		return nil, &ValidationError{&APIError{Message: "Federal Register document uuid is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/federal_register/"+pathEscape(documentUUID)+"/", opts.toQuery())
}
