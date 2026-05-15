package tango

import (
	"context"
	"errors"
	"testing"
)

func TestAgencyContractsOptionsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *AgencyContractsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"ordering", "search", "joiner"},
		},
		{
			name: "ordering and search",
			opts: &AgencyContractsOptions{
				Ordering: "-award_date",
				Search:   "cloud",
			},
			wantQS: map[string]string{
				"ordering": "-award_date",
				"search":   "cloud",
			},
		},
		{
			name: "joiner only sent when Flat=true",
			opts: &AgencyContractsOptions{
				ListOptions: ListOptions{Flat: true},
				Joiner:      ".",
			},
			wantQS: map[string]string{"flat": "true", "joiner": "."},
		},
		{
			name:    "joiner not sent when Flat=false",
			opts:    &AgencyContractsOptions{Joiner: "."},
			wantQS:  map[string]string{},
			notInQS: []string{"joiner"},
		},
		{
			name:   "extra map",
			opts:   &AgencyContractsOptions{Extra: map[string]any{"tag": "test"}},
			wantQS: map[string]string{"tag": "test"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListAgencyAwardingContracts(context.Background(), "9700", tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListAgencyAwardingContractsRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListAgencyAwardingContracts(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListAgencyAwardingContractsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAgencyAwardingContracts(context.Background(), "9700", nil)
	assertPathContains(t, capturedURL, "/api/agencies/9700/contracts/awarding/")
}

func TestListAgencyFundingContractsRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListAgencyFundingContracts(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListAgencyFundingContractsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAgencyFundingContracts(context.Background(), "9800", nil)
	assertPathContains(t, capturedURL, "/api/agencies/9800/contracts/funding/")
}

func TestListAgencyContractsPathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	// Agency code with slash or special char should be path-escaped
	_, _ = c.ListAgencyAwardingContracts(context.Background(), "97/00", nil)
	assertPathContains(t, capturedURL, "97%2F00")
}
