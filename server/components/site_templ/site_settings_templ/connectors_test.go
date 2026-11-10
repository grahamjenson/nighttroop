package site_settings_templ

import (
	"bytes"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// A backfill's toast says what it fetches, and that it fetches the recent days again when it does
func TestBackfillQueuedSaysWhatItFetches(t *testing.T) {
	from := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC).UnixMilli()
	to, recent := from+24*time.Hour.Milliseconds(), from+5*24*time.Hour.Milliseconds()
	for _, test := range []struct {
		name       string
		recentFrom int64
		want       string
	}{
		{"one range", 0, "Fetching records from 1 Nov 10:00 to 2 Nov 10:00 UTC<"},
		{"and the recent days", recent, "Fetching records from 1 Nov 10:00 to 2 Nov 10:00 UTC, and again from 6 Nov 10:00 UTC to fill any gaps"},
	} {
		var buf bytes.Buffer
		require.NoError(t, BackfillQueued(consts.ConnectorTypeNightscout, from, to, test.recentFrom).Render(t.Context(), &buf))
		require.Contains(t, buf.String(), test.want, test.name)
	}
}

// A connector's password is not this site's, so it is no password box, which Chrome offers to save whatever its
// autocomplete, but a text box masked as one that password managers are told to leave alone
func TestConnectorPasswordsAreNotOfferedToBeSaved(t *testing.T) {
	for _, test := range []struct {
		name string
		form templ.Component
		id   string
	}{
		{"dexcom", DexcomConnectorForm(nil), "site--settings--dexcom-connector--password"},
		{"libre", LibreConnectorForm(nil), "site--settings--libre-connector--password"},
		{"nightscout", NightscoutConnectorForm(nil), "site--settings--nightscout-connector--password"},
	} {
		var buf bytes.Buffer
		require.NoError(t, test.form.Render(t.Context(), &buf), test.name)
		html := buf.String()
		require.NotContains(t, html, `type="password"`, test.name)
		require.Contains(t, html, `id="`+test.id+`" type="text" data-tui-input-masked style="-webkit-text-security: disc" `+
			`autocomplete="off" data-1p-ignore data-lpignore="true" data-bwignore data-form-type="other" name="password"`, test.name)
		// It can still be shown, as a password box can
		require.Contains(t, html, `data-tui-input-toggle-password="`+test.id+`"`, test.name)
	}
}
