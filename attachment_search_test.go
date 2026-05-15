package tango

import (
	"context"
	"errors"
	"testing"
)

func TestSearchOpportunityAttachmentsRequiresQ(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.SearchOpportunityAttachments(context.Background(), SearchOpportunityAttachmentsOptions{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty Q, got %T: %v", err, err)
	}
}

func TestSearchOpportunityAttachmentsBuildsQuery(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.SearchOpportunityAttachments(context.Background(), SearchOpportunityAttachmentsOptions{
		Q:                    "statement of work cloud migration",
		TopK:                 5,
		IncludeExtractedText: true,
	})
	assertPathContains(t, capturedURL, "/api/opportunities/attachment-search/")
	assertQueryContains(t, capturedURL, map[string]string{
		"q":                      "statement of work cloud migration",
		"top_k":                  "5",
		"include_extracted_text": "true",
	}, nil)
}

func TestSearchOpportunityAttachmentsTopKZeroOmitted(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.SearchOpportunityAttachments(context.Background(), SearchOpportunityAttachmentsOptions{
		Q:    "test query",
		TopK: 0,
	})
	assertQueryContains(t, capturedURL, map[string]string{"q": "test query"}, []string{"top_k"})
}

func TestSearchOpportunityAttachmentsExtractedTextOmittedWhenFalse(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.SearchOpportunityAttachments(context.Background(), SearchOpportunityAttachmentsOptions{
		Q: "test",
	})
	assertQueryContains(t, capturedURL, nil, []string{"include_extracted_text"})
}
