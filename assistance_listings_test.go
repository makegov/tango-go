package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListAssistanceListingsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListAssistanceListings(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestListAssistanceListingsWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAssistanceListings(context.Background(), &ListOptions{Limit: 50})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "50"}, nil)
}

func TestListAssistanceListingsBuildsCorrectPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAssistanceListings(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/assistance_listings/")
}

func TestGetAssistanceListingRequiresNumber(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetAssistanceListing(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetAssistanceListingBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetAssistanceListing(context.Background(), "10.001")
	assertPathContains(t, capturedURL, "/api/assistance_listings/10.001/")
}
