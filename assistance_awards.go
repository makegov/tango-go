package tango

import (
	"context"
	"net/url"
	"strconv"
)

// Financial-assistance awards — /api/assistance/{grants,loans,direct_payments,other}/.
//
// These are USAspending financial-assistance awards: money an agency has already obligated to a recipient, split into four endpoints by the GSDM assistance type of the award's latest action.
// Grants (F001, F002) are not the grant opportunities on /api/grants/, which ListGrants serves; an award's funding_opportunity_number names the opportunity it was made under.
// The four endpoints take the same filters and shapes, so they share ListAssistanceAwardsOptions and AssistanceAwardRecord.
//
// An award lives under one endpoint at a time: if its latest action's assistance type moves to another group, the award moves with the same key and its old URL returns 404.
// A retired key still resolves, to the award that carries it in former_award_keys, on the detail endpoint and both sub-routes.

const (
	assistanceGrantsPath         = "/api/assistance/grants/"
	assistanceLoansPath          = "/api/assistance/loans/"
	assistanceDirectPaymentsPath = "/api/assistance/direct_payments/"
	assistanceOtherPath          = "/api/assistance/other/"
)

// ListAssistanceAwardsOptions filters the four assistance award lists: ListAssistanceGrants, ListAssistanceLoans, ListAssistanceDirectPayments and ListAssistanceOther.
//
// The lists page by cursor only, so leave Page unset and let the iterator (or the response's Cursor) drive paging.
// The API rejects a parameter it does not recognize with a 400, so anything passed through Extra must be a real filter on these endpoints.
type ListAssistanceAwardsOptions struct {
	ListOptions

	// Joiner is the separator used when Flat is true (e.g. "." or "__").
	// Empty means "use the server default".
	Joiner string

	// Award identifiers. Each matches exactly; join several with "|" for OR.
	// Key matches the award key the detail endpoint takes.
	Key string
	// FAIN is the Federal Award Identification Number, carried by non-aggregate records.
	FAIN string
	// URI is the Unique Record Identifier, carried by aggregate records.
	URI string
	// FundingOpportunityNumber is the Grants.gov funding opportunity the award was made under.
	FundingOpportunityNumber string

	// UEI matches the recipient's UEI exactly; join several with "|" for OR.
	UEI string
	// Recipient matches a recipient name or UEI, including recipients without a UEI.
	// A broad term covers only the recipients closest to it by name, so filter on UEI to pin one; join several with "|" for OR.
	Recipient string

	// AwardingAgency and FundingAgency take an agency code, name or abbreviation; join several with "|" for OR.
	AwardingAgency string
	FundingAgency  string

	// CFDANumber matches the award's primary assistance listing number (e.g. "93.778") only; its other programs are on the listings sub-route.
	CFDANumber string
	// AssistanceType is a GSDM assistance type code ("F001"–"F010"); each endpoint holds only its own codes, so another code matches nothing.
	AssistanceType string
	// RecordType is "1" (aggregate), "2" (non-aggregate) or "3" (non-aggregate to an individual, PII-redacted).
	RecordType string
	// BusinessFundsIndicator is "NON" (not Recovery Act) or "REC" (Recovery Act).
	BusinessFundsIndicator string
	// DisasterEmergencyFundCode matches a whole code within the award's DEFC list (e.g. "L"), so "L" does not match inside another code.
	DisasterEmergencyFundCode string

	// Date bounds (wire format: YYYY-MM-DD).
	AwardDate       string
	AwardDateGte    string
	AwardDateLte    string
	PopStartDateGte string
	PopStartDateLte string
	PopEndDateGte   string
	PopEndDateLte   string

	// Fiscal-year bounds (YYYY).
	FiscalYear    string
	FiscalYearGte string
	FiscalYearLte string

	// Place of performance: a state code ("VA"), an ISO 3166 alpha-3 country code ("USA") and a five-digit ZIP code; join several with "|" for OR.
	PopState   string
	PopCountry string
	PopZip     string

	// Obligated amount bounds (USD).
	// To compare dollars across the four endpoints, order by pragmatic_obligations instead: it is the obligation for grants and payments and the subsidy cost for loans.
	ObligatedGte string
	ObligatedLte string

	// Search is a full-text query over identifiers, recipient names, listing titles and descriptions.
	Search string
	// Ordering is award_date, obligated or pragmatic_obligations, with a "-" prefix for descending.
	// The default is -award_date, newest first.
	Ordering string

	// Extra is an escape hatch for filter keys not first-classed here.
	Extra map[string]any
}

func (o *ListAssistanceAwardsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "joiner", o.Joiner)
	setIfNotEmpty(q, "key", o.Key)
	setIfNotEmpty(q, "fain", o.FAIN)
	setIfNotEmpty(q, "uri", o.URI)
	setIfNotEmpty(q, "funding_opportunity_number", o.FundingOpportunityNumber)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "recipient", o.Recipient)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "cfda_number", o.CFDANumber)
	setIfNotEmpty(q, "assistance_type", o.AssistanceType)
	setIfNotEmpty(q, "record_type", o.RecordType)
	setIfNotEmpty(q, "business_funds_indicator", o.BusinessFundsIndicator)
	setIfNotEmpty(q, "disaster_emergency_fund_code", o.DisasterEmergencyFundCode)
	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "pop_end_date_gte", o.PopEndDateGte)
	setIfNotEmpty(q, "pop_end_date_lte", o.PopEndDateLte)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "pop_state", o.PopState)
	setIfNotEmpty(q, "pop_country", o.PopCountry)
	setIfNotEmpty(q, "pop_zip", o.PopZip)
	setIfNotEmpty(q, "obligated_gte", o.ObligatedGte)
	setIfNotEmpty(q, "obligated_lte", o.ObligatedLte)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// AssistanceSubresourceOptions pages the transactions and listings sub-routes of an assistance award.
// Those routes take page and limit and nothing else: a shape or filter is rejected with a 400, which is why this is not ListOptions.
type AssistanceSubresourceOptions struct {
	// Page is 1-based; 0 means the first page.
	Page int
	// Limit is the page size; 0 means the server default (100 for transactions, max 500; 10 for listings, max 100).
	Limit int
}

func (o *AssistanceSubresourceOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	setIfNonZeroInt(q, "limit", o.Limit)
	return q
}

func (c *Client) listAssistanceAwards(ctx context.Context, path string, opts *ListAssistanceAwardsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, path, opts.toQuery())
}

func (c *Client) iterateAssistanceAwards(ctx context.Context, path string, opts *ListAssistanceAwardsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListAssistanceAwardsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.listAssistanceAwards(ctx, path, &next)
		},
	}
}

func (c *Client) getAssistanceAward(ctx context.Context, path, noun, key string, opts *GetEntityOptions) (*AssistanceAwardRecord, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: noun + " key is required"}}
	}
	return getGeneric[*AssistanceAwardRecord](ctx, c, path+pathEscape(key)+"/", opts.toQuery())
}

func (c *Client) listAssistanceTransactions(ctx context.Context, path, noun, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceTransaction], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: noun + " key is required"}}
	}
	return listGeneric[AssistanceTransaction](ctx, c, path+pathEscape(key)+"/transactions/", opts.toQuery())
}

func (c *Client) listAssistanceAwardListings(ctx context.Context, path, noun, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceAwardListing], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: noun + " key is required"}}
	}
	return listGeneric[AssistanceAwardListing](ctx, c, path+pathEscape(key)+"/listings/", opts.toQuery())
}

// ListAssistanceGrants queries /api/assistance/grants/ — grant and cooperative-agreement awards (F001, F002).
// These are awards already made; grant opportunities are ListGrants.
func (c *Client) ListAssistanceGrants(ctx context.Context, opts *ListAssistanceAwardsOptions) (*PaginatedResponse[Record], error) {
	return c.listAssistanceAwards(ctx, assistanceGrantsPath, opts)
}

// IterateAssistanceGrants walks every grant award matching opts, following the server's cursor.
func (c *Client) IterateAssistanceGrants(ctx context.Context, opts *ListAssistanceAwardsOptions) *Iterator[Record] {
	return c.iterateAssistanceAwards(ctx, assistanceGrantsPath, opts)
}

// GetAssistanceGrant fetches one grant award by key (/api/assistance/grants/{key}/).
// A retired key returns the award that now carries it, with its current key in the body.
func (c *Client) GetAssistanceGrant(ctx context.Context, key string, opts *GetEntityOptions) (*AssistanceAwardRecord, error) {
	return c.getAssistanceAward(ctx, assistanceGrantsPath, "assistance grant", key, opts)
}

// ListAssistanceGrantTransactions lists one grant award's transactions, latest first (/api/assistance/grants/{key}/transactions/).
func (c *Client) ListAssistanceGrantTransactions(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceTransaction], error) {
	return c.listAssistanceTransactions(ctx, assistanceGrantsPath, "assistance grant", key, opts)
}

// ListAssistanceGrantListings lists one grant award's money by assistance listing, primary listing first (/api/assistance/grants/{key}/listings/).
func (c *Client) ListAssistanceGrantListings(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceAwardListing], error) {
	return c.listAssistanceAwardListings(ctx, assistanceGrantsPath, "assistance grant", key, opts)
}

// ListAssistanceLoans queries /api/assistance/loans/ — direct loans and loan guarantees (F003, F004).
// On a loan, face_value_of_loan is the amount lent or guaranteed and pragmatic_obligations holds loan_subsidy_cost, the estimated cost to the government.
func (c *Client) ListAssistanceLoans(ctx context.Context, opts *ListAssistanceAwardsOptions) (*PaginatedResponse[Record], error) {
	return c.listAssistanceAwards(ctx, assistanceLoansPath, opts)
}

// IterateAssistanceLoans walks every loan award matching opts, following the server's cursor.
func (c *Client) IterateAssistanceLoans(ctx context.Context, opts *ListAssistanceAwardsOptions) *Iterator[Record] {
	return c.iterateAssistanceAwards(ctx, assistanceLoansPath, opts)
}

// GetAssistanceLoan fetches one loan award by key (/api/assistance/loans/{key}/).
// A retired key returns the award that now carries it, with its current key in the body.
func (c *Client) GetAssistanceLoan(ctx context.Context, key string, opts *GetEntityOptions) (*AssistanceAwardRecord, error) {
	return c.getAssistanceAward(ctx, assistanceLoansPath, "assistance loan", key, opts)
}

// ListAssistanceLoanTransactions lists one loan award's transactions, latest first (/api/assistance/loans/{key}/transactions/).
func (c *Client) ListAssistanceLoanTransactions(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceTransaction], error) {
	return c.listAssistanceTransactions(ctx, assistanceLoansPath, "assistance loan", key, opts)
}

// ListAssistanceLoanListings lists one loan award's money by assistance listing, primary listing first (/api/assistance/loans/{key}/listings/).
func (c *Client) ListAssistanceLoanListings(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceAwardListing], error) {
	return c.listAssistanceAwardListings(ctx, assistanceLoansPath, "assistance loan", key, opts)
}

// ListAssistanceDirectPayments queries /api/assistance/direct_payments/ — direct payments for a specified use or with unrestricted use (F006, F007).
// Most rows are aggregate (record type "1") or PII-redacted (record type "3") records with a null recipient; RecordType "2" keeps payments to an identified recipient.
func (c *Client) ListAssistanceDirectPayments(ctx context.Context, opts *ListAssistanceAwardsOptions) (*PaginatedResponse[Record], error) {
	return c.listAssistanceAwards(ctx, assistanceDirectPaymentsPath, opts)
}

// IterateAssistanceDirectPayments walks every direct-payment award matching opts, following the server's cursor.
func (c *Client) IterateAssistanceDirectPayments(ctx context.Context, opts *ListAssistanceAwardsOptions) *Iterator[Record] {
	return c.iterateAssistanceAwards(ctx, assistanceDirectPaymentsPath, opts)
}

// GetAssistanceDirectPayment fetches one direct-payment award by key (/api/assistance/direct_payments/{key}/).
// A retired key returns the award that now carries it, with its current key in the body.
func (c *Client) GetAssistanceDirectPayment(ctx context.Context, key string, opts *GetEntityOptions) (*AssistanceAwardRecord, error) {
	return c.getAssistanceAward(ctx, assistanceDirectPaymentsPath, "assistance direct payment", key, opts)
}

// ListAssistanceDirectPaymentTransactions lists one direct-payment award's transactions, latest first (/api/assistance/direct_payments/{key}/transactions/).
func (c *Client) ListAssistanceDirectPaymentTransactions(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceTransaction], error) {
	return c.listAssistanceTransactions(ctx, assistanceDirectPaymentsPath, "assistance direct payment", key, opts)
}

// ListAssistanceDirectPaymentListings lists one direct-payment award's money by assistance listing, primary listing first (/api/assistance/direct_payments/{key}/listings/).
func (c *Client) ListAssistanceDirectPaymentListings(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceAwardListing], error) {
	return c.listAssistanceAwardListings(ctx, assistanceDirectPaymentsPath, "assistance direct payment", key, opts)
}

// ListAssistanceOther queries /api/assistance/other/ — insurance, asset forfeiture, property sales and other assistance (F005, F008, F009, F010), plus awards with no assistance type ("-1").
func (c *Client) ListAssistanceOther(ctx context.Context, opts *ListAssistanceAwardsOptions) (*PaginatedResponse[Record], error) {
	return c.listAssistanceAwards(ctx, assistanceOtherPath, opts)
}

// IterateAssistanceOther walks every other-assistance award matching opts, following the server's cursor.
func (c *Client) IterateAssistanceOther(ctx context.Context, opts *ListAssistanceAwardsOptions) *Iterator[Record] {
	return c.iterateAssistanceAwards(ctx, assistanceOtherPath, opts)
}

// GetAssistanceOther fetches one other-assistance award by key (/api/assistance/other/{key}/).
// A retired key returns the award that now carries it, with its current key in the body.
func (c *Client) GetAssistanceOther(ctx context.Context, key string, opts *GetEntityOptions) (*AssistanceAwardRecord, error) {
	return c.getAssistanceAward(ctx, assistanceOtherPath, "assistance other", key, opts)
}

// ListAssistanceOtherTransactions lists one other-assistance award's transactions, latest first (/api/assistance/other/{key}/transactions/).
func (c *Client) ListAssistanceOtherTransactions(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceTransaction], error) {
	return c.listAssistanceTransactions(ctx, assistanceOtherPath, "assistance other", key, opts)
}

// ListAssistanceOtherListings lists one other-assistance award's money by assistance listing, primary listing first (/api/assistance/other/{key}/listings/).
func (c *Client) ListAssistanceOtherListings(ctx context.Context, key string, opts *AssistanceSubresourceOptions) (*PaginatedResponse[AssistanceAwardListing], error) {
	return c.listAssistanceAwardListings(ctx, assistanceOtherPath, "assistance other", key, opts)
}
