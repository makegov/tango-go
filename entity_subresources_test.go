package tango

import (
	"context"
	"errors"
	"testing"
)

func TestEntitySubresourceOptionsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *EntitySubresourceOptions
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
			name: "all options",
			opts: &EntitySubresourceOptions{
				ListOptions: ListOptions{Limit: 10, Shape: "contracts(minimal)"},
				Joiner:      ".",
				Ordering:    "-award_date",
				Search:      "software",
			},
			wantQS: map[string]string{
				"limit":    "10",
				"shape":    "contracts(minimal)",
				"ordering": "-award_date",
				"search":   "software",
				// joiner only sent when Flat=true AND Joiner is set
			},
			// joiner not sent because Flat=false
			notInQS: []string{"joiner"},
		},
		{
			name: "joiner sent when Flat=true",
			opts: &EntitySubresourceOptions{
				ListOptions: ListOptions{Flat: true},
				Joiner:      "__",
			},
			wantQS: map[string]string{
				"flat":   "true",
				"joiner": "__",
			},
		},
		{
			name: "extra map",
			opts: &EntitySubresourceOptions{
				Extra: map[string]any{"custom_x": "val"},
			},
			wantQS: map[string]string{"custom_x": "val"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListEntityContracts(context.Background(), "UEI12345", tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListEntityContractsRequiresUEI(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListEntityContracts(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListEntityContractsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityContracts(context.Background(), "ABCDE1234567", nil)
	assertPathContains(t, capturedURL, "/api/entities/ABCDE1234567/contracts/")
}

func TestListEntityIDVsRequiresUEI(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListEntityIDVs(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListEntityIDVsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityIDVs(context.Background(), "UEI99999", nil)
	assertPathContains(t, capturedURL, "/api/entities/UEI99999/idvs/")
}

func TestListEntityOTAsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityOTAs(context.Background(), "UEI99999", nil)
	assertPathContains(t, capturedURL, "/api/entities/UEI99999/otas/")
}

func TestListEntityOTIDVsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityOTIDVs(context.Background(), "UEI99999", nil)
	assertPathContains(t, capturedURL, "/api/entities/UEI99999/otidvs/")
}

func TestListEntitySubawardsRequiresUEI(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListEntitySubawards(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListEntitySubawardsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntitySubawards(context.Background(), "UEI12345", nil)
	assertPathContains(t, capturedURL, "/api/entities/UEI12345/subawards/")
}

func TestEntitySubawardsOptionsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntitySubawards(context.Background(), "UEI12345", &EntitySubawardsOptions{
		Ordering: "last_modified_date",
		Extra:    map[string]any{"flag": "v2"},
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"ordering": "last_modified_date",
		"flag":     "v2",
	}, nil)
}

func TestListEntityLcatsRequiresUEI(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListEntityLcats(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListEntityLcatsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityLcats(context.Background(), "UEI12345", nil)
	assertPathContains(t, capturedURL, "/api/entities/UEI12345/lcats/")
}

func TestEntityLcatsOptionsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListEntityLcats(context.Background(), "UEI12345", &EntityLcatsOptions{
		Ordering: "labor_category",
		Search:   "software",
		Extra:    map[string]any{"region": "west"},
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"ordering": "labor_category",
		"search":   "software",
		"region":   "west",
	}, nil)
}

func TestListEntitySubresourcePathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	// UEI with a space should be path-escaped
	_, _ = c.ListEntityContracts(context.Background(), "UEI 1234", nil)
	assertPathContains(t, capturedURL, "UEI%201234")
}
