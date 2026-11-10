package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// Sites created at the same time must each get their own subdomain and site DB
func TestCreateSiteConcurrently(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)

	const count = 10
	subdomains := make([]string, count)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := controlDB.CreateSite(t.Context())
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

		siteDB, _, err := controlDB.GetSite(t.Context(), subdomains[i])
		require.NoError(t, err, "site %s was not created", subdomains[i])
		_, err = siteDB.Queries().GetMostRecentProfileStore(t.Context())
		require.NoError(t, err, "site %s DB is not usable", subdomains[i])
	}
}

// A site DB that cannot be created is an error, not a panic
func TestCreateSiteWhenSiteDBFails(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)

	// A file where the data folder should be, so the site DB cannot be written
	badCfg := *cfg
	badCfg.Mode = consts.Development
	badCfg.DataDir = filepath.Join(t.TempDir(), "not-a-folder")
	require.NoError(t, os.WriteFile(badCfg.DataDir, nil, 0o600))
	controlDB.ServerConfig = &badCfg

	_, err := controlDB.CreateSite(t.Context())
	require.ErrorContains(t, err, "Site DB Creation Error")
}

func TestArchiveSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)

	// An upload that fails keeps the site
	require.Error(t, controlDB.ArchiveSite(ctx, siteID, func(context.Context, string, s3utils.Retention, []string) error { return errors.New("S3 is down") }))
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "disabled", row.State)

	// The site's backup is uploaded to its archive path, then it is archived and closed here
	var uploadedKey string
	var uploadedRetention s3utils.Retention
	var uploadedPaths []string
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(_ context.Context, key string, retention s3utils.Retention, paths []string) error {
		uploadedKey, uploadedRetention, uploadedPaths = key, retention, paths
		return nil
	}))
	require.Equal(t, sitedb.GetS3ArchivePath(siteID, cfg), uploadedKey)
	require.Equal(t, s3utils.SiteRetention, uploadedRetention)
	require.Len(t, uploadedPaths, 1)
	require.Equal(t, filepath.Base(site.SiteDB.BackupPath), filepath.Base(uploadedPaths[0]))
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)
	require.NotNil(t, row.ArchivedAt)
	require.NoFileExists(t, site.SiteDB.BackupPath)

	// It is no longer open, so it cannot be archived again
	require.ErrorContains(t, controlDB.ArchiveSite(ctx, siteID, func(context.Context, string, s3utils.Retention, []string) error { return nil }), "is not open")
}

// A batch backup rewriting the site's backup file while it is archived does not change the archive uploaded
func TestArchiveSiteDuringBatchBackup(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)

	var uploaded string
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(_ context.Context, _ string, _ s3utils.Retention, paths []string) error {
		require.NoError(t, utils.MkParentDirs(site.SiteDB.BackupPath))
		require.NoError(t, os.WriteFile(site.SiteDB.BackupPath, []byte("half written"), 0o600))
		data, err := os.ReadFile(paths[0])
		require.NoError(t, err)
		uploaded = paths[0]
		require.True(t, strings.HasPrefix(string(data), "SQLite format 3\x00"), "the archive is a whole DB")
		return nil
	}))
	// Named as the site's DB so it hydrates, and removed once uploaded
	require.Equal(t, filepath.Base(site.SiteDB.BackupPath), filepath.Base(uploaded))
	require.NoFileExists(t, uploaded)
}

// An archived site hydrates from its archive with its data, live again and open, and only an archived one does
func TestHydrateSite(t *testing.T) {
	cfg, controlDB := testutils.NewControlDB(t)
	// Site DBs in files, so the archive holds the site's data
	fileCfg := *cfg
	fileCfg.Mode = consts.Development
	fileCfg.DataDir = t.TempDir()
	controlDB.ServerConfig = &fileCfg
	ctx := t.Context()
	site, err := controlDB.CreateSite(ctx)
	require.NoError(t, err)
	siteID := site.SiteDB.Subdomain
	_, err = site.SiteDB.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)

	// A stand in for S3, holding the archive's file
	archive := filepath.Join(t.TempDir(), siteID+".sqlite")
	download := func(_ context.Context, key string, outpath string) (int64, bool, error) {
		require.Equal(t, sitedb.GetS3ArchivePath(siteID, &fileCfg), key)
		data, err := os.ReadFile(archive)
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		require.NoError(t, err)
		return int64(len(data)), true, os.WriteFile(filepath.Join(outpath, siteID+".sqlite"), data, 0o600)
	}

	// A live site, open here, is not hydrated
	require.ErrorContains(t, controlDB.HydrateSite(ctx, siteID, download), "already open")

	_, err = controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID, func(_ context.Context, _ string, _ s3utils.Retention, paths []string) error {
		data, err := os.ReadFile(paths[0])
		require.NoError(t, err)
		return os.WriteFile(archive, data, 0o600)
	}))
	require.NoFileExists(t, sitedb.GetPath(siteID, &fileCfg))

	// A download that fails, or finds no archive, leaves it archived
	require.Error(t, controlDB.HydrateSite(ctx, siteID, func(context.Context, string, string) (int64, bool, error) { return 0, false, errors.New("S3 is down") }))
	require.ErrorContains(t, controlDB.HydrateSite(ctx, siteID, func(context.Context, string, string) (int64, bool, error) { return 0, false, nil }), "has no archive")
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)

	// Hydrated, it is live and open with its data, as if just used so it is not disabled again straight away
	require.NoError(t, controlDB.HydrateSite(ctx, siteID, download))
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "live", row.State)
	require.Nil(t, row.ArchivedAt)
	require.Nil(t, row.DisabledAt)
	require.NotNil(t, row.LastSeenAt)
	require.InDelta(t, time.Now().UnixMilli(), *row.LastSeenAt, 60_000)
	siteDB, _, err := controlDB.GetSite(ctx, siteID)
	require.NoError(t, err)
	reading, err := siteDB.Queries().GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, 120.0, reading.Mgdl)
}

// A loop's status is kept only on its device's newest report with one; the others' are cleared, keeping the IOB and
// COB extracted into their columns: for every open site every 10 minutes, and for one site when it is truncated
func TestTruncateClearsOldLoopStatuses(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	aaps := `{"iob":{"iob":1.2},"suggested":{"COB":20,"reason":"COB 20","predBGs":{"IOB":[120,118]}}}`
	loop := `{"iob":{"iob":3},"cob":{"cob":12},"predicted":{"startDate":"2024-01-01T00:05:00Z","values":[120,130]}}`
	create := func(id, device string, status *string, minutesAgo int) {
		_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: id, LoopStatusJson: status, LoopIob: new(1.2),
			OccurredAt: time.Now().Add(-time.Duration(minutesAgo) * time.Minute).UnixMilli(), CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	for device, status := range map[string]string{"aaps": aaps, "loop": loop} {
		create(device+"-old", device, &status, 10)
		create(device+"-newest", device, &status, 5)
	}
	// A newer report of AAPS's phone without a loop status, as a pump's own, leaves the one before it the newest
	create("aaps-pump-only", "aaps", nil, 1)
	report := func(id string) *sitedb.DeviceReport {
		report, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: id})
		require.NoError(t, err)
		return report
	}

	require.NoError(t, controlDB.TruncateSites(ctx))
	for device, kept := range map[string]string{"aaps": aaps, "loop": loop} {
		require.Nil(t, report(device+"-old").LoopStatusJson, "%s's older status is cleared", device)
		require.Equal(t, new(1.2), report(device+"-old").LoopIob, "%s's IOB extracted from it is kept", device)
		require.JSONEq(t, kept, *report(device + "-newest").LoopStatusJson, "%s's newest with a loop status is kept", device)
	}

	// The job looks only at reports written in the last day; truncating the site looks at them all
	newer := aaps
	create("aaps-newer", "aaps", &newer, 0)
	require.NoError(t, db.TruncateSiteSince(ctx, site.SiteDB, time.Now().Add(time.Hour).UnixMilli()))
	require.NotNil(t, report("aaps-newest").LoopStatusJson, "a report written before since is not looked at")
	// An in-memory DB has no size to record
	sizeKB, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.Nil(t, sizeKB)
	require.Nil(t, report("aaps-newest").LoopStatusJson)
	require.JSONEq(t, aaps, *report("aaps-newer").LoopStatusJson)

	_, err = controlDB.TruncateSite(ctx, "not-a-site")
	require.ErrorContains(t, err, "is not open")
}

// A pump as sent is kept, like a loop's status, only on its device's newest report with one; the others' are cleared,
// keeping their pump's battery and reservoir columns
func TestTruncateClearsOldPumps(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	pump := `{"battery":{"percent":60},"reservoir":100,"extended":{"ActiveProfile":"Default"}}`
	create := func(id string, pump *string, minutesAgo int) {
		_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: id, PumpJson: pump, PumpBatteryPercent: new(60.0),
			OccurredAt: time.Now().Add(-time.Duration(minutesAgo) * time.Minute).UnixMilli(), CreatedByDeviceName: new("aaps"), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	create("old", &pump, 10)
	create("newest", &pump, 5)
	// A newer report without a pump, as the loop's alone, leaves the one before it the newest
	create("loop-only", nil, 1)
	report := func(id string) *sitedb.DeviceReport {
		report, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: id})
		require.NoError(t, err)
		return report
	}

	require.NoError(t, controlDB.TruncateSites(ctx))
	require.Nil(t, report("old").PumpJson, "an older pump is cleared")
	require.Equal(t, new(60.0), report("old").PumpBatteryPercent, "its battery extracted from it is kept")
	require.JSONEq(t, pump, *report("newest").PumpJson, "the newest with a pump is kept")
}

// Truncating a site deletes the loop settings no profile or device report refers to any more, such as a deleted
// profile's, but each app's last seen, and keeps those one does; the job every 10 minutes leaves them all
func TestTruncateDeletesUnreferencedLoopSettings(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	ids := map[string]string{}
	upsert := func(name, app string, seenAt int64) {
		ids[name] = fmt.Sprintf("%064x", len(ids)+1)
		require.NoError(t, q.UpsertLoopSettings(ctx, sitedb.UpsertLoopSettingsParams{ID: ids[name], App: app, Source: "profile", SettingsJson: `{"maximumBolus":10}`, SeenAt: seenAt}))
	}
	for _, name := range []string{"loop", "remote", "configuration", "deleted-profile", "unreferenced"} {
		upsert(name, "loop", 1)
	}
	upsert("loop-newest", "loop", 2)
	upsert("aaps-older", "aaps", 1)
	upsert("aaps-newest", "aaps", 2)
	profile := func(id, loopSettingsID string, remoteSettingsID *string) {
		_, err := q.CreateProfilePreset(ctx, sitedb.CreateProfilePresetParams{ID: id, Name: "Default", ProfileStoreID: id, LoopSettingsID: &loopSettingsID,
			RemoteSettingsID: remoteSettingsID, OccurredAt: 1, CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	profile("kept", ids["loop"], new(ids["remote"]))
	profile("deleted", ids["deleted-profile"], nil)
	_, err := q.HardDeleteProfilePreset(ctx, sitedb.HardDeleteProfilePresetParams{ID: "deleted"})
	require.NoError(t, err)
	_, err = q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "report", ConfigurationID: new(ids["configuration"]), OccurredAt: time.Now().UnixMilli(), CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	kept := func(name string) bool {
		_, err := q.GetLoopSettings(ctx, sitedb.GetLoopSettingsParams{ID: ids[name]})
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		require.NoError(t, err)
		return true
	}

	require.NoError(t, controlDB.TruncateSites(ctx))
	require.True(t, kept("unreferenced") && kept("deleted-profile"), "the job every 10 minutes keeps them")
	_, err = controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	for _, name := range []string{"loop", "remote", "configuration", "loop-newest", "aaps-newest"} {
		require.True(t, kept(name), "%s's settings are kept", name)
	}
	for _, name := range []string{"deleted-profile", "unreferenced", "aaps-older"} {
		require.False(t, kept(name), "%s's settings, which no record refers to, are deleted", name)
	}
}

// deviceReportsFixture makes device reports and says which of them thinning keeps
type deviceReportsFixture struct {
	t       *testing.T
	q       *sitedb.Queries
	reports map[string]deviceReportMade
}

type deviceReportMade struct {
	device     string
	occurredAt int64
	deleted    bool
}

// every makes a report of device every 5 minutes, from 2 minutes past from ago for hours, so none is near a tier's edge
func (f *deviceReportsFixture) every(device string, from, hours time.Duration, now time.Time) {
	for ago := from + 2*time.Minute; ago < from+hours; ago += 5 * time.Minute {
		id := fmt.Sprintf("%s-%d", device, ago/time.Minute)
		occurredAt := now.Add(-ago).UnixMilli()
		_, err := f.q.CreateDeviceReport(f.t.Context(), sitedb.CreateDeviceReportParams{ID: id, OccurredAt: occurredAt, CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
		require.NoError(f.t, err)
		f.reports[id] = deviceReportMade{device: device, occurredAt: occurredAt}
	}
}

func (f *deviceReportsFixture) delete(id string) {
	_, err := f.q.SoftDeleteDeviceReport(f.t.Context(), sitedb.SoftDeleteDeviceReportParams{ID: id, UpdatedByIdentityName: new("phone")})
	require.NoError(f.t, err)
	report := f.reports[id]
	report.deleted = true
	f.reports[id] = report
}

// kept is whether thinning at now keeps a report: one of the last hour, deleted, or not thinned, is kept, as is the
// newest of its device's not deleted in its bucket, 10 minutes until a week old, 20 until two weeks and then an hour
func (f *deviceReportsFixture) kept(id string, now time.Time, thinned func(deviceReportMade) bool) bool {
	report := f.reports[id]
	age := time.Duration(now.UnixMilli()-report.occurredAt) * time.Millisecond
	if age < time.Hour || report.deleted || !thinned(report) {
		return true
	}
	bucket := int64(600_000)
	if age >= 14*24*time.Hour {
		bucket = 3_600_000
	} else if age >= 7*24*time.Hour {
		bucket = 1_200_000
	}
	for _, other := range f.reports {
		if other.device == report.device && !other.deleted && other.occurredAt > report.occurredAt && other.occurredAt/bucket == report.occurredAt/bucket {
			return false
		}
	}
	return true
}

// remaining is each report left, deleted or not
func (f *deviceReportsFixture) remaining() map[string]bool {
	left, err := f.q.AllDeviceReports(f.t.Context(), sitedb.AllDeviceReportsParams{IncludeDeleted: true, Limit: 10_000})
	require.NoError(f.t, err)
	remaining := map[string]bool{}
	for _, report := range left {
		remaining[report.ID] = true
	}
	return remaining
}

// A device's reports are all kept for the last hour, then the newest in every 10 minutes until a week old, 20 minutes
// until two weeks and an hour until 30 days, when all are deleted. It keeps its newest however old, until 30 days.
// A deleted report is neither removed, until 30 days, nor kept in another's place.
func TestTruncateThinsOldDeviceReports(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	f := &deviceReportsFixture{t: t, q: site.SiteDB.Queries(), reports: map[string]deviceReportMade{}}
	now, day := time.Now(), 24*time.Hour
	// A loop reporting every 5 minutes for 3 hours now, and either side of a week, two weeks and 30 days old
	for _, from := range []time.Duration{0, 7*day - 90*time.Minute, 14*day - 90*time.Minute, 30*day - 90*time.Minute} {
		f.every("aaps", from, 3*time.Hour, now)
	}
	f.every("pump", 4*time.Hour, 5*time.Minute, now) // a pump that last reported 4 hours ago
	f.every("phone", 31*day, 5*time.Minute, now)     // a phone that last reported 31 days ago
	f.delete("aaps-102")
	f.delete(fmt.Sprintf("aaps-%d", (14*day+time.Hour)/time.Minute+2))
	f.delete(fmt.Sprintf("aaps-%d", (30*day+time.Hour)/time.Minute+2))

	require.NoError(t, controlDB.TruncateSites(t.Context()))
	remaining := f.remaining()
	removed := 0
	for id, report := range f.reports {
		age := time.Duration(now.UnixMilli()-report.occurredAt) * time.Millisecond
		kept := age < 30*day && f.kept(id, now, func(deviceReportMade) bool { return true })
		require.Equal(t, kept, remaining[id], id)
		if !remaining[id] {
			removed++
		}
	}
	require.Greater(t, removed, 80, "most reports from before the last hour are removed")
	require.True(t, remaining["pump-242"] && remaining["aaps-2"] && remaining["aaps-57"] && remaining["aaps-102"])
	require.False(t, remaining["phone-44642"], "a device's newest report goes once 30 days old")
}

// The job only looks at reports written in the last day, and those which turned a week or two weeks old in it, as
// those are thinned again; other old reports were thinned by an earlier run
func TestThinOldDeviceReportsLooksAtThoseThatChangedTier(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	f := &deviceReportsFixture{t: t, q: site.SiteDB.Queries(), reports: map[string]deviceReportMade{}}
	day := 24 * time.Hour
	written := time.Now()
	for _, from := range []time.Duration{6*day - 3*time.Hour, 13*day - 3*time.Hour} {
		f.every("aaps", from, 30*time.Hour, written)
	}
	// A day after they were written, the job looks at those written since, none of them
	since := time.Now().Add(time.Millisecond)
	now := since.Add(day)
	require.NoError(t, f.q.ThinOldDeviceReports(t.Context(), sitedb.ThinOldDeviceReportsParams{Since: since.UnixMilli(), Now: now.UnixMilli()}))
	remaining, removed := f.remaining(), 0
	changedTier := func(report deviceReportMade) bool {
		return (report.occurredAt >= since.Add(-7*day).UnixMilli() && report.occurredAt <= now.Add(-7*day).UnixMilli()) ||
			(report.occurredAt >= since.Add(-14*day).UnixMilli() && report.occurredAt <= now.Add(-14*day).UnixMilli())
	}
	for id := range f.reports {
		require.Equal(t, f.kept(id, now, changedTier), remaining[id], id)
		if !remaining[id] {
			removed++
		}
	}
	require.Greater(t, removed, 300, "those that turned a week or two weeks old are thinned")
}

func TestNewControlDBOpensEverySite(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode = consts.Development // DBs in files, so they are there to open again
	cfg.DataDir = t.TempDir()
	ctx := context.Background()

	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	subdomains := []string{}
	for range 10 {
		site, err := controlDB.CreateSite(ctx)
		require.NoError(t, err)
		subdomains = append(subdomains, site.SiteDB.Subdomain)
	}
	require.NoError(t, controlDB.Close())

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	require.Equal(t, 10, controlDB.GetNSitesDBs())
	for _, subdomain := range subdomains {
		site, _, err := controlDB.GetSite(ctx, subdomain)
		require.NoError(t, err)
		_, err = site.Queries().GetMostRecentProfileStore(ctx)
		require.NoError(t, err)
	}
	require.NoError(t, controlDB.Close())
}

// A site whose identities cannot be made is not kept, so no live site is left without its DB
func TestCreateSiteWhenIdentitiesFail(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	ctx, siteID := t.Context(), "failsite"
	// An identity already holding the device's name makes the site's second identity fail, after the site and its admin
	// are written. Only a test's in-memory control DB, without foreign keys, lets it be made before its site.
	require.NoError(t, controlDB.Queries.CreateSiteIdentity(ctx, controldb.CreateSiteIdentityParams{SiteID: siteID,
		Name: consts.FIRST_PWD_NAME, PasswordHash: "taken", Role: consts.IdentityRoleDevicePWDPhone, CreatedByIdentityName: "test"}))

	_, err := controlDB.CreateSiteWithID(ctx, siteID)
	require.ErrorContains(t, err, "cannot create admin identity")
	_, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.ErrorIs(t, err, sql.ErrNoRows, "the site is not kept")
	_, err = controlDB.Queries.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.ErrorIs(t, err, sql.ErrNoRows, "the site's admin is not kept")
}

// A site that is not open is left out of its batch, and the rest of the batch is still backed up
func TestBatchLeavesOutSiteNotOpen(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	batch := controlDB.GetSiteDBsFromModels([]*controldb.Site{{ID: "notopen"}, {ID: site.SiteDB.Subdomain}})
	require.Equal(t, []*db.SiteDB{site.SiteDB}, batch)
}

// A live site whose DB file is missing, such as one restored before its batch was backed up, is left out of the start
// and of its batch, while the rest of its batch is opened
func TestStartWhenSiteDBIsMissing(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode = consts.Development
	cfg.DataDir = t.TempDir()
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	for _, siteID := range []string{"keptsite", "kostsite"} {
		_, err := controlDB.CreateSiteWithID(ctx, siteID)
		require.NoError(t, err)
	}
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.Remove(sitedb.GetPath("kostsite", &cfg)))

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err, "the server starts")
	t.Cleanup(func() { controlDB.Close() })
	_, _, err = controlDB.GetSite(ctx, "keptsite")
	require.NoError(t, err)
	_, _, err = controlDB.GetSite(ctx, "kostsite")
	require.ErrorContains(t, err, "not found in memory")

	sites, err := controlDB.Queries.GetLiveAndDisabledSites(ctx)
	require.NoError(t, err)
	require.Len(t, sites, 2)
	batch := controlDB.GetSiteDBsFromModels(sites)
	require.Len(t, batch, 1)
	require.Equal(t, "keptsite", batch[0].Subdomain)
}

// A live site whose DB file is missing at start is restored, with its data, from its batch's backup, downloaded once
// for the batch; one the backup does not have stops the start, named in its error, and each is a site_db_missing event
// the control admin sees
func TestRestoreMissingSites(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode = consts.Development
	cfg.DataDir = t.TempDir()
	cfg.Status.S3Working = true
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	for _, siteID := range []string{"keptsite", "kostsite", "kingsite"} {
		_, err := controlDB.CreateSiteWithID(ctx, siteID)
		require.NoError(t, err)
	}
	site, _, err := controlDB.GetSite(ctx, "kostsite")
	require.NoError(t, err)
	_, err = site.Queries().CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())

	// The batch's backup, in a stand in for S3, has kostsite but not kingsite
	backup := filepath.Join(t.TempDir(), "kostsite.sqlite")
	require.NoError(t, os.Rename(sitedb.GetPath("kostsite", &cfg), backup))
	require.NoError(t, os.Remove(sitedb.GetPath("kingsite", &cfg)))
	downloads := 0
	download := func(_ context.Context, key string, outpath string) (int64, bool, error) {
		downloads++
		require.Equal(t, sitedb.GetS3BatchPath("k", &cfg), key)
		data, err := os.ReadFile(backup)
		require.NoError(t, err)
		return int64(len(data)), true, os.WriteFile(filepath.Join(outpath, "kostsite.sqlite"), data, 0o600)
	}

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	err = controlDB.RestoreMissingSites(ctx, download)
	require.ErrorIs(t, err, db.ErrSitesMissing, "kingsite is not in the backup, so the start stops")
	require.ErrorContains(t, err, "kingsite")
	require.NotContains(t, err.Error(), "kostsite")
	require.Equal(t, 1, downloads, "the batch is downloaded once")

	site, _, err = controlDB.GetSite(ctx, "kostsite")
	require.NoError(t, err, "the site is restored and open")
	reading, err := site.Queries().GetGlucoseReading(ctx, sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, float64(120), reading.Mgdl)
	_, _, err = controlDB.GetSite(ctx, "kingsite")
	require.ErrorContains(t, err, "not found in memory")
	entries, err := os.ReadDir(filepath.Dir(sitedb.GetBatchFolder("k", &cfg)))
	require.NoError(t, err)
	for _, entry := range entries {
		require.False(t, strings.HasPrefix(entry.Name(), "."), "no staging folder is left behind")
	}

	for siteID, attributes := range map[string]string{"kostsite": `{"restored":true}`, "kingsite": `{"restored":false,"reason":"not_in_backup"}`, "keptsite": ""} {
		found, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_db_missing"), Limit: 10})
		require.NoError(t, err)
		if attributes == "" {
			require.Empty(t, found, siteID)
			continue
		}
		require.Len(t, found, 1, siteID)
		require.JSONEq(t, attributes, *found[0].Attributes, siteID)
	}
}

// Without S3 a live site whose DB file is missing stops the start, and the control admin sees it as a site_db_missing
// event
func TestRestoreMissingSitesWithoutS3(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode = consts.Development
	cfg.DataDir = t.TempDir()
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(ctx, "kostsite")
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.Remove(sitedb.GetPath("kostsite", &cfg)))

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	err = controlDB.RestoreMissingSites(ctx, s3utils.NewS3Client(&cfg).DownloadFilesFromS3)
	require.ErrorIs(t, err, db.ErrSitesMissing)
	require.ErrorContains(t, err, "kostsite")

	_, _, err = controlDB.GetSite(ctx, "kostsite")
	require.ErrorContains(t, err, "not found in memory")
	siteID := "kostsite"
	found, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_db_missing"), Limit: 10})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.JSONEq(t, `{"restored":false,"reason":"s3_not_working"}`, *found[0].Attributes)
}

// A server whose missing sites are all restored from their batch's backup starts
func TestRestoreAllMissingSites(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir, cfg.Status.S3Working = consts.Development, t.TempDir(), true
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(ctx, "kostsite")
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())

	backup := filepath.Join(t.TempDir(), "kostsite.sqlite")
	require.NoError(t, os.Rename(sitedb.GetPath("kostsite", &cfg), backup))
	download := func(_ context.Context, _ string, outpath string) (int64, bool, error) {
		return 0, true, os.Rename(backup, filepath.Join(outpath, "kostsite.sqlite"))
	}

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	require.NoError(t, controlDB.RestoreMissingSites(ctx, download))
	_, _, err = controlDB.GetSite(ctx, "kostsite")
	require.NoError(t, err, "the site is restored and open")
}

// A missing site restored from its batch's backup that cannot be opened, as when its migration fails, stops the start,
// as a site that is there and cannot be opened does, so a release that cannot open it is rolled back
func TestRestoredSiteThatCannotOpenStopsTheStart(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir, cfg.Status.S3Working = consts.Development, t.TempDir(), true
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(ctx, "kostsite")
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())

	// Its backup is dirty at a migration it already has, so migrating it again fails, as a migration that keeps failing
	// does
	backup := filepath.Join(t.TempDir(), "kostsite.sqlite")
	require.NoError(t, os.Rename(sitedb.GetPath("kostsite", &cfg), backup))
	raw, err := sql.Open("sqlite", backup)
	require.NoError(t, err)
	_, err = raw.Exec("UPDATE schema_migrations SET dirty = 1")
	require.NoError(t, err)
	require.NoError(t, raw.Close())
	download := func(_ context.Context, _ string, outpath string) (int64, bool, error) {
		data, err := os.ReadFile(backup)
		require.NoError(t, err)
		return int64(len(data)), true, os.WriteFile(filepath.Join(outpath, "kostsite.sqlite"), data, 0o600)
	}

	controlDB, err = db.NewControlDB(ctx, &cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	require.ErrorIs(t, controlDB.RestoreMissingSites(ctx, download), db.ErrRestoredSiteCannotOpen)
}
