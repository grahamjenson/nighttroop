package httputils

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// Settings asked for before the identity is known are logged as an error without the site's or identity's secrets,
// as a site's logs are shown to its admin and kept in the logs DB
func TestGetSettingsLogsNoSecrets(t *testing.T) {
	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	defer slog.SetDefault(previous)

	site := &SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh", BackupKeyHash: "backup-key-hash"}}
	require.NotNil(t, site.GetSettings())
	identity := &SiteContext{SiteID: "abcdefgh", IdentityModel: &controldb.SiteIdentity{PasswordHash: "password-hash", TotpSecretEnc: "totp-secret"}}
	require.NotNil(t, identity.GetSettings())

	require.Contains(t, logged.String(), "abcdefgh")
	for _, secret := range []string{"backup-key-hash", "password-hash", "totp-secret"} {
		require.NotContains(t, logged.String(), secret)
	}
}

// Without a signed in identity, as for a live update sent to a site's sockets, the site's own settings are in effect,
// and that is not an error
func TestGetSettingsWithoutAnIdentityIsTheSites(t *testing.T) {
	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	defer slog.SetDefault(previous)

	site := &SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh", Units: new(consts.UnitsMGDL)}}
	require.Equal(t, consts.UnitsMGDL, *site.GetSettings().GetUnits())
	require.NotContains(t, logged.String(), "ERROR")
}

// A JWT issued in the same millisecond as its identity's credentials changed is refused too, as the reset may have come
// after it: a sign in, a request and a reset can all be within one millisecond
func TestCheckIssuedAfterCredentialsChanged(t *testing.T) {
	identity := controldb.SiteIdentity{CredentialsChangedAt: 1000}
	tests := []struct {
		name       string
		issuedAtMs int64
		refused    bool
	}{
		{"issued before", 999, true},
		{"issued in the same millisecond", 1000, true},
		{"issued after", 1001, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckIssuedAfterCredentialsChanged(identity, tt.issuedAtMs)
			require.Equal(t, tt.refused, err != nil)
		})
	}
}

// A write's model changes are each published alone as they are made, marked as batched, then all together once it is
// done; a change made after that is published alone, as is one made with no batch
func TestBatchModelChanges(t *testing.T) {
	bus := events.NewSiteBus()
	var published []events.Event
	bus.Subscribe("abcdefgh", func(event events.Event) { published = append(published, event) })
	ctx := CreateSiteContext(context.Background(), &config.ServerConfig{Events: bus}, nil, nil, nil, "abcdefgh", "", HTTPActionWrite, AcceptTypeJSON, ContentTypeJSON)

	batched, publish := BatchModelChanges(ctx)
	GetSiteContext(batched).PublishModelChange(events.Created, &controldb.Site{ID: "one"})
	GetSiteContext(batched).PublishModelChange(events.Deleted, &controldb.Site{ID: "two"})
	require.Len(t, published, 2)
	require.True(t, published[0].Batched)
	require.True(t, published[1].Batched)

	publish()
	require.Len(t, published, 3)
	require.Equal(t, events.ModelsChanged, published[2].Name)
	require.Equal(t, "abcdefgh", published[2].SiteID)
	require.Equal(t, []events.Event{published[0], published[1]}, published[2].Attributes)

	// Published once: a change after it, or with the site context not batched, is sent alone
	GetSiteContext(batched).PublishModelChange(events.Updated, &controldb.Site{ID: "one"})
	publish()
	GetSiteContext(ctx).PublishModelChange(events.Updated, &controldb.Site{ID: "two"})
	require.Len(t, published, 5)
	require.False(t, published[3].Batched)
	require.False(t, published[4].Batched)

	// A write with no changes publishes nothing, and one with no site context is left as it is
	_, publish = BatchModelChanges(ctx)
	publish()
	require.Len(t, published, 5)
	none, publish := BatchModelChanges(context.Background())
	publish()
	require.Nil(t, GetSiteContext(none))
}
