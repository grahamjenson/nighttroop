package control_sites_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

// TestSiteModalTruncate checks the Truncate button shows only for a site over its DB size limit, which the browser
// tests cannot make as their DBs are in memory
func TestSiteModalTruncate(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		name     string
		sizeKB   *int64
		truncate bool
	}{
		{"over its limit", size(4000), true},
		{"within its limit", size(1000), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			site := &controlmodels.Site{ID: "site", State: "live", EstimatedDbSizeKb: test.sizeKB, MaxDbSizeKbLimit: size(3000)}
			require.NoError(t, SiteModal(site).Render(context.Background(), &buf))
			require.Equal(t, test.truncate, bytes.Contains(buf.Bytes(), []byte(`id="control--site--truncate"`)), buf.String())
		})
	}
}

// A site's DB size, kept in KB, is shown in MB
func TestSizeMB(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	for _, test := range []struct {
		kb   *int64
		want string
	}{
		{nil, "-"},
		{size(0), "0.00"},
		{size(40), "0.04"},
		{size(1024), "1.00"},
		{size(3000), "2.93"},
		{size(512000), "500.00"},
	} {
		require.Equal(t, test.want, sizeMB(test.kb))
	}
}
