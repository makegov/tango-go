package tango

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"
)

// Record is the default result type for list endpoints. It mirrors the
// permissive map[string]any shape used by the API — fields vary based on
// the requested shape, so callers either work with the map directly or
// json.Unmarshal it into their own struct.
type Record = map[string]any

// PaginatedResponse is the standard envelope for list endpoints.
//
// For keyset/cursor endpoints (e.g. /api/idvs/), the next cursor is
// extracted from Next for convenience and exposed on Cursor.
type PaginatedResponse[T any] struct {
	// Count is the total number of items across all pages, when the
	// server reports it. For some endpoints this matches len(Results).
	Count int `json:"count"`

	// Next is the URL of the next page, or "" when on the last page.
	Next string `json:"next"`

	// Previous is the URL of the previous page, or "" when on the first.
	Previous string `json:"previous"`

	// PageMetadata carries opaque page metadata for keyset endpoints.
	// The API does not currently send a page_metadata key, so this is normally nil; read Meta instead.
	PageMetadata map[string]any `json:"page_metadata,omitempty"`

	// Meta is the response-level meta block the API attached to this page, or nil when it sent none (most responses).
	// It currently carries agency-filter diagnostics: resolved_filters maps each agency filter to the organizations its "|"-separated tokens resolved to, and warnings lists notes about tokens that were dropped or matched loosely.
	// Each resolved_filters entry that resolved also carries matched_by (Tango API 5.9.0+), saying how the token matched: "key" (an organization UUID), "code" (a 3-digit CGAC or 4-digit FPDS code), "name" (the organization's name, including a department's everyday name, a spelling variant or a rename), "alias" (an abbreviation or the organization's own alias) or "fuzzy" (a looser text match, worth checking against the resolved name).
	// An entry that did not resolve has no matched_by.
	// AgencyWarnings, UnresolvedAgencyTokens and ResolvedAgencies read it for you.
	Meta map[string]any `json:"meta,omitempty"`

	// CountType is "exact" or "approximate", read from this response's X-Results-CountType header; empty when the API did not send it.
	// Past 1,000 matches the contract, IDV, opportunity and notice lists may return a query-planner estimate as Count, which can be well off.
	// It comes from each response itself, so concurrent calls on one client cannot see each other's value the way LastResponseHeaders can.
	CountType string `json:"-"`

	// Cursor is the cursor extracted from Next on keyset-paginated
	// endpoints. Pass it back via the Cursor option to fetch the next
	// page. Empty when the endpoint is page-based or no cursor is set.
	Cursor string `json:"-"`

	// Results is the actual data for this page.
	Results []T `json:"results"`
}

// rawListResponse is what the wire actually returns; we decode into this
// and then build a typed PaginatedResponse[T].
type rawListResponse struct {
	Count        json.Number     `json:"count"`
	Next         *string         `json:"next"`
	Previous     *string         `json:"previous"`
	PageMetadata map[string]any  `json:"page_metadata"`
	Meta         map[string]any  `json:"meta"`
	Results      json.RawMessage `json:"results"`
}

func decodePaginated[T any](raw []byte) (*PaginatedResponse[T], error) {
	var envelope rawListResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, &APIError{Message: "decode list response: " + err.Error()}
	}
	out := &PaginatedResponse[T]{}
	if envelope.Count != "" {
		if n, err := envelope.Count.Int64(); err == nil {
			out.Count = int(n)
		}
	}
	if envelope.Next != nil {
		out.Next = *envelope.Next
		out.Cursor = extractCursor(out.Next)
	}
	if envelope.Previous != nil {
		out.Previous = *envelope.Previous
	}
	out.PageMetadata = envelope.PageMetadata
	out.Meta = envelope.Meta

	if len(envelope.Results) > 0 {
		if err := json.Unmarshal(envelope.Results, &out.Results); err != nil {
			return nil, &APIError{Message: "decode list results: " + err.Error()}
		}
	}
	if out.Count == 0 {
		out.Count = len(out.Results)
	}
	return out, nil
}

// AgencyWarnings returns the warnings the API raised about agency filters on this request, or nil when there were none.
//
// A non-empty result means part of an agency filter did not apply, so a small or empty Results is not evidence that no such records exist.
func (p *PaginatedResponse[T]) AgencyWarnings() []string {
	if p == nil || p.Meta == nil {
		return nil
	}
	raw, ok := p.Meta["warnings"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, w := range raw {
		if s, ok := w.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// UnresolvedAgencyTokens returns the agency tokens that matched no organization, keyed by filter name; empty when everything resolved.
//
// Use it to fail loudly rather than treat a silently narrowed result set as an answer.
func (p *PaginatedResponse[T]) UnresolvedAgencyTokens() map[string][]string {
	out := map[string][]string{}
	for name, entries := range p.resolvedFilterEntries() {
		for _, entry := range entries {
			token, ok := entry["token"].(string)
			if !ok {
				continue
			}
			if resolved, present := entry["resolved"]; !present || resolved == nil {
				out[name] = append(out[name], token)
			}
		}
	}
	return out
}

// ResolvedAgencies returns the organization each agency token resolved to, keyed by filter name; empty when nothing resolved.
//
// Agency resolution is fuzzy, so a token can match an organization the caller did not intend and scope the query to that subtree.
// Checking the resolved name (and each entry's matched_by in Meta) is the only way to catch that client-side, since nothing was dropped.
func (p *PaginatedResponse[T]) ResolvedAgencies() map[string][]map[string]any {
	out := map[string][]map[string]any{}
	for name, entries := range p.resolvedFilterEntries() {
		for _, entry := range entries {
			if org, ok := entry["resolved"].(map[string]any); ok {
				out[name] = append(out[name], org)
			}
		}
	}
	return out
}

// resolvedFilterEntries reads meta.resolved_filters, skipping anything malformed.
func (p *PaginatedResponse[T]) resolvedFilterEntries() map[string][]map[string]any {
	out := map[string][]map[string]any{}
	if p == nil || p.Meta == nil {
		return out
	}
	filters, ok := p.Meta["resolved_filters"].(map[string]any)
	if !ok {
		return out
	}
	for name, v := range filters {
		list, ok := v.([]any)
		if !ok {
			continue
		}
		for _, e := range list {
			if entry, ok := e.(map[string]any); ok {
				out[name] = append(out[name], entry)
			}
		}
	}
	return out
}

func extractCursor(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}

// Iterator walks every result of a paginated list endpoint, following
// either page= or cursor= on the server's Next URL.
//
// Usage:
//
//	iter := client.IterateContracts(ctx, &tango.ListContractsOptions{
//	    AwardingAgency: "9700",
//	})
//	for iter.Next() {
//	    contract := iter.Item()
//	    // ...
//	}
//	if err := iter.Err(); err != nil { ... }
type Iterator[T any] struct {
	ctx     context.Context
	fetch   func(ctx context.Context, page int, cursor string) (*PaginatedResponse[T], error)
	page    int    // 0 means "haven't set"
	cursor  string // empty means "haven't set"
	buf     []T
	idx     int
	current T
	done    bool
	err     error
}

// Next advances the iterator and returns true if Item() is now valid.
// Returns false when the iteration is finished or an error occurred (in
// which case Err() returns the error).
func (it *Iterator[T]) Next() bool {
	if it.err != nil {
		return false
	}
	// Drain the buffer first — even after the previous fetch flagged
	// "done" (last page reached), buf may still hold un-yielded items.
	// Yielding them before checking done is what makes single-page and
	// multi-item-last-page iteration correct.
	if it.idx < len(it.buf) {
		it.current = it.buf[it.idx]
		it.idx++
		return true
	}
	if it.done {
		return false
	}
	// Need a new page.
	resp, err := it.fetch(it.ctx, it.page, it.cursor)
	if err != nil {
		it.err = err
		return false
	}
	if len(resp.Results) == 0 {
		it.done = true
		return false
	}
	it.buf = resp.Results
	it.idx = 0

	// Decide how to advance.
	if resp.Next == "" {
		it.done = true // last page; we'll still drain buf
		it.page = 0
		it.cursor = ""
	} else if resp.Cursor != "" {
		it.cursor = resp.Cursor
		it.page = 0
	} else if u, err := url.Parse(resp.Next); err == nil {
		if p := u.Query().Get("page"); p != "" {
			if n, err := parseIntStrict(p); err == nil {
				it.page = n
				it.cursor = ""
			} else {
				it.done = true
			}
		} else {
			it.done = true
		}
	} else {
		it.done = true
	}

	it.current = it.buf[it.idx]
	it.idx++
	return true
}

// Item returns the current item. Only valid after a true return from Next.
func (it *Iterator[T]) Item() T { return it.current }

// Err returns the error that ended iteration, if any.
func (it *Iterator[T]) Err() error { return it.err }

// Seq adapts the iterator to a Go 1.23+ range-over-func sequence.
// Yields (item, nil) for each result and (zero, err) once if iteration
// fails. The Next/Item/Err surface still works — Seq is additive sugar.
//
// Usage:
//
//	for record, err := range client.IterateContracts(ctx, opts).Seq() {
//	    if err != nil {
//	        return err
//	    }
//	    // use record
//	}
//
// Break out of the loop normally; the underlying iterator stops fetching
// pages as soon as the range body returns false.
func (it *Iterator[T]) Seq() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for it.Next() {
			if !yield(it.Item(), nil) {
				return
			}
		}
		if err := it.Err(); err != nil {
			var zero T
			yield(zero, err)
		}
	}
}

func parseIntStrict(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, &APIError{Message: "empty page"}
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, &APIError{Message: "invalid page: " + s}
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
