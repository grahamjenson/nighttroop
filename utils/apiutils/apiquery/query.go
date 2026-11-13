package apiquery

import (
	"net/url"
	"time"

	"nighttroop.t1d.nz/utils/mapstructure"
	"nighttroop.t1d.nz/utils/types"
)

// MaxQueryLimit is the most models one query returns, so no request can read a whole table at once
const MaxQueryLimit = 1000

// MaxEntriesQueryLimit is the most readings one Nightscout API search returns, readings being small rows read by
// their date's index: a day of 1 minute readings, twice over, as xDrip+'s follower asks for up to 2880 since its
// newest, and keeps the readings it was not given after a gap unread for good
const MaxEntriesQueryLimit = 2880

// MaxQuerySkip is the furthest a search pages, as SQLite walks every row it skips: 32 searches' worth, enough for
// Nightguard's 32 pages of a report and a few months of glucose on a site's pages
const MaxQuerySkip = 32 * MaxQueryLimit

// DefaultQueryLimit is how many models a query without a limit returns, a page of a table
const DefaultQueryLimit = 50

// MaxSearchLength is the longest text a search looks for
const MaxSearchLength = 64

// MaxQueryHours is a week, the furthest back a summary goes
const MaxQueryHours = 7 * 24

type Query struct {
	Limit int64 `json:"limit,omitempty"`
	Skip  int64 `json:"skip,omitempty"`
	// Before, a time in ms, pages by date: the rows older than it, or with BeforeID, the ID of a row at it, older than
	// that row, as its rows are ordered by time then ID. BeforeDate, as 2006-01-02, is the rows on or before that day.
	Before     *int64  `json:"before,omitempty"`
	BeforeID   *string `json:"before_id,omitempty"`
	BeforeDate *string `json:"before_date,omitempty"`
	// Device, when set, keeps only what that device sent
	Device *string `json:"device,omitempty"`
	// Hours, when set, is how far back a summary goes, at most MaxQueryHours
	Hours *int64 `json:"hours,omitempty"`
	// Order, when set, is what a collection that can be ordered more than one way is ordered by
	Order *string `json:"order,omitempty"`
	// Type, when set, keeps only rows of that type, e.g. an event's name or an error's kind
	Type *string `json:"type,omitempty"`
	// Search, when set, keeps only rows whose text has it in, ignoring case
	Search *string `json:"search,omitempty"`
	// SiteID, RequestID and JobID, when set, keep only the logs DB's rows of that site, request or background job run.
	// Only the control logs pages read them: a site's own rows are always those of the site in its context.
	SiteID    *string `json:"site_id,omitempty"`
	RequestID *string `json:"request_id,omitempty"`
	JobID     *string `json:"job_id,omitempty"`
}

// QueryValues are the query string values, before they are checked and mapped to a Query
type QueryValues struct {
	Limit     *types.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip      *types.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`
	Device    *string              `mapstructure:"device,omitempty" json:"device,omitempty"`
	Hours     *types.FlexibleInt64 `mapstructure:"hours,omitempty" json:"hours,omitempty"`
	Order     *string              `mapstructure:"order,omitempty" json:"order,omitempty"`
	Type      *string              `mapstructure:"type,omitempty" json:"type,omitempty"`
	Search    *string              `mapstructure:"search,omitempty" json:"search,omitempty"`
	SiteID    *string              `mapstructure:"site_id,omitempty" json:"site_id,omitempty"`
	RequestID *string              `mapstructure:"request_id,omitempty" json:"request_id,omitempty"`
	JobID     *string              `mapstructure:"job_id,omitempty" json:"job_id,omitempty"`

	Before   *types.FlexibleInt64 `mapstructure:"before,omitempty" json:"before,omitempty"`
	BeforeID *string              `mapstructure:"before_id,omitempty" json:"before_id,omitempty"`
	// BeforeDate is a day in the site's timezone, which the site's search makes Before
	BeforeDate *string `mapstructure:"before_date,omitempty" json:"before_date,omitempty"`
}

func ParseQueryValues(values url.Values) (QueryValues, error) {
	params := map[string]string{}
	for key := range values {
		params[key] = values.Get(key)
	}

	var queryValues QueryValues
	err := mapstructure.Decode(params, &queryValues, nil)
	return queryValues, err
}

// BeforeDay makes a BeforeDate, a day in loc, Before: the rows before the next day starts. A day that is not a date
// is left out.
func (q *Query) BeforeDay(loc *time.Location) {
	day, err := time.ParseInLocation(time.DateOnly, *q.BeforeDate, loc)
	if err == nil {
		before := day.AddDate(0, 0, 1).UnixMilli()
		q.Before, q.BeforeID = &before, nil
	}
}
