package db_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/testutils"
)

// Sites created at the same time must each get their own subdomain and site DB
func TestCreateSiteConcurrently(t *testing.T) {
	_, controlDB, err := testutils.GetConfigControlDBForTest()
	require.NoError(t, err)

	const count = 10
	subdomains := make([]string, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := controlDB.CreateSite()
			errs[i] = err
			if err == nil {
				subdomains[i] = result.SiteDB.Subdomain
			}
		}()
	}
	wg.Wait()

	seen := map[string]bool{}
	for i := range count {
		require.NoError(t, errs[i])
		require.False(t, seen[subdomains[i]], "subdomain %s created twice", subdomains[i])
		seen[subdomains[i]] = true

		siteDB, _, err := controlDB.GetSite(subdomains[i])
		require.NoError(t, err, "site %s was not created", subdomains[i])
		_, err = siteDB.Queries.GetMostRecentProfileStore(context.TODO())
		require.NoError(t, err, "site %s DB is not usable", subdomains[i])
	}
}
