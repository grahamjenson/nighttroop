package connectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/libremodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync keeps its new readings, leaves out those the site has, refuses one it cannot keep without losing the rest, and
// publishes each kept once, so a sync again of the same readings keeps and publishes none
func TestKeepReadingsKeepsOnlyNewOnes(t *testing.T) {
	cfg, controlDB, created := testutils.NewControlDBWithSite(t)
	siteDB, siteModel, err := controlDB.GetSite(t.Context(), created.SiteDB.Subdomain)
	require.NoError(t, err)
	site := &ConnectorSite{SiteModel: siteModel, SiteDB: siteDB, ControlDB: controlDB, Config: cfg}
	published := 0
	defer cfg.Events.Subscribe(siteModel.ID, func(event events.Event) {
		if _, ok := event.Model.(*sitedb.GlucoseReading); ok {
			published++
		}
	})()
	at := time.Now().Truncate(time.Minute)
	readings := []libremodels.Reading{{Mgdl: 100, FactoryTime: at.Add(-30 * time.Minute)}, {Mgdl: 110, FactoryTime: at.Add(-15 * time.Minute)},
		// Over the most a site keeps, so the site refuses it
		{Mgdl: 1500, FactoryTime: at.Add(-10 * time.Minute)}}
	toCreate := func(r libremodels.Reading) sitedb.CreateGlucoseReadingParams { return libremodels.ReadingToCreate(r, "libre_connector") }
	refused := []error{}
	refuse := func(err error, _ libremodels.Reading) { refused = append(refused, err) }

	kept, err := keepReadings(t.Context(), site, readings, toCreate, refuse)
	require.NoError(t, err)
	require.Equal(t, 2, kept)
	require.Equal(t, 2, published)
	require.Len(t, refused, 1)

	// The next sync has the same readings and a new one, so only the new one is kept
	readings = append(readings[:2], libremodels.Reading{Mgdl: 120, FactoryTime: at})
	kept, err = keepReadings(t.Context(), site, readings, toCreate, refuse)
	require.NoError(t, err)
	require.Equal(t, 1, kept)
	require.Equal(t, 3, published)
	require.Len(t, refused, 1)
	stored, err := siteDB.Queries().GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: toCreate(readings[2]).ID})
	require.NoError(t, err)
	require.Equal(t, 120.0, stored.Mgdl)

	// A sync whose context is done keeps none
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	kept, err = keepReadings(cancelled, site, []libremodels.Reading{{Mgdl: 130, FactoryTime: at.Add(time.Minute)}}, toCreate, refuse)
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, kept)
	require.Equal(t, 3, published)
}

// A connector's copy keeps its readings as written by the connector, skips one it cannot read or keep, records only the
// first of them as an error of the site's connector, and names the connector in an error that stops it
func TestCopyReadingsRecordsOnlyTheFirstRefused(t *testing.T) {
	cfg, controlDB, created := testutils.NewControlDBWithSite(t)
	siteDB, siteModel, err := controlDB.GetSite(t.Context(), created.SiteDB.Subdomain)
	require.NoError(t, err)
	jobInfo := &ConnectorJobInfo{SiteID: siteModel.ID, Type: consts.ConnectorTypeLibre,
		Site: &ConnectorSite{SiteModel: siteModel, SiteDB: siteDB, ControlDB: controlDB, Config: cfg}}
	at := time.Now().Truncate(time.Minute)
	var unreadable func(json.RawMessage, error)
	fetch := func(context.Context) ([]libremodels.Reading, error) {
		unreadable(json.RawMessage(`{"bad":1}`), errors.New("no value"))
		unreadable(json.RawMessage(`{"bad":2}`), errors.New("no value"))
		// Over the most a site keeps, so the site refuses it
		return []libremodels.Reading{{Mgdl: 100, FactoryTime: at}, {Mgdl: 1500, FactoryTime: at.Add(-5 * time.Minute)}}, nil
	}

	added, err := copyReadings(t.Context(), jobInfo, "Libre", &unreadable, fetch, libremodels.ReadingToCreate)
	require.NoError(t, err)
	require.Equal(t, 1, added)
	stored, err := siteDB.Queries().GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: libremodels.ReadingToCreate(libremodels.Reading{Mgdl: 100, FactoryTime: at}, "").ID})
	require.NoError(t, err)
	require.Equal(t, "libre_connector", stored.CreatedByIdentityName)
	errs, err := controlDB.Logs.Queries.SearchErrors(t.Context(), logsdb.SearchErrorsParams{SiteID: &siteModel.ID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "Libre: cannot read a reading")
	require.Contains(t, *errs[0].Attributes, `"record":{"bad":1}`)

	failing := func(context.Context) ([]libremodels.Reading, error) { return nil, errors.New("signed out") }
	_, err = copyReadings(t.Context(), jobInfo, "Libre", &unreadable, failing, libremodels.ReadingToCreate)
	require.EqualError(t, err, "Libre Error: signed out")
}
