package tango

import (
	"context"
	"strconv"
)

// MetricsOwnerType enumerates the three owner types under which Tango
// exposes rolling-window metrics.
const (
	MetricsOwnerNAICS  = "naics"
	MetricsOwnerPSC    = "psc"
	MetricsOwnerEntity = "entity"
)

// ListMetricsOptions selects an owner (NAICS code, PSC code, or entity
// UEI) and the rolling-window parameters for the metrics endpoint.
//
// Metrics live under three distinct owner paths in the API:
//   - /api/naics/{code}/metrics/{months}/{period_grouping}/
//   - /api/psc/{code}/metrics/{months}/{period_grouping}/
//   - /api/entities/{uei}/metrics/{months}/{period_grouping}/
//
// ListMetrics is a thin dispatcher over the three typed wrappers
// (GetNAICSMetrics, GetPSCMetrics, GetEntityMetrics).
type ListMetricsOptions struct {
	// OwnerType is one of "naics", "psc", or "entity" (see the
	// MetricsOwner* constants).
	OwnerType string

	// OwnerID is the code/UEI of the owner: NAICS code, PSC code, or
	// entity UEI.
	OwnerID string

	// Months is the rolling-window length in months. Must be > 0.
	Months int

	// PeriodGrouping is the aggregation granularity for the window —
	// typically "month", "quarter", or "year". Free-text in the path;
	// path-escaped by the client.
	PeriodGrouping string
}

// GetNAICSMetrics returns rolling-window metrics for a NAICS code.
// Endpoint: GET /api/naics/{code}/metrics/{months}/{periodGrouping}/.
// Mirrors Node `getNAICSMetrics` and Python `get_naics_metrics`.
func (c *Client) GetNAICSMetrics(ctx context.Context, code string, months int, periodGrouping string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "NAICS code is required"}}
	}
	if months <= 0 {
		return nil, &ValidationError{&APIError{Message: "months must be > 0"}}
	}
	if periodGrouping == "" {
		return nil, &ValidationError{&APIError{Message: "periodGrouping is required"}}
	}
	path := "/api/naics/" + pathEscape(code) + "/metrics/" + strconv.Itoa(months) + "/" + pathEscape(periodGrouping) + "/"
	return getGeneric[Record](ctx, c, path, nil)
}

// GetPSCMetrics returns rolling-window metrics for a PSC code.
// Endpoint: GET /api/psc/{code}/metrics/{months}/{periodGrouping}/.
// Mirrors Node `getPSCMetrics` and Python `get_psc_metrics`.
func (c *Client) GetPSCMetrics(ctx context.Context, code string, months int, periodGrouping string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "PSC code is required"}}
	}
	if months <= 0 {
		return nil, &ValidationError{&APIError{Message: "months must be > 0"}}
	}
	if periodGrouping == "" {
		return nil, &ValidationError{&APIError{Message: "periodGrouping is required"}}
	}
	path := "/api/psc/" + pathEscape(code) + "/metrics/" + strconv.Itoa(months) + "/" + pathEscape(periodGrouping) + "/"
	return getGeneric[Record](ctx, c, path, nil)
}

// GetEntityMetrics returns rolling-window metrics for an entity (by
// UEI). Endpoint: GET /api/entities/{uei}/metrics/{months}/{periodGrouping}/.
// Mirrors Node `getEntityMetrics` and Python `get_entity_metrics`.
func (c *Client) GetEntityMetrics(ctx context.Context, uei string, months int, periodGrouping string) (Record, error) {
	if uei == "" {
		return nil, &ValidationError{&APIError{Message: "UEI is required"}}
	}
	if months <= 0 {
		return nil, &ValidationError{&APIError{Message: "months must be > 0"}}
	}
	if periodGrouping == "" {
		return nil, &ValidationError{&APIError{Message: "periodGrouping is required"}}
	}
	path := "/api/entities/" + pathEscape(uei) + "/metrics/" + strconv.Itoa(months) + "/" + pathEscape(periodGrouping) + "/"
	return getGeneric[Record](ctx, c, path, nil)
}

// ListMetrics is a convenience dispatcher that routes to one of
// GetNAICSMetrics, GetPSCMetrics, or GetEntityMetrics based on
// opts.OwnerType. Mirrors Node `listMetrics`.
//
// Returns *ValidationError on unknown OwnerType, empty OwnerID,
// non-positive Months, or empty PeriodGrouping.
func (c *Client) ListMetrics(ctx context.Context, opts ListMetricsOptions) (Record, error) {
	if opts.OwnerID == "" {
		return nil, &ValidationError{&APIError{Message: "OwnerID is required"}}
	}
	if opts.Months <= 0 {
		return nil, &ValidationError{&APIError{Message: "Months must be > 0"}}
	}
	if opts.PeriodGrouping == "" {
		return nil, &ValidationError{&APIError{Message: "PeriodGrouping is required"}}
	}
	switch opts.OwnerType {
	case MetricsOwnerNAICS:
		return c.GetNAICSMetrics(ctx, opts.OwnerID, opts.Months, opts.PeriodGrouping)
	case MetricsOwnerPSC:
		return c.GetPSCMetrics(ctx, opts.OwnerID, opts.Months, opts.PeriodGrouping)
	case MetricsOwnerEntity:
		return c.GetEntityMetrics(ctx, opts.OwnerID, opts.Months, opts.PeriodGrouping)
	default:
		return nil, &ValidationError{&APIError{Message: "OwnerType must be one of: naics, psc, entity"}}
	}
}
