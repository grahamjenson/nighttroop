package connectors_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync that works or fails, and a backfill, each tell the site's pages its source changed, so the connector settings
// page shows its new stats and errors
func TestSyncPublishesItsSourceChanging(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID := controlDB.Queries, t.Context(), site.SiteDB.Subdomain
	jobs := connectors.SetupConnectorJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.EncryptFor(security.PurposeConnector, siteID, s)
		require.NoError(t, err)
		return &v
	}

	var lock sync.Mutex
	published := []consts.ConnectorType{}
	unsubscribe := cfg.Events.Subscribe(siteID, func(event events.Event) {
		if source, ok := event.Model.(*sitemodels.Connector); ok && event.Action == events.Updated {
			lock.Lock()
			defer lock.Unlock()
			published = append(published, source.Type)
		}
	})
	defer unsubscribe()
	takePublished := func() []consts.ConnectorType {
		lock.Lock()
		defer lock.Unlock()
		taken := published
		published = []consts.ConnectorType{}
		return taken
	}

	tests := []struct {
		name     string
		typ      consts.ConnectorType
		user     string
		backfill bool
		fails    bool
	}{
		{name: "a failed sync", typ: consts.ConnectorTypeDexcom, user: "bad-user", fails: true},
		{name: "a sync that works", typ: consts.ConnectorTypeVirtual, user: "virtual"},
		{name: "a backfill", typ: consts.ConnectorTypeVirtual, backfill: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != "" {
				require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: tt.typ, UrlOrUsernameEnc: enc(tt.user), PasswordEnc: enc("pass"), Enabled: true}))
			}
			params := connectors.ConnectorJobParams{SiteID: siteID, Type: tt.typ}
			if tt.backfill {
				to := time.Now().UnixMilli()
				params.Backfill = &connectors.BackfillParams{From: to - time.Hour.Milliseconds(), To: to}
			}
			err := jobs.RunConnector(ctx, params)
			if tt.fails {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []consts.ConnectorType{tt.typ}, takePublished())
		})
	}
}
