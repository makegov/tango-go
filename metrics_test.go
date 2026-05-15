package tango

import (
	"context"
	"errors"
	"testing"
)

func TestGetNAICSMetricsRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetNAICSMetrics(context.Background(), "", 12, "month")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty code, got %T: %v", err, err)
	}
}

func TestGetNAICSMetricsRequiresPositiveMonths(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetNAICSMetrics(context.Background(), "541512", 0, "month")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for months=0, got %T: %v", err, err)
	}
}

func TestGetNAICSMetricsRequiresPeriodGrouping(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetNAICSMetrics(context.Background(), "541512", 12, "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty periodGrouping, got %T: %v", err, err)
	}
}

func TestGetNAICSMetricsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetNAICSMetrics(context.Background(), "541512", 12, "month")
	assertPathContains(t, capturedURL, "/api/naics/541512/metrics/12/month/")
}

func TestGetPSCMetricsRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetPSCMetrics(context.Background(), "", 6, "quarter")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetPSCMetricsRequiresPositiveMonths(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetPSCMetrics(context.Background(), "D302", -1, "month")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetPSCMetricsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetPSCMetrics(context.Background(), "D302", 24, "year")
	assertPathContains(t, capturedURL, "/api/psc/D302/metrics/24/year/")
}

func TestGetEntityMetricsRequiresUEI(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetEntityMetrics(context.Background(), "", 12, "month")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetEntityMetricsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetEntityMetrics(context.Background(), "UEI12345", 6, "quarter")
	assertPathContains(t, capturedURL, "/api/entities/UEI12345/metrics/6/quarter/")
}

func TestListMetricsDispatchesToNAICS(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      MetricsOwnerNAICS,
		OwnerID:        "541512",
		Months:         12,
		PeriodGrouping: "month",
	})
	assertPathContains(t, capturedURL, "/api/naics/541512/metrics/")
}

func TestListMetricsDispatchesToPSC(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      MetricsOwnerPSC,
		OwnerID:        "D302",
		Months:         6,
		PeriodGrouping: "quarter",
	})
	assertPathContains(t, capturedURL, "/api/psc/D302/metrics/")
}

func TestListMetricsDispatchesToEntity(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      MetricsOwnerEntity,
		OwnerID:        "UEI99999",
		Months:         24,
		PeriodGrouping: "year",
	})
	assertPathContains(t, capturedURL, "/api/entities/UEI99999/metrics/")
}

func TestListMetricsRequiresOwnerID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      MetricsOwnerNAICS,
		Months:         12,
		PeriodGrouping: "month",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty OwnerID, got %T: %v", err, err)
	}
}

func TestListMetricsRequiresPositiveMonths(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      MetricsOwnerNAICS,
		OwnerID:        "541512",
		Months:         0,
		PeriodGrouping: "month",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for Months=0, got %T: %v", err, err)
	}
}

func TestListMetricsRequiresPeriodGrouping(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType: MetricsOwnerNAICS,
		OwnerID:   "541512",
		Months:    12,
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty PeriodGrouping, got %T: %v", err, err)
	}
}

func TestListMetricsUnknownOwnerType(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListMetrics(context.Background(), ListMetricsOptions{
		OwnerType:      "bogus",
		OwnerID:        "123",
		Months:         12,
		PeriodGrouping: "month",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for unknown OwnerType, got %T: %v", err, err)
	}
}
