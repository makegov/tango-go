package tango

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

// GSA eBuy — /api/ebuy/requests/ and /api/ebuy/access/ (Tango API 5.1.0).
//
// eBuy requests (RFQs, RFPs and RFIs) are scoped to the caller: you see only requests posted under the GSA schedule contracts your account is linked to.
// The list and detail endpoints require the Pro tier or above; below it they return 403.
// With no linked contract the list is empty rather than an error, so call GetEbuyAccess to tell "no access" from "no matches".
//
// Data caveats:
//
//   - status is frozen at the last-seen state. eBuy only carries currently-active requests, so a request that closes stops appearing rather than getting a final row; "Open" means "open the last time it was seen". Read last_seen for staleness.
//   - contract_number is accepted as a filter but never returned in any payload.
//   - buyer_agency_code and some other buyer and contact fields are sparse on older requests.

// ListEbuyRequestsOptions filters /api/ebuy/requests/.
//
// Every string filter except the date bounds accepts "|" for OR.
type ListEbuyRequestsOptions struct {
	ListOptions

	// Search is full text over title, description, reference number, request id and attachment text; results rank by relevance unless Ordering is set.
	Search string

	// RfqID is the exact eBuy request id (for example "RFQ1835158").
	RfqID string
	// ReferenceNumber is the buyer's own solicitation number; matching ignores dashes.
	ReferenceNumber string

	// RequestType is "RFQ", "RFP" or "RFI".
	RequestType string
	// Status is "Open" or "Cancelled", frozen at the last-seen state (see the package notes above).
	Status   string
	Sin      string
	Schedule string
	// BuyerAgency matches the buying agency name exactly as fed.
	BuyerAgency string
	// Agency matches an agency by name, abbreviation or code, including its sub-agencies and offices.
	Agency string

	// ContractNumber narrows results to one of your own linked contracts; a contract you do not hold returns nothing rather than an error.
	ContractNumber string

	// Date bounds, inclusive (wire format: YYYY-MM-DD).
	IssueDateAfter  string
	IssueDateBefore string
	CloseDateAfter  string
	CloseDateBefore string

	// Ordering is one of issue_date, close_date, last_seen or modified; prefix "-" for descending. The API default is "-issue_date".
	Ordering string

	Extra map[string]any
}

func (o *ListEbuyRequestsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "rfq_id", o.RfqID)
	setIfNotEmpty(q, "reference_number", o.ReferenceNumber)
	setIfNotEmpty(q, "request_type", o.RequestType)
	setIfNotEmpty(q, "status", o.Status)
	setIfNotEmpty(q, "sin", o.Sin)
	setIfNotEmpty(q, "schedule", o.Schedule)
	setIfNotEmpty(q, "buyer_agency", o.BuyerAgency)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "contract_number", o.ContractNumber)
	setIfNotEmpty(q, "issue_date_after", o.IssueDateAfter)
	setIfNotEmpty(q, "issue_date_before", o.IssueDateBefore)
	setIfNotEmpty(q, "close_date_after", o.CloseDateAfter)
	setIfNotEmpty(q, "close_date_before", o.CloseDateBefore)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListEbuyRequests queries /api/ebuy/requests/.
//
// An empty page can mean either no matches or no linked contract; GetEbuyAccess distinguishes the two.
func (c *Client) ListEbuyRequests(ctx context.Context, opts *ListEbuyRequestsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/ebuy/requests/", opts.toQuery())
}

// IterateEbuyRequests walks every eBuy request matching opts.
func (c *Client) IterateEbuyRequests(ctx context.Context, opts *ListEbuyRequestsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListEbuyRequestsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListEbuyRequests(ctx, &next)
		},
	}
}

// GetEbuyRequest fetches one eBuy request by its request id (/api/ebuy/requests/{rfq_id}/).
//
// A request outside your scope returns *NotFoundError, the same as an id that does not exist.
// The default shape carries every field plus organization(*) and attachments(*).
func (c *Client) GetEbuyRequest(ctx context.Context, rfqID string, opts *GetEntityOptions) (Record, error) {
	if rfqID == "" {
		return nil, &ValidationError{&APIError{Message: "eBuy request id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/ebuy/requests/"+pathEscape(rfqID)+"/", opts.toQuery())
}

// EbuyLinkAttachmentError is returned by GetEbuyAttachmentURL when the attachment is an outbound link (is_link=true) rather than a stored document.
// URL is the link target.
type EbuyLinkAttachmentError struct {
	*ValidationError
	URL string
}

// Error implements the error interface.
func (e *EbuyLinkAttachmentError) Error() string { return e.ValidationError.Error() }

// Unwrap returns the embedded *ValidationError so errors.As can reach it and the base *APIError.
func (e *EbuyLinkAttachmentError) Unwrap() error { return e.ValidationError }

// GetEbuyAttachmentURL returns a short-lived download URL for one attachment of an eBuy request (/api/ebuy/requests/{rfq_id}/attachments/{doc_seq_num}/download/).
//
// The API answers with a redirect to a presigned URL that expires after about five minutes; this method does not follow it and returns the target instead, so fetch it promptly.
// An attachment that is an outbound link returns *EbuyLinkAttachmentError carrying the link.
// A document that has not been captured yet returns *NotFoundError.
func (c *Client) GetEbuyAttachmentURL(ctx context.Context, rfqID string, docSeqNum int) (string, error) {
	if rfqID == "" {
		return "", &ValidationError{&APIError{Message: "eBuy request id is required"}}
	}
	if docSeqNum < 0 {
		return "", &ValidationError{&APIError{Message: "eBuy attachment doc_seq_num must be non-negative"}}
	}
	path := "/api/ebuy/requests/" + pathEscape(rfqID) + "/attachments/" + strconv.Itoa(docSeqNum) + "/download/"
	raw, err := c.do(ctx, requestSpec{method: "GET", path: path, noRedirect: true})
	if err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			if body, ok := ve.ResponseData.(map[string]any); ok {
				if link, ok := body["url"].(string); ok && link != "" {
					return "", &EbuyLinkAttachmentError{ValidationError: ve, URL: link}
				}
			}
		}
		return "", err
	}
	return string(raw), nil
}

// EbuyAccess is the response from GetEbuyAccess.
type EbuyAccess struct {
	// Enabled is true when the caller is on the Pro tier or above and holds at least one linked contract.
	Enabled bool `json:"enabled"`
	// Reason explains a false Enabled: "tier_required" or "no_contract_grant" (tier_required wins when both apply). Nil when Enabled is true.
	Reason *string `json:"reason"`
	// Contracts are the caller's own linked contracts, sorted.
	Contracts []string `json:"contracts"`

	// Extra captures any forward-compatible fields the server adds. Not serialized on the way out.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *EbuyAccess) UnmarshalJSON(data []byte) error {
	type alias EbuyAccess
	return unmarshalWithExtra(data, (*alias)(a))
}

// GetEbuyAccess reports whether the caller can read eBuy requests (/api/ebuy/access/).
func (c *Client) GetEbuyAccess(ctx context.Context) (*EbuyAccess, error) {
	return getGeneric[*EbuyAccess](ctx, c, "/api/ebuy/access/", nil)
}
