package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListIDVAwardsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListIDVAwards(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListIDVAwardsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVAwards(context.Background(), "IDV-KEY-001", nil)
	assertPathContains(t, capturedURL, "/api/idvs/IDV-KEY-001/awards/")
}

func TestListIDVAwardsWithFilters(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVAwards(context.Background(), "KEY1", &ListIDVsOptions{
		AwardingAgency: "9700",
		FiscalYear:     "2024",
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"awarding_agency": "9700",
		"fiscal_year":     "2024",
	}, nil)
}

func TestListIDVChildIDVsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListIDVChildIDVs(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListIDVChildIDVsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVChildIDVs(context.Background(), "PARENT-KEY", nil)
	assertPathContains(t, capturedURL, "/api/idvs/PARENT-KEY/idvs/")
}

func TestListIDVTransactionsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListIDVTransactions(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListIDVTransactionsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVTransactions(context.Background(), "IDV-TX-KEY", nil)
	assertPathContains(t, capturedURL, "/api/idvs/IDV-TX-KEY/transactions/")
}

func TestListIDVTransactionsWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVTransactions(context.Background(), "KEY1", &ListOptions{
		Limit: 20, Cursor: "abc123",
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"limit":  "20",
		"cursor": "abc123",
	}, nil)
}

func TestListIDVLcatsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListIDVLcats(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListIDVLcatsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVLcats(context.Background(), "IDV-KEY-LCAT", nil)
	assertPathContains(t, capturedURL, "/api/idvs/IDV-KEY-LCAT/lcats/")
}

func TestListIDVLcatsWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVLcats(context.Background(), "KEY1", &EntityLcatsOptions{
		Ordering: "-labor_category",
		Search:   "senior",
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"ordering": "-labor_category",
		"search":   "senior",
	}, nil)
}

func TestListIDVSubresourcePathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVAwards(context.Background(), "KEY WITH SPACE", nil)
	assertPathContains(t, capturedURL, "KEY%20WITH%20SPACE")
}
