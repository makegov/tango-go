package tango

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestListAwardAnnouncementsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAwardAnnouncements(context.Background(), &ListAwardAnnouncementsOptions{
		ListOptions:         ListOptions{Limit: 10, Shape: ShapeAwardAnnouncementsMinimal},
		Search:              `"hypersonic"`,
		Source:              AwardAnnouncementSourceOfficialRelease,
		Agency:              "Navy|Army",
		Component:           "NAVY|ARMY",
		PublishedDateAfter:  "2026-09-01",
		PublishedDateBefore: "2026-09-30",
		AwardDateAfter:      "2026-08-01",
		AwardDateBefore:     "2026-08-31",
		AmountMin:           "100000000",
		AmountMax:           "500000000.50",
		PIID:                "N00019-25-C-0070|N0001925C0071",
		Awardee:             "lockheed",
		UEI:                 "ABC123DEF456",
		ActionType:          "new_award|modification",
		AnnouncementType:    "contract",
		Civilian:            boolPtr(false),
		SmallBusiness:       boolPtr(false),
		IsModification:      boolPtr(false),
		IsMultipleAward:     boolPtr(false),
		IncludeWithdrawn:    boolPtr(true),
		Ordering:            "-amount",
	})
	assertPathContains(t, capturedURL, "/api/news/award_announcements/")
	assertQueryContains(t, capturedURL, map[string]string{
		"limit":                 "10",
		"shape":                 ShapeAwardAnnouncementsMinimal,
		"search":                `"hypersonic"`,
		"source":                "official_release",
		"agency":                "Navy|Army",
		"component":             "NAVY|ARMY",
		"published_date_after":  "2026-09-01",
		"published_date_before": "2026-09-30",
		"award_date_after":      "2026-08-01",
		"award_date_before":     "2026-08-31",
		"amount_min":            "100000000",
		"amount_max":            "500000000.50",
		"piid":                  "N00019-25-C-0070|N0001925C0071",
		"awardee":               "lockheed",
		"uei":                   "ABC123DEF456",
		"action_type":           "new_award|modification",
		"announcement_type":     "contract",
		"civilian":              "false",
		"small_business":        "false",
		"is_modification":       "false",
		"is_multiple_award":     "false",
		"include_withdrawn":     "true",
		"ordering":              "-amount",
	}, nil)
}

func TestListAwardAnnouncementsOmitsUnsetFilters(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAwardAnnouncements(context.Background(), nil)
	assertQueryContains(t, capturedURL, nil, []string{"civilian", "small_business", "is_modification", "is_multiple_award", "include_withdrawn", "search", "source", "shape"})
}

func TestListAwardAnnouncementsForwardsExtra(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAwardAnnouncements(context.Background(), &ListAwardAnnouncementsOptions{Extra: map[string]any{"future_filter": "x"}})
	assertQueryContains(t, capturedURL, map[string]string{"future_filter": "x"}, nil)
}

func TestGetAwardAnnouncement(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	rec, err := c.GetAwardAnnouncement(context.Background(), "7d1e3a5c-2b4f-5a6e-8c0d-9f1b3d5e7a9c", &GetEntityOptions{Shape: ShapeAwardAnnouncementsComprehensive})
	if err != nil {
		t.Fatalf("GetAwardAnnouncement: %v", err)
	}
	if rec["id"] != "x" {
		t.Errorf("unexpected record %v", rec)
	}
	assertPathContains(t, capturedURL, "/api/news/award_announcements/7d1e3a5c-2b4f-5a6e-8c0d-9f1b3d5e7a9c/")
	assertQueryContains(t, capturedURL, map[string]string{"shape": ShapeAwardAnnouncementsComprehensive}, nil)

	var ve *ValidationError
	if _, err := c.GetAwardAnnouncement(context.Background(), "", nil); !errors.As(err, &ve) {
		t.Errorf("empty uuid: want *ValidationError, got %v", err)
	}
}

func TestIterateAwardAnnouncementsForwardsFilters(t *testing.T) {
	var seen []string
	c, _ := newTestClient(t, twoPageHandler(&seen))
	it := c.IterateAwardAnnouncements(context.Background(), &ListAwardAnnouncementsOptions{Agency: "Navy"})
	n := 0
	for it.Next() {
		n++
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	if n != 2 || len(seen) != 2 {
		t.Fatalf("want 2 rows over 2 requests, got %d rows over %v", n, seen)
	}
	assertQueryContains(t, seen[1], map[string]string{"agency": "Navy", "page": "2"}, nil)

	if c.IterateAwardAnnouncements(context.Background(), nil) == nil {
		t.Fatal("nil opts should still build an iterator")
	}
}

func TestAwardAnnouncementShapesLeaveTextToTheDetail(t *testing.T) {
	fields := func(shape string) map[string]bool {
		set := map[string]bool{}
		for _, f := range strings.Split(shape, ",") {
			set[f] = true
		}
		return set
	}
	minimal, comprehensive := fields(ShapeAwardAnnouncementsMinimal), fields(ShapeAwardAnnouncementsComprehensive)
	if minimal["text"] {
		t.Error("ShapeAwardAnnouncementsMinimal names text, which the list default leaves out")
	}
	for _, f := range []string{"uuid", "source", "source_name", "text", "awardees(*)", "organization(*)"} {
		if !comprehensive[f] {
			t.Errorf("ShapeAwardAnnouncementsComprehensive is missing %s", f)
		}
	}
	for f := range minimal {
		if !comprehensive[f] {
			t.Errorf("ShapeAwardAnnouncementsComprehensive drops list field %s", f)
		}
	}
}
