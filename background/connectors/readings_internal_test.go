package connectors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/libremodels"
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
