package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListBusinessTypesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListBusinessTypes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestListBusinessTypesWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListBusinessTypes(context.Background(), &ListOptions{Limit: 50, Page: 2})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "50", "page": "2"}, nil)
}

func TestListBusinessTypesBuildsCorrectPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListBusinessTypes(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/business_types/")
}

func TestGetBusinessTypeRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBusinessType(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBusinessTypeBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetBusinessType(context.Background(), "8A")
	assertPathContains(t, capturedURL, "/api/business_types/8A/")
}
