package sitemodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/logsdb"
)

// A batch's event is about every site in it, so a site sees that it happened but not its attributes or job
func TestSiteEventFromDBHidesItsBatchsAttributes(t *testing.T) {
	site, batch, attributes, request, job := "site", "s", `{"batch":"s","site_count":12}`, "req-1", "job-1"

	own := SiteEventFromDB(&logsdb.Event{ID: 1, SiteID: &site, Name: "site_disabled", Attributes: &attributes, RequestID: &request})
	require.Nil(t, own.Batch)
	require.JSONEq(t, attributes, string(own.Attributes))
	require.Equal(t, request, *own.RequestID)

	batchEvent := SiteEventFromDB(&logsdb.Event{ID: 2, Batch: &batch, Name: "site_batch_backed_up", Attributes: &attributes, JobID: &job})
	require.Equal(t, batch, *batchEvent.Batch)
	require.Nil(t, batchEvent.Attributes)
	require.Nil(t, batchEvent.JobID)
	body, err := json.Marshal(batchEvent)
	require.NoError(t, err)
	require.NotContains(t, string(body), "site_count")
}
