package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListMasSinsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListMasSinsOptions
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
			opts: &ListMasSinsOptions{
				ListOptions: ListOptions{Limit: 20},
				Search:      "cloud",
			},
			wantQS: map[string]string{
				"limit":  "20",
				"search": "cloud",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListMasSinsOptions{},
			notInQS: []string{"search"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListMasSins(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetMasSinRequiresSin(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetMasSin(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetMasSinBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetMasSin(context.Background(), "54151S")
	assertPathContains(t, capturedURL, "/api/mas_sins/54151S/")
}
