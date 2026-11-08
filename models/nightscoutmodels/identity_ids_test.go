package nightscoutmodels_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/testutils"
	"nighttroop.t1d.nz/utils/types"
)

// Records keep the writer's identity id, as its name can be reused after the identity is deleted
func TestRecordsStoreTheWritersIdentityID(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx := site.SiteDB.Queries(), context.Background()
	sgv := types.FlexibleFloat64(120)
	entry := nightscoutmodels.Entry{SGV: &sgv, Mgdl: &sgv}
	rawJSON := "{}"
	// The API stores the raw JSON after converting, so the tests do too
	create := func(cc nightscoutmodels.ConverterContext) sitedb.CreateGlucoseReadingParams {
		params := nightscoutmodels.EntryToCreate(entry, cc)
		params.RawJsonData = &rawJSON
		return params
	}

	created, err := q.CreateGlucoseReading(ctx, create(nightscoutmodels.ConverterContext{IdentityID: 7, IdentityName: "kitchen"}))
	require.NoError(t, err)
	require.Equal(t, int64(7), *created.CreatedByIdentityID)
	require.Equal(t, "kitchen", created.CreatedByIdentityName)

	update := nightscoutmodels.EntryToUpdate(entry, nightscoutmodels.ConverterContext{IdentityID: 8, IdentityName: "phone"})
	update.ID, update.RawJsonData = created.ID, &rawJSON
	updated, err := q.UpdateGlucoseReading(ctx, update)
	require.NoError(t, err)
	require.Equal(t, int64(8), *updated.UpdatedByIdentityID)
	require.Equal(t, int64(7), *updated.CreatedByIdentityID)

	// Background jobs write as an identity with no id
	background, err := q.CreateGlucoseReading(ctx, create(nightscoutmodels.ConverterContext{IdentityName: "virtual_connector"}))
	require.NoError(t, err)
	require.Nil(t, background.CreatedByIdentityID)
}
