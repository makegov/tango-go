package tango

import (
	"context"
	"net/url"
)

// Award announcements — /api/news/award_announcements/ (Tango API 5.14.0).
//
// One feed of federal award announcements from two sources, named on every row in source and source_name:
//
//   - official_release rows are the awards of $7.5 million or more that the Department of War publishes in its daily contracts release, with the text as published.
//   - makegov_generated rows cover new civilian IDVs of $25 million or more and every new OTA and OT IDV. They are written by MakeGov from FPDS data with a language model and are not government text.
//
// A generated announcement is dropped when an official release covers the same award.
// Announcements DoD later withdrew are hidden from lists unless IncludeWithdrawn is true, and stay retrievable by uuid.
//
// The resource is available at every tier.
// A list row omits text, which can run past 16,000 characters; name it in the shape to get it, or use GetAwardAnnouncement, whose default shape carries it along with the awardees and organization expands.
// Alerts on this resource use query type "award_announcement" and deliver alerts.award_announcement.match.

// Values of an award announcement's source field, for ListAwardAnnouncementsOptions.Source.
const (
	// AwardAnnouncementSourceOfficialRelease is the Department of War's daily contracts release, with the text as published.
	AwardAnnouncementSourceOfficialRelease = "official_release"
	// AwardAnnouncementSourceMakeGovGenerated is an announcement written by MakeGov from FPDS data with a language model; it is not government text.
	AwardAnnouncementSourceMakeGovGenerated = "makegov_generated"
)

// ListAwardAnnouncementsOptions filters /api/news/award_announcements/.
type ListAwardAnnouncementsOptions struct {
	ListOptions

	// Search is full-text search over the title and text, where a title match ranks above a text match; wrap it in double quotes for a phrase.
	Search string

	// Source is AwardAnnouncementSourceOfficialRelease or AwardAnnouncementSourceMakeGovGenerated.
	Source string

	// Agency is the awarding organization: a Tango agency name, abbreviation, code or organization key (for example "Navy"); join several with "|".
	Agency string
	// Component is the DoD component as printed in the official release (for example "NAVY"), matched exactly but case-insensitively; join several with "|".
	// Generated announcements have no component.
	Component string

	// Date bounds, inclusive (wire format: YYYY-MM-DD).
	PublishedDateAfter  string
	PublishedDateBefore string
	AwardDateAfter      string
	AwardDateBefore     string

	// AmountMin and AmountMax bound the announced amount in US dollars, inclusive.
	AmountMin string
	AmountMax string

	// PIID is a contract number named in the announcement, as printed or not (for example "N00019-25-C-0070" or "N0001925C0070"), since dashes and case are ignored; join several with "|".
	PIID string
	// Awardee is a case-insensitive substring of an awardee name, at least 3 characters.
	Awardee string
	// UEI is the exact UEI of an awardee.
	// Official releases print no UEI, so this matches only where one was resolved.
	UEI string

	// ActionType is the action type, such as "new_award"; join several with "|".
	ActionType string
	// AnnouncementType is the announcement type; join several with "|".
	AnnouncementType string

	// Civilian keeps announcements of a civilian award (true) or not (false).
	// It and the other *bool filters are pointers so that false reaches the server as a filter value.
	Civilian *bool
	// SmallBusiness keeps announcements with at least one small-business awardee (true), or those whose awardees are stated not to be small (false).
	// Generated announcements state neither, so they match neither value.
	SmallBusiness *bool
	// IsModification keeps announcements of a modification to an existing contract (true) or not (false).
	IsModification *bool
	// IsMultipleAward keeps announcements naming more than one awardee (true) or not (false).
	IsMultipleAward *bool
	// IncludeWithdrawn, when true, includes announcements DoD later withdrew.
	IncludeWithdrawn *bool

	// Ordering is one of published_date, amount or rank; prefix "-" for descending.
	// The API default is "-published_date", or "-rank" under Search. "rank" requires a non-empty Search.
	Ordering string

	Extra map[string]any
}

func (o *ListAwardAnnouncementsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "source", o.Source)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "component", o.Component)
	setIfNotEmpty(q, "published_date_after", o.PublishedDateAfter)
	setIfNotEmpty(q, "published_date_before", o.PublishedDateBefore)
	setIfNotEmpty(q, "award_date_after", o.AwardDateAfter)
	setIfNotEmpty(q, "award_date_before", o.AwardDateBefore)
	setIfNotEmpty(q, "amount_min", o.AmountMin)
	setIfNotEmpty(q, "amount_max", o.AmountMax)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "awardee", o.Awardee)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "action_type", o.ActionType)
	setIfNotEmpty(q, "announcement_type", o.AnnouncementType)
	setIfNotNilBool(q, "civilian", o.Civilian)
	setIfNotNilBool(q, "small_business", o.SmallBusiness)
	setIfNotNilBool(q, "is_modification", o.IsModification)
	setIfNotNilBool(q, "is_multiple_award", o.IsMultipleAward)
	setIfNotNilBool(q, "include_withdrawn", o.IncludeWithdrawn)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListAwardAnnouncements queries /api/news/award_announcements/.
func (c *Client) ListAwardAnnouncements(ctx context.Context, opts *ListAwardAnnouncementsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/news/award_announcements/", opts.toQuery())
}

// IterateAwardAnnouncements walks every award announcement matching opts.
func (c *Client) IterateAwardAnnouncements(ctx context.Context, opts *ListAwardAnnouncementsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListAwardAnnouncementsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListAwardAnnouncements(ctx, &next)
		},
	}
}

// GetAwardAnnouncement fetches one award announcement by its uuid (/api/news/award_announcements/{uuid}/).
//
// A withdrawn announcement stays retrievable here, with is_withdrawn set.
// The default shape carries the full field set, including text, awardees(*) and organization(*).
func (c *Client) GetAwardAnnouncement(ctx context.Context, announcementUUID string, opts *GetEntityOptions) (Record, error) {
	if announcementUUID == "" {
		return nil, &ValidationError{&APIError{Message: "award announcement uuid is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/news/award_announcements/"+pathEscape(announcementUUID)+"/", opts.toQuery())
}
