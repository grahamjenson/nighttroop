package sitemodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/logsdb"
)

// A batch's event is about every site in it, so a site sees that it happened but not its attributes
func TestSiteEventFromDBHidesItsBatchsAttributes(t *testing.T) {
	site, batch, attributes := "site", "s", `{"batch":"s","site_count":12}`

	own := SiteEventFromDB(&logsdb.Event{ID: 1, SiteID: &site, Name: "site_disabled", Attributes: &attributes})
	require.Nil(t, own.Batch)
	require.JSONEq(t, attributes, string(own.Attributes))

	batchEvent := SiteEventFromDB(&logsdb.Event{ID: 2, Batch: &batch, Name: "site_batch_backed_up", Attributes: &attributes})
	require.Equal(t, batch, *batchEvent.Batch)
	require.Nil(t, batchEvent.Attributes)
	body, err := json.Marshal(batchEvent)
	require.NoError(t, err)
	require.NotContains(t, string(body), "site_count")
}
