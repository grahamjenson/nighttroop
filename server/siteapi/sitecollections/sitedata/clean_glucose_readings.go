package sitedata

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_data_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/glucose"
	"nighttroop.t1d.nz/utils/httputils"
)

// CleanGlucoseReadingsCollection is the site's glucose readings cleaned, as the summary shows them: the same glucose
// read by more than one source is one reading, which says its change since the one before. It is only read.
type CleanGlucoseReadingsCollection struct{}

// trendLookback is how long before a time readings are read, so the first reading after it has one to change from
const trendLookback = 15 * time.Minute

// maxRawReadings caps the readings cleaned at a week of them every five minutes from three sources
const maxRawReadings = 3 * maxRows

func (s *CleanGlucoseReadingsCollection) GetName() string {
	return "clean_glucose"
}

// GetRoute is where the collection is in its URL
func (s *CleanGlucoseReadingsCollection) GetRoute() string {
	return "clean_glucose"
}

// Watches is any change to a glucose reading, which can change what is kept
func (s *CleanGlucoseReadingsCollection) Watches(event events.Event, _ *httputils.SiteContext) bool {
	_, reading := event.Model.(*sitedb.GlucoseReading)
	return reading
}

func (s *CleanGlucoseReadingsCollection) ToModel(row *sitemodels.SiteCleanGlucoseReading, _ *httputils.SiteContext) *sitemodels.SiteCleanGlucoseReading {
	return row
}

// cleanReadingsDefaultHours is how far back the cleaned readings go when the request does not say
const cleanReadingsDefaultHours = 48

// SearchRows is the cleaned readings of the last hours, 48 when not given, newest first, paged by limit and skip
func (s *CleanGlucoseReadingsCollection) SearchRows(ctx context.Context, collectionTx *siteapiutils.CollectionTx, query *apiutils.Query, siteContext *httputils.SiteContext) ([]*sitemodels.SiteCleanGlucoseReading, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	hours := int64(cleanReadingsDefaultHours)
	if query.Hours != nil {
		hours = *query.Hours
	}
	readings, err := cleanGlucoseReadings(ctx, q, time.Now().Add(-time.Duration(hours)*time.Hour).UnixMilli(), nil, siteContext)
	if err != nil {
		return nil, err
	}
	start := min(query.Skip, int64(len(readings)))
	return readings[start:min(start+query.Limit, int64(len(readings)))], nil
}

func (s *CleanGlucoseReadingsCollection) SearchHTMXResponse(models []*sitemodels.SiteCleanGlucoseReading, query *apiutils.Query) templ.Component {
	readings := make([]*sitemodels.SiteGlucoseReading, len(models))
	for i, model := range models {
		readings[i] = &model.SiteGlucoseReading
	}
	return site_data_templ.GlucoseReadingsTable(readings, query.Skip, query.Limit)
}

// GetRow is the cleaned reading a reading is, or was merged into
func (s *CleanGlucoseReadingsCollection) GetRow(ctx context.Context, collectionTx *siteapiutils.CollectionTx, id string, siteContext *httputils.SiteContext) (*sitemodels.SiteCleanGlucoseReading, error) {
	q, err := collectionTx.Site()
	if err != nil {
		return nil, err
	}
	row, err := q.GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: id})
	if err != nil {
		return nil, err
	}
	until := row.OccurredAt + glucose.SameWithinMs
	readings, err := cleanGlucoseReadings(ctx, q, row.OccurredAt-glucose.SameWithinMs, &until, siteContext)
	if err != nil {
		return nil, err
	}
	for _, reading := range readings {
		if reading.ID == id {
			return reading, nil
		}
		for _, other := range reading.Others {
			if other.ID == id {
				return reading, nil
			}
		}
	}
	return nil, nil
}

func (s *CleanGlucoseReadingsCollection) GetHTMXResponse(model *sitemodels.SiteCleanGlucoseReading) templ.Component {
	return site_data_templ.GlucoseReadingModal(&model.SiteGlucoseReading)
}

// cleanGlucoseReadings is the site's readings from since, until when given, cleaned, newest first. It reads from
// trendLookback before since, so the first has a change, and keeps only those from since.
func cleanGlucoseReadings(ctx context.Context, q *sitedb.Queries, since int64, until *int64, siteContext *httputils.SiteContext) ([]*sitemodels.SiteCleanGlucoseReading, error) {
	rows, err := q.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{DateGTE: new(since - trendLookback.Milliseconds()), DateLTE: until, Limit: maxRawReadings})
	if err != nil {
		return nil, err
	}
	readings := []*sitemodels.SiteCleanGlucoseReading{}
	for _, c := range glucose.CleanReadings(rows) {
		if c.Row.OccurredAt >= since {
			readings = append(readings, cleanGlucoseReadingFromDB(c, siteContext))
		}
	}
	return readings, nil
}

// cleanGlucoseReadingFromDB is the site model of a cleaned reading, in the site context's units
func cleanGlucoseReadingFromDB(c *glucose.Clean, siteContext *httputils.SiteContext) *sitemodels.SiteCleanGlucoseReading {
	units := sitemodels.ContextUnits(siteContext)
	model := &sitemodels.SiteCleanGlucoseReading{SiteGlucoseReading: *sitemodels.GlucoseReadingFromDB(c.Row, siteContext), Source: c.Source,
		Others: []sitemodels.SiteOtherGlucoseReading{}}
	if c.DiffPer5mMgdl != nil {
		model.DiffPer5m = new(units.FromMGDL(*c.DiffPer5mMgdl))
	}
	for _, other := range c.Others {
		model.Others = append(model.Others, sitemodels.SiteOtherGlucoseReading{ID: other.ID, Reading: units.FromMGDL(other.Mgdl), Source: other.Source})
	}
	return model
}
