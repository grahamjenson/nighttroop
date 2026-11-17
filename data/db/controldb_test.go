package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
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
	cfg, controlDB, site := testutils.NewReplicatedControlDB(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	addReading(t, site.SiteDB, "kept")
	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)

	// Its last writes are sent to its replica, which is kept as its archive, then it is archived and its files here go
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID))
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)
	require.NotNil(t, row.ArchivedAt)
	require.NoFileExists(t, site.SiteDB.Path)
	require.NoDirExists(t, filepath.Join(filepath.Dir(site.SiteDB.Path), "."+siteID+".sqlite-litestream"))
	require.NotEmpty(t, replicaKeys(t, cfg, siteID))

	// It is no longer open, so it cannot be archived again
	require.ErrorContains(t, controlDB.ArchiveSite(ctx, siteID), "is not open")
}

// Without S3 a site is not replicated, so it has no archive to keep and is not archived
func TestArchiveSiteWithoutS3(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)

	require.ErrorIs(t, controlDB.ArchiveSite(ctx, siteID), db.ErrNotReplicated)
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "disabled", row.State)
	require.ErrorIs(t, controlDB.HydrateSite(ctx, "noreplic"), db.ErrNotReplicated)
}

// addReading writes a glucose reading with id to the site
func addReading(t *testing.T, site *db.SiteDB, id string) {
	t.Helper()
	_, err := site.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: id, Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
}

// replicaKeys are the files of the site's replica in S3
func replicaKeys(t *testing.T, cfg *config.ServerConfig, siteID string) []string {
	t.Helper()
	return testutils.ReplicaFiles(t, cfg, s3utils.NewS3Client(cfg), replica.SiteName(siteID))
}

// requireReading checks the open site has the reading with id
func requireReading(t *testing.T, controlDB *db.ControlDB, siteID, id string) {
	t.Helper()
	site, _, err := controlDB.GetSite(t.Context(), siteID)
	require.NoError(t, err, "the site is open")
	reading, err := site.Queries().GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: id})
	require.NoError(t, err)
	require.Equal(t, 120.0, reading.Mgdl)
}

// An archived site hydrates from its replica with its data, live again, open and replicated, and only an archived one
// does
func TestHydrateSite(t *testing.T) {
	_, controlDB, site := testutils.NewReplicatedControlDB(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	addReading(t, site.SiteDB, "kept")

	// A live site, open here, is not hydrated
	require.ErrorContains(t, controlDB.HydrateSite(ctx, siteID), "already open")

	_, err := controlDB.Queries.DisableSite(ctx, controldb.DisableSiteParams{ID: siteID})
	require.NoError(t, err)
	require.NoError(t, controlDB.ArchiveSite(ctx, siteID))
	require.Nil(t, controlDB.SiteReplicatedAt(siteID))

	// A site without a replica has no archive, and stays archived
	_, err = controlDB.Queries.CreateSite(ctx, controldb.CreateSiteParams{ID: "noreplic", BackupKeyHash: "hash"})
	require.NoError(t, err)
	_, err = controlDB.Queries.ArchiveSites(ctx, controldb.ArchiveSitesParams{JsonEach: `["noreplic"]`})
	require.NoError(t, err)
	require.ErrorContains(t, controlDB.HydrateSite(ctx, "noreplic"), "has no archive")
	row, err := controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: "noreplic"})
	require.NoError(t, err)
	require.Equal(t, "archived", row.State)

	// Hydrated, it is live and open with its data, as if just used so it is not disabled again straight away, and is
	// replicated again
	require.NoError(t, controlDB.HydrateSite(ctx, siteID))
	row, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	require.Equal(t, "live", row.State)
	require.Nil(t, row.ArchivedAt)
	require.Nil(t, row.DisabledAt)
	require.NotNil(t, row.LastSeenAt)
	require.InDelta(t, time.Now().UnixMilli(), *row.LastSeenAt, 60_000)
	requireReading(t, controlDB, siteID, "kept")
	require.NotNil(t, controlDB.SiteReplicatedAt(siteID))
}

// A replicated site is truncated, its DB shrunk and its write-ahead log checkpointed by Litestream, and what it keeps
// is still sent to its replica
func TestTruncateReplicatedSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewReplicatedControlDB(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	for i := range 200 {
		addReading(t, site.SiteDB, fmt.Sprint("reading-", i))
	}
	start := time.Now()
	_, err := controlDB.TruncateSite(ctx, siteID)
	require.NoError(t, err)
	require.Less(t, time.Since(start), 4*time.Second, "it does not wait for Litestream's read to end")
	wal, err := os.Stat(site.SiteDB.Path + "-wal")
	require.NoError(t, err)
	require.Less(t, wal.Size(), int64(16<<10), "the write-ahead log is truncated, but for the frame Litestream writes after")
	addReading(t, site.SiteDB, "after the truncate")
	require.NoError(t, controlDB.Close())

	require.NoError(t, os.Remove(site.SiteDB.Path))
	require.NoError(t, os.RemoveAll(filepath.Join(filepath.Dir(site.SiteDB.Path), "."+siteID+".sqlite-litestream")))
	controlDB, err = testutils.OpenReplicated(t, cfg, s3utils.NewS3Client(cfg))
	require.NoError(t, err)
	requireReading(t, controlDB, siteID, "after the truncate")
}

// A deleted site's replica is deleted with it, and so is what Litestream kept beside its DB
func TestDeleteSitesDeletesTheirReplicas(t *testing.T) {
	cfg, controlDB, site := testutils.NewReplicatedControlDB(t)
	siteID := site.SiteDB.Subdomain
	require.NotEmpty(t, replicaKeys(t, cfg, siteID), "a new site is sent at once")

	require.NoError(t, controlDB.DeleteSites(t.Context(), []string{siteID}))
	require.Empty(t, replicaKeys(t, cfg, siteID))
	require.NoFileExists(t, site.SiteDB.Path)
	require.NoDirExists(t, filepath.Join(filepath.Dir(site.SiteDB.Path), "."+siteID+".sqlite-litestream"))
}

// A loop's status is kept whole for a day, as followers read a day of them; then only its predictions are cleared,
// keeping the rest as sent and its updated_at: for every open site every 10 minutes, and for one site when it is
// truncated
func TestTruncateClearsOldLoopStatuses(t *testing.T) {
	_, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), site.SiteDB.Queries()
	aaps := `{"iob":{"iob":1.2},"suggested":{"bg":110,"COB":20,"reason":"COB 20","predBGs":{"IOB":[120,118]}},"enacted":{"rate":1,"predBGs":{"ZT":[120]}}}`
	aapsCleared := `{"iob":{"iob":1.2},"suggested":{"bg":110,"COB":20,"reason":"COB 20"},"enacted":{"rate":1}}`
	loop := `{"name":"Loop","iob":{"iob":3},"cob":{"cob":12},"recommendedBolus":0.1,"predicted":{"startDate":"2024-01-01T00:05:00Z","values":[120,130]}}`
	loopCleared := `{"name":"Loop","iob":{"iob":3},"cob":{"cob":12},"recommendedBolus":0.1,"predicted":{"startDate":"2024-01-01T00:05:00Z"}}`
	createAt := func(id, device string, status *string, occurredAt time.Time) {
		_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: id, LoopStatusJson: status, LoopIob: new(1.2),
			OccurredAt: occurredAt.UnixMilli(), CreatedByDeviceName: &device, CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	create := func(id, device string, status *string, minutesAgo int) {
		createAt(id, device, status, time.Now().Add(-time.Duration(minutesAgo)*time.Minute))
	}
	statuses := map[string][2]string{"aaps": {aaps, aapsCleared}, "loop": {loop, loopCleared}}
	for device, status := range statuses {
		create(device+"-old", device, &status[0], 25*60)
		create(device+"-recent", device, &status[0], 10)
		create(device+"-newest", device, &status[0], 5)
	}
	// A newer report of AAPS's phone without a loop status, as a pump's own, does not change which are cleared
	create("aaps-pump-only", "aaps", nil, 1)
	// A device that stopped two days ago keeps its last day whole
	create("stopped-old", "stopped", &loop, 80*60)
	create("stopped-last-day", "stopped", &loop, 60*60)
	create("stopped-newest", "stopped", &loop, 50*60)
	report := func(id string) *sitedb.DeviceReport {
		report, err := q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: id})
		require.NoError(t, err)
		return report
	}
	updatedAt := report("aaps-old").UpdatedAt

	require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites()))
	for device, status := range statuses {
		require.JSONEq(t, status[1], *report(device + "-old").LoopStatusJson, "%s's status over a day old has no predictions", device)
		require.Equal(t, new(1.2), report(device+"-old").LoopIob, "%s's IOB extracted from it is kept", device)
		require.JSONEq(t, status[0], *report(device + "-recent").LoopStatusJson, "%s's status under a day old is kept whole", device)
		require.JSONEq(t, status[0], *report(device + "-newest").LoopStatusJson, "%s's newest is kept whole", device)
	}
	require.Equal(t, updatedAt, report("aaps-old").UpdatedAt, "clients syncing by srvModified do not fetch it again")
	require.JSONEq(t, loopCleared, *report("stopped-old").LoopStatusJson)
	require.JSONEq(t, loop, *report("stopped-last-day").LoopStatusJson, "a day before its device's newest is kept whole")
	require.JSONEq(t, loop, *report("stopped-newest").LoopStatusJson)

	// The job looks only at reports written in the last day, and those that turned a day old since then
	create("aaps-written-before", "aaps", &aaps, 26*60)
	require.NoError(t, db.TruncateSiteSince(ctx, site.SiteDB, time.Now().Add(time.Hour).UnixMilli()))
	require.JSONEq(t, aaps, *report("aaps-written-before").LoopStatusJson, "a report written before since is not looked at")
	start := time.Now()
	createAt("turning-newest", "turning", &aaps, start.Add(time.Second))
	createAt("turning-old", "turning", &aaps, start.Add(-24*time.Hour+200*time.Millisecond))
	since := time.Now().Add(time.Millisecond)
	time.Sleep(time.Until(start.Add(400 * time.Millisecond)))
	require.NoError(t, db.TruncateSiteSince(ctx, site.SiteDB, since.UnixMilli()))
	require.JSONEq(t, aapsCleared, *report("turning-old").LoopStatusJson, "a report written before since that turned a day old since is cleared")

	// Truncating the site looks at them all
	// An in-memory DB has no size to record
	sizeKB, err := controlDB.TruncateSite(ctx, site.SiteDB.Subdomain)
	require.NoError(t, err)
	require.Nil(t, sizeKB)
	require.JSONEq(t, aapsCleared, *report("aaps-written-before").LoopStatusJson)
	require.JSONEq(t, aaps, *report("aaps-newest").LoopStatusJson)

	_, err = controlDB.TruncateSite(ctx, "not-a-site")
	require.ErrorContains(t, err, "is not open")
}

// A pump as sent is kept only on its device's newest report with one; the others' are cleared,
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

	require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites()))
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

	require.NoError(t, controlDB.TruncateSites(ctx, controlDB.AllSites()))
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

	require.NoError(t, controlDB.TruncateSites(t.Context(), controlDB.AllSites()))
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
	// The site's second identity fails, after the site and its admin are written
	db.FailSiteIdentitiesAfterAdmin(t)

	_, err := controlDB.CreateSiteWithID(ctx, siteID)
	require.ErrorContains(t, err, "cannot create admin identity")
	_, err = controlDB.Queries.GetSiteInAnyState(ctx, controldb.GetSiteInAnyStateParams{ID: siteID})
	require.ErrorIs(t, err, sql.ErrNoRows, "the site is not kept")
	_, err = controlDB.Queries.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: consts.FIRST_ADMIN_NAME})
	require.ErrorIs(t, err, sql.ErrNoRows, "the site's admin is not kept")
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

// A live site whose DB file is missing at start is restored, with its data, from its replica; one without a replica
// stops the start, named in its error, and each is a site_db_missing event the control admin sees. The sites restored
// and kept are replicated again.
func TestReplicateRestoresMissingSites(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	ctx := t.Context()
	controlDB, err := db.NewControlDB(ctx, cfg)
	require.NoError(t, err)
	require.NoError(t, controlDB.Replicate(ctx, replica.New(cfg, s3)))
	for _, siteID := range []string{"keptsite", "kostsite", "kingsite"} {
		_, err := controlDB.CreateSiteWithID(ctx, siteID)
		require.NoError(t, err)
	}
	site, _, err := controlDB.GetSite(ctx, "kostsite")
	require.NoError(t, err)
	addReading(t, site, "kept")
	require.NoError(t, controlDB.Close())

	// kostsite's files are lost, but its replica has it; kingsite has neither
	for _, siteID := range []string{"kostsite", "kingsite"} {
		require.NoError(t, os.Remove(sitedb.GetPath(siteID, cfg)))
		require.NoError(t, os.RemoveAll(filepath.Join(filepath.Dir(sitedb.GetPath(siteID, cfg)), "."+siteID+".sqlite-litestream")))
	}
	require.NoError(t, replica.Client(cfg, s3, replica.SiteName("kingsite"), false).DeleteAll(ctx))

	controlDB, err = testutils.OpenReplicated(t, cfg, s3)
	require.ErrorIs(t, err, db.ErrSitesMissing, "kingsite has no replica, so the start stops")
	require.ErrorContains(t, err, "kingsite")
	require.NotContains(t, err.Error(), "kostsite")
	requireReading(t, controlDB, "kostsite", "kept")
	_, _, err = controlDB.GetSite(ctx, "kingsite")
	require.ErrorContains(t, err, "not found in memory")
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

	// Once kingsite is gone, the server starts, and the sites restored and kept are replicated again
	_, err = controlDB.Queries.DeleteSites(ctx, controldb.DeleteSitesParams{JsonEach: `["kingsite"]`})
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	controlDB, err = testutils.OpenReplicated(t, cfg, s3)
	require.NoError(t, err)
	site, _, err = controlDB.GetSite(ctx, "kostsite")
	require.NoError(t, err)
	addReading(t, site, "after the restore")
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.Remove(sitedb.GetPath("kostsite", cfg)))
	require.NoError(t, os.RemoveAll(filepath.Join(filepath.Dir(sitedb.GetPath("kostsite", cfg)), ".kostsite.sqlite-litestream")))
	controlDB, err = testutils.OpenReplicated(t, cfg, s3)
	require.NoError(t, err)
	requireReading(t, controlDB, "kostsite", "after the restore")
}

// A control DB missing at start is restored from its replica, with its sites
func TestControlDBRestoresFromItsReplica(t *testing.T) {
	cfg, s3 := testutils.ReplicatedConfig(t)
	controlDB, err := testutils.OpenReplicated(t, cfg, s3)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(t.Context(), "kostsite")
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.RemoveAll(cfg.DBDir()))

	found, err := replica.New(cfg, s3).Restore(t.Context(), "control", controldb.GetPath(cfg))
	require.NoError(t, err)
	require.True(t, found)
	controlDB, err = testutils.OpenReplicated(t, cfg, s3)
	require.NoError(t, err, "the sites are restored too")
	_, _, err = controlDB.GetSite(t.Context(), "kostsite")
	require.NoError(t, err)
}

// A live site whose DB file is missing at start and is not restored stops the start: without S3, or when its restored
// DB cannot be opened, as when its migration fails, as a site that is there and cannot be opened does, so a release
// that cannot open it is rolled back. The control admin sees each as a site_db_missing event.
func TestRestoreMissingSite(t *testing.T) {
	tests := []struct {
		name      string
		s3Working bool
		// dirty is a site dirty at a migration it already has, so migrating it again fails, as one that keeps failing does
		dirty bool
		err   error
		event string
	}{
		{name: "without S3 the start stops", err: db.ErrSitesMissing, event: `{"restored":false,"reason":"s3_not_working"}`},
		{name: "restored but cannot be opened, the start stops", s3Working: true, dirty: true, err: db.ErrRestoredSiteCannotOpen,
			event: `{"restored":false,"reason":"restore_failed"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, s3 := testutils.ReplicatedConfig(t)
			ctx, siteID := t.Context(), "kostsite"
			controlDB, err := testutils.OpenReplicated(t, cfg, s3)
			require.NoError(t, err)
			_, err = controlDB.CreateSiteWithID(ctx, siteID)
			require.NoError(t, err)
			if tt.dirty {
				site, _, err := controlDB.GetSite(ctx, siteID)
				require.NoError(t, err)
				_, err = site.SQLiteDB.Exec("UPDATE schema_migrations SET dirty = 1")
				require.NoError(t, err)
			}
			require.NoError(t, controlDB.Close())
			require.NoError(t, os.Remove(sitedb.GetPath(siteID, cfg)))

			cfg.Status.S3Working = tt.s3Working
			controlDB, err = testutils.OpenReplicated(t, cfg, s3)
			require.ErrorIs(t, err, tt.err)
			require.ErrorContains(t, err, siteID)
			_, _, openErr := controlDB.GetSite(ctx, siteID)
			require.ErrorContains(t, openErr, "not found in memory")
			found, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Type: new("site_db_missing"), Limit: 10})
			require.NoError(t, err)
			require.Len(t, found, 1)
			require.JSONEq(t, tt.event, *found[0].Attributes)
		})
	}
}
