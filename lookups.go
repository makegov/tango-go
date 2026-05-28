package tango

import (
	"context"
	"net/url"
	"strconv"
)

// ListAgenciesOptions filters /api/agencies/.
type ListAgenciesOptions struct {
	Page   int
	Limit  int
	Search string
}

// ListAgencies queries /api/agencies/.
func (c *Client) ListAgencies(ctx context.Context, opts *ListAgenciesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		setIfNonZeroInt(q, "page", opts.Page)
		if opts.Limit > 0 {
			limit := opts.Limit
			if limit > 100 {
				limit = 100
			}
			q.Set("limit", strconv.Itoa(limit))
		}
		setIfNotEmpty(q, "search", opts.Search)
	}
	return listGeneric[Record](ctx, c, "/api/agencies/", q)
}

// GetAgency fetches a single agency by code (e.g. "9700" for DoD).
//
// Returns a typed *AgencyRecord (v1.0.0 — breaking change from v0.1.0, which
// returned Record). Use the named fields when present, or AgencyRecord.Extra
// for forward-compatible fields the server adds that aren't in the typed
// surface yet.
func (c *Client) GetAgency(ctx context.Context, code string) (*AgencyRecord, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "agency code is required"}}
	}
	return getGeneric[*AgencyRecord](ctx, c, "/api/agencies/"+pathEscape(code)+"/", nil)
}

// ListOrganizationsOptions filters /api/organizations/.
type ListOrganizationsOptions struct {
	ListOptions
	Search          string
	Type            string
	Level           string
	CGAC            string
	Parent          string
	IncludeInactive *bool
}

// ListOrganizations queries /api/organizations/.
func (c *Client) ListOrganizations(ctx context.Context, opts *ListOrganizationsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.ListOptions.applyTo(q)
		setIfNotEmpty(q, "search", opts.Search)
		setIfNotEmpty(q, "type", opts.Type)
		setIfNotEmpty(q, "level", opts.Level)
		setIfNotEmpty(q, "cgac", opts.CGAC)
		setIfNotEmpty(q, "parent", opts.Parent)
		setIfNotNilBool(q, "include_inactive", opts.IncludeInactive)
	}
	return listGeneric[Record](ctx, c, "/api/organizations/", q)
}

// GetOrganization fetches a single organization by key.
func (c *Client) GetOrganization(ctx context.Context, key string) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "organization key is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/organizations/"+pathEscape(key)+"/", nil)
}

// ListNAICSOptions filters /api/naics/.
type ListNAICSOptions struct {
	ListOptions
	Search           string
	RevenueLimit     string
	EmployeeLimit    string
	RevenueLimitGte  string
	RevenueLimitLte  string
	EmployeeLimitGte string
	EmployeeLimitLte string
}

// ListNAICS queries /api/naics/.
func (c *Client) ListNAICS(ctx context.Context, opts *ListNAICSOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.ListOptions.applyTo(q)
		setIfNotEmpty(q, "search", opts.Search)
		setIfNotEmpty(q, "revenue_limit", opts.RevenueLimit)
		setIfNotEmpty(q, "employee_limit", opts.EmployeeLimit)
		setIfNotEmpty(q, "revenue_limit_gte", opts.RevenueLimitGte)
		setIfNotEmpty(q, "revenue_limit_lte", opts.RevenueLimitLte)
		setIfNotEmpty(q, "employee_limit_gte", opts.EmployeeLimitGte)
		setIfNotEmpty(q, "employee_limit_lte", opts.EmployeeLimitLte)
	}
	return listGeneric[Record](ctx, c, "/api/naics/", q)
}

// GetNAICS fetches a single NAICS by code.
func (c *Client) GetNAICS(ctx context.Context, code string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "NAICS code is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/naics/"+pathEscape(code)+"/", nil)
}

// ListPSC queries /api/psc/.
func (c *Client) ListPSC(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/psc/", q)
}

// GetPSC fetches a single PSC by code.
func (c *Client) GetPSC(ctx context.Context, code string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "PSC code is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/psc/"+pathEscape(code)+"/", nil)
}

// ListSubawardsOptions filters /api/subawards/.
type ListSubawardsOptions struct {
	ListOptions
	AwardKey       string
	PrimeUEI       string
	SubUEI         string
	AwardingAgency string
	FundingAgency  string
	FiscalYear     string
	FiscalYearGte  string
	FiscalYearLte  string
	Recipient      string
	// Ordering must be "last_modified_date" or "-last_modified_date";
	// the server rejects others (tango#2254).
	Ordering string
}

// ListSubawards queries /api/subawards/.
func (c *Client) ListSubawards(ctx context.Context, opts *ListSubawardsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.ListOptions.applyTo(q)
		setIfNotEmpty(q, "award_key", opts.AwardKey)
		setIfNotEmpty(q, "prime_uei", opts.PrimeUEI)
		setIfNotEmpty(q, "sub_uei", opts.SubUEI)
		setIfNotEmpty(q, "awarding_agency", opts.AwardingAgency)
		setIfNotEmpty(q, "funding_agency", opts.FundingAgency)
		setIfNotEmpty(q, "fiscal_year", opts.FiscalYear)
		setIfNotEmpty(q, "fiscal_year_gte", opts.FiscalYearGte)
		setIfNotEmpty(q, "fiscal_year_lte", opts.FiscalYearLte)
		setIfNotEmpty(q, "recipient", opts.Recipient)
		setIfNotEmpty(q, "ordering", opts.Ordering)
	}
	return listGeneric[Record](ctx, c, "/api/subawards/", q)
}

// GetSubaward fetches a single subaward record by its key
// (/api/subawards/{key}/).
func (c *Client) GetSubaward(ctx context.Context, key string, opts *ListOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "subaward key is required"}}
	}
	q := url.Values{}
	opts.applyTo(q)
	return getGeneric[Record](ctx, c, "/api/subawards/"+pathEscape(key)+"/", q)
}

// GetVersion returns the API version metadata.
func (c *Client) GetVersion(ctx context.Context) (Record, error) {
	return getGeneric[Record](ctx, c, "/api/version/", nil)
}
