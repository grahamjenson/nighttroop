package nightscoutapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
)

// testConfig allows 3 writes a site and identity, then 1 a minute; a config can only be made once per process
var testConfig *config.ServerConfig

func TestMain(m *testing.M) {
	os.Setenv("SITE_POST_PER_MINUTE_LIMIT", "1")
	os.Setenv("SITE_POST_PER_MINUTE_BURST", "3")
	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	testConfig = cfg
	os.Exit(m.Run())
}

var runs atomic.Int64

// A websocket's events skip the HTTP rate limits, so writes over one must count against them too
func TestSocketWritesAreRateLimited(t *testing.T) {
	cfg := testConfig
	// A new site and identity each run, so -count does not share their limits
	run := runs.Add(1)
	ctx := httputils.CreateSiteContext(context.Background(), cfg, nil, nil, nil, fmt.Sprint("site", run), "127.0.0.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{ID: run, Name: "phone"})

	ws := &Websocket{config: cfg}
	writes := 0
	write := func(context.Context, nsutils.Collection, string, json.RawMessage) any {
		writes++
		return result("success")
	}
	args := []json.RawMessage{json.RawMessage(`{"collection":"treatments","data":{}}`)}
	auth := &authorization{Read: true, Write: true, WriteTreatment: true}

	for range 3 {
		require.Equal(t, result("success"), ws.dbWrite("dbAdd", args, auth, ctx, write))
	}
	require.Equal(t, result("Too many requests"), ws.dbWrite("dbAdd", args, auth, ctx, write))
	require.Equal(t, 3, writes)
}
