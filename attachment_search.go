package tango

import (
	"context"
	"net/url"
)

// SearchOpportunityAttachmentsOptions controls the
// /api/opportunities/attachment-search/ endpoint — semantic search over
// the extracted text of opportunity attachments (SOWs, PWSs, J&As, etc.).
//
// Q is required. TopK and IncludeExtractedText are optional and mirror
// the Node + Python options of the same names.
type SearchOpportunityAttachmentsOptions struct {
	// Q is the natural-language query. Required; client-side validated
	// (empty Q raises *ValidationError before any network call).
	Q string

	// TopK is the maximum number of matches to return. 0 means
	// "use the server default".
	TopK int

	// IncludeExtractedText, when true, returns the matched attachment
	// text alongside the metadata. False / unset omits the text payload
	// to keep responses small.
	IncludeExtractedText bool
}

// SearchOpportunityAttachments runs a semantic search over opportunity
// attachment extracted text. Endpoint:
// GET /api/opportunities/attachment-search/. Mirrors Node
// `searchOpportunityAttachments` and Python
// `search_opportunity_attachments`.
//
// Returns *ValidationError if opts.Q is empty.
func (c *Client) SearchOpportunityAttachments(ctx context.Context, opts SearchOpportunityAttachmentsOptions) (Record, error) {
	if opts.Q == "" {
		return nil, &ValidationError{&APIError{Message: "searchOpportunityAttachments: 'q' is required"}}
	}
	q := url.Values{}
	setIfNotEmpty(q, "q", opts.Q)
	setIfNonZeroInt(q, "top_k", opts.TopK)
	if opts.IncludeExtractedText {
		// Match the Node SDK wire format: send the boolean as a string.
		q.Set("include_extracted_text", "true")
	}
	return getGeneric[Record](ctx, c, "/api/opportunities/attachment-search/", q)
}
