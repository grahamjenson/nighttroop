package sync_test

import (
	stdsync "sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync that works or fails, and a backfill, each tell the site's pages its source changed, so the sync settings
// page shows its new stats and errors
func TestSyncPublishesItsSourceChanging(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID := controlDB.Queries, t.Context(), site.SiteDB.Subdomain
	jobs := sync.SetupSyncJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}

	var lock stdsync.Mutex
	published := []consts.SyncSourceType{}
	unsubscribe := cfg.Events.Subscribe(siteID, func(event events.Event) {
		if source, ok := event.Model.(*sitemodels.SyncSource); ok && event.Action == events.Updated {
			lock.Lock()
			defer lock.Unlock()
			published = append(published, source.Type)
		}
	})
	defer unsubscribe()
	takePublished := func() []consts.SyncSourceType {
		lock.Lock()
		defer lock.Unlock()
		taken := published
		published = []consts.SyncSourceType{}
		return taken
	}

	tests := []struct {
		name     string
		typ      consts.SyncSourceType
		user     string
		backfill bool
		fails    bool
	}{
		{name: "a failed sync", typ: consts.SyncSourceTypeDexcom, user: "bad-user", fails: true},
		{name: "a sync that works", typ: consts.SyncSourceTypeVirtual, user: "virtual"},
		{name: "a backfill", typ: consts.SyncSourceTypeVirtual, backfill: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != "" {
				require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: tt.typ, UrlOrUsernameEnc: enc(tt.user), PasswordEnc: enc("pass"), Enabled: true}))
			}
			params := sync.SyncJobParams{SiteID: siteID, Type: tt.typ}
			if tt.backfill {
				to := time.Now().UnixMilli()
				params.Backfill = &sync.BackfillParams{From: to - time.Hour.Milliseconds(), To: to}
			}
			err := jobs.SyncSource(ctx, params)
			if tt.fails {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []consts.SyncSourceType{tt.typ}, takePublished())
		})
	}
}
