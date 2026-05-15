package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListOfficesFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListOfficesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"search", "limit"},
		},
		{
			name: "search and pagination",
			opts: &ListOfficesOptions{
				ListOptions: ListOptions{Limit: 25},
				Search:      "procurement",
			},
			wantQS: map[string]string{
				"limit":  "25",
				"search": "procurement",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListOfficesOptions{},
			notInQS: []string{"search"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListOffices(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetOfficeRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetOffice(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetOfficeBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOffice(context.Background(), "FA8650")
	assertPathContains(t, capturedURL, "/api/offices/FA8650/")
}
