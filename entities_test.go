package tango

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListEntitiesFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListEntitiesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"search", "uei", "naics"},
		},
		{
			name: "all filters",
			opts: &ListEntitiesOptions{
				Search:                    "Acme",
				CageCode:                  "1ABC5",
				NAICS:                     "541512",
				Name:                      "Acme Corp",
				PSC:                       "D302",
				PurposeOfRegistrationCode: "Z1",
				Socioeconomic:             "A5",
				State:                     "VA",
				TotalAwardsObligatedGte:   "100000",
				TotalAwardsObligatedLte:   "999999",
				UEI:                       "UEI123456789",
				ZipCode:                   "22201",
			},
			wantQS: map[string]string{
				"search":                       "Acme",
				"cage_code":                    "1ABC5",
				"naics":                        "541512",
				"name":                         "Acme Corp",
				"psc":                          "D302",
				"purpose_of_registration_code": "Z1",
				"socioeconomic":                "A5",
				"state":                        "VA",
				"total_awards_obligated_gte":   "100000",
				"total_awards_obligated_lte":   "999999",
				"uei":                          "UEI123456789",
				"zip_code":                     "22201",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListEntitiesOptions{},
			notInQS: []string{"search", "cage_code", "naics", "name", "uei"},
		},
		{
			name:   "extra map",
			opts:   &ListEntitiesOptions{Extra: map[string]any{"custom": "val"}},
			wantQS: map[string]string{"custom": "val"},
		},
		{
			name: "pagination applied",
			opts: &ListEntitiesOptions{
				ListOptions: ListOptions{Page: 2, Limit: 25},
			},
			wantQS: map[string]string{"page": "2", "limit": "25"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListEntities(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetEntityRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetEntity(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetEntityBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetEntity(context.Background(), "ABCDE1234567", nil)
	assertPathContains(t, capturedURL, "/api/entities/ABCDE1234567/")
}

func TestGetEntityWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetEntity(context.Background(), "ABCDE1234567", &GetEntityOptions{
		Shape:     ShapeEntitiesMinimal,
		Flat:      true,
		FlatLists: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape":      ShapeEntitiesMinimal,
		"flat":       "true",
		"flat_lists": "true",
	}, nil)
}

func TestGetEntityPathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	// UEI with special chars should be path-escaped
	_, _ = c.GetEntity(context.Background(), "ABC DEF", nil)
	assertPathContains(t, capturedURL, "ABC%20DEF")
}

func TestIterateEntitiesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(emptyListHandler))
	it := c.IterateEntities(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestListEntitiesHTTP404(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	_, err := c.GetEntity(context.Background(), "NONEXISTENT", nil)
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
}
