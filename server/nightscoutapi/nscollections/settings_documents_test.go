package nscollections

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// A site keeps only a few settings documents of apps' own: one more is refused, while those it has can still be
// replaced, and one deleted makes room for another
func TestSettingsDocumentsPerSiteAreCapped(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx := t.Context()
	siteContext := &httputils.SiteContext{SiteDB: site.SiteDB, IdentityModel: &controldb.SiteIdentity{Name: "phone"}}
	put := func(identifier string) (bool, *nsutils.HTTPError) {
		_, created, err := putDocument(ctx, siteContext, identifier, map[string]any{"date": 946684800001, "app": "AAPS"}, true)
		return created, err
	}

	for i := range maxSettingsDocuments {
		created, err := put(fmt.Sprintf("doc-%d", i))
		require.Nil(t, err)
		require.True(t, created)
	}
	_, err := put("one-too-many")
	require.Equal(t, ErrTooManySettingsDocuments, err)

	created, err := put("doc-0")
	require.Nil(t, err)
	require.False(t, created, "replacing a document kept is not creating one")

	require.Nil(t, deleteDocument(ctx, siteContext, "doc-1"))
	created, err = put("one-too-many")
	require.Nil(t, err)
	require.True(t, created)
}
