package check_alarms

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/testutils"
)

// rig is a checker of one site, its clock moved by the test, that adds readings as a write through the site would,
// and keeps the alarm events it sends
type rig struct {
	t       *testing.T
	checker *Checker
	site    string
	db      *sitedb.Queries
	control *controldb.Queries
	now     time.Time
	sent    []alarms.Checked
}

func newRig(t *testing.T) *rig {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	r := &rig{t: t, site: site.SiteDB.Subdomain, db: site.SiteDB.Queries, control: controlDB.Queries, now: time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)}
	r.checker = NewChecker(cfg, controlDB)
	r.checker.now = func() time.Time { return r.now }
	cfg.Events.Subscribe(r.site, func(event events.Event) {
		if event.Name == events.Alarm {
			r.sent = append(r.sent, event.Attributes.(alarms.Checked))
		}
	})
	return r
}

// reading adds a reading at the rig's time and publishes it, as the site's writes do
func (r *rig) reading(mgdl float64) {
	row, err := r.db.CreateGlucoseReading(context.TODO(), sitedb.CreateGlucoseReadingParams{
		ID: fmt.Sprint(r.now.UnixMilli()), Mgdl: mgdl, ReadingType: "sgv", OccurredAt: r.now.UnixMilli(), CreatedByIdentityName: "test",
	})
	require.NoError(r.t, err)
	r.checker.events.Publish(r.site, events.ModelChange(r.site, events.Created, row))
}

// check runs a check, of the changed sites or all, and is what it sent, by identity
func (r *rig) check(all bool) map[int64]alarms.Checked {
	r.sent = nil
	require.NoError(r.t, r.checker.Check(context.TODO(), all))
	sent := map[int64]alarms.Checked{}
	for _, checked := range r.sent {
		sent[checked.IdentityID] = checked
	}
	return sent
}

func (r *rig) identities() []*controldb.SiteIdentity {
	identities, err := r.control.AllSiteIdentities(context.TODO(), controldb.AllSiteIdentitiesParams{SiteID: r.site})
	require.NoError(r.t, err)
	return identities
}

func notified(kind alarms.Kind, alarm alarms.Type, at time.Time) []alarms.Notification {
	return []alarms.Notification{{Type: alarm, Kind: kind, At: at.UnixMilli()}}
}

func TestEachIdentityIsCheckedOnNewReadingsAndPeriodically(t *testing.T) {
	r := newRig(t)
	ctx := context.TODO()
	// The site turns on the urgent low, and one of its identities turns it off for itself
	require.NoError(t, r.control.UpdateSitesSettings(ctx, controldb.UpdateSitesSettingsParams{ID: r.site, AlarmUrgentLowEnabled: new(true)}))
	identities := r.identities()
	require.Len(t, identities, 3, "a site starts with three identities")
	off := identities[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(ctx, controldb.UpdateSiteIdentitySettingsParams{ID: off.ID, SiteID: r.site, AlarmUrgentLowEnabled: new(false)}))
	on := []int64{identities[1].ID, identities[2].ID}

	// Nothing is sent before there is a reading under it
	require.Empty(t, r.check(true))
	r.reading(100)
	require.Empty(t, r.check(false))

	// A reading under it starts it for the identities it is on for, on the check of changed sites
	r.now = r.now.Add(5 * time.Minute)
	r.reading(50)
	sent := r.check(false)
	require.Len(t, sent, 2)
	for _, id := range on {
		require.Equal(t, alarms.Checked{IdentityID: id, Notifications: notified(alarms.Started, alarms.UrgentLow, r.now),
			Alerting: []alarms.Type{alarms.UrgentLow}}, sent[id])
	}
	require.Empty(t, r.check(false), "a site whose readings did not change is not checked again")

	// The periodic check repeats it without a new reading, once its repeat is due
	r.now = r.now.Add(4 * time.Minute)
	require.Empty(t, r.check(true), "nothing changed, so nothing is sent")
	r.now = r.now.Add(time.Minute)
	sent = r.check(true)
	require.Len(t, sent, 2)
	require.Equal(t, notified(alarms.Repeated, alarms.UrgentLow, r.now), sent[on[0]].Notifications)

	// A reading over it clears it
	r.now = r.now.Add(time.Minute)
	r.reading(100)
	sent = r.check(false)
	require.Len(t, sent, 2)
	require.Equal(t, alarms.Checked{IdentityID: on[1], Notifications: notified(alarms.Cleared, alarms.UrgentLow, r.now),
		WasAlerting: []alarms.Type{alarms.UrgentLow}}, sent[on[1]])

	// Once its last notification no longer matters, its state is forgotten
	require.NotEmpty(t, r.kept(), "its last notification still matters")
	r.now = r.now.Add(10 * time.Minute)
	require.Empty(t, r.check(true))
	require.Empty(t, r.kept())
}

func TestNoDataStartsOnThePeriodicCheck(t *testing.T) {
	r := newRig(t)
	identity := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(context.TODO(), controldb.UpdateSiteIdentitySettingsParams{
		ID: identity.ID, SiteID: r.site, AlarmNoDataEnabled: new(true), AlarmNoDataMinutes: new(int64(20)),
	}))
	r.reading(100)
	require.Empty(t, r.check(false))

	r.now = r.now.Add(20 * time.Minute)
	require.Empty(t, r.check(false), "no reading changed")
	sent := r.check(true)
	require.Equal(t, map[int64]alarms.Checked{identity.ID: {IdentityID: identity.ID, Notifications: notified(alarms.Started, alarms.NoData, r.now),
		Alerting: []alarms.Type{alarms.NoData}}}, sent)
}

func TestTurningAlarmsOffClearsThem(t *testing.T) {
	r := newRig(t)
	ctx := context.TODO()
	identity := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(ctx, controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: r.site, AlarmUrgentLowEnabled: new(true)}))
	r.reading(50)
	require.Len(t, r.check(false), 1)

	// With every alarm off the site is no longer one with alarms, but its kept state has it checked once more
	require.NoError(t, r.control.UpdateSiteIdentitySettings(ctx, controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: r.site}))
	sites, err := r.control.GetSiteIDsWithAlarms(ctx)
	require.NoError(t, err)
	require.NotContains(t, sites, r.site)
	r.now = r.now.Add(time.Minute)
	require.Equal(t, map[int64]alarms.Checked{identity.ID: {IdentityID: identity.ID, Notifications: notified(alarms.Cleared, alarms.UrgentLow, r.now),
		WasAlerting: []alarms.Type{alarms.UrgentLow}}}, r.check(true))
}

func TestSitesWithoutAlarmsAreNotChecked(t *testing.T) {
	r := newRig(t)
	sites, err := r.control.GetSiteIDsWithAlarms(context.TODO())
	require.NoError(t, err)
	require.NotContains(t, sites, r.site)
	r.reading(40)
	require.Empty(t, r.check(false))
	require.Empty(t, r.check(true))
	require.Empty(t, r.kept())
}

func TestALowLastsOverEarlierReadings(t *testing.T) {
	r := newRig(t)
	ctx := context.TODO()
	require.NoError(t, r.control.UpdateSitesSettings(ctx, controldb.UpdateSitesSettingsParams{ID: r.site, AlarmLowEnabled: new(true), AlarmLow: new(int64(72)), AlarmLowMinutes: new(int64(5))}))
	r.reading(70)
	require.Empty(t, r.check(false), "one reading under it is not enough")
	r.now = r.now.Add(5 * time.Minute)
	r.reading(70)
	sent := r.check(false)
	require.Len(t, sent, 3, "it has lasted 5 minutes, for each of the site's identities")
	require.Equal(t, notified(alarms.Started, alarms.Low, r.now), sent[r.identities()[0].ID].Notifications)
}

func TestDisabledIdentitiesAreNotChecked(t *testing.T) {
	r := newRig(t)
	ctx := context.TODO()
	require.NoError(t, r.control.UpdateSitesSettings(ctx, controldb.UpdateSitesSettingsParams{ID: r.site, AlarmUrgentLowEnabled: new(true)}))
	disabled := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentityEnabled(ctx, controldb.UpdateSiteIdentityEnabledParams{SiteID: r.site, ID: disabled.ID, Enabled: false}))
	r.reading(50)
	sent := r.check(false)
	require.Len(t, sent, 2)
	require.NotContains(t, sent, disabled.ID)
}

func TestASnoozeIsTheIdentitysOwn(t *testing.T) {
	r := newRig(t)
	require.NoError(t, r.control.UpdateSitesSettings(context.TODO(), controldb.UpdateSitesSettingsParams{ID: r.site, AlarmUrgentLowEnabled: new(true)}))
	identities := r.identities()
	snoozer, other := identities[0].ID, identities[1].ID
	r.reading(50)
	require.Len(t, r.check(false), 3)

	// The snooze, written as the site and a socket's ack do, stops it alerting on the next check of changed sites
	r.now = r.now.Add(time.Minute)
	r.snooze(snoozer, alarms.UrgentLow, r.now.Add(30*time.Minute))
	sent := r.check(false)
	require.Equal(t, map[int64]alarms.Checked{snoozer: {IdentityID: snoozer, WasAlerting: []alarms.Type{alarms.UrgentLow}}}, sent,
		"only the snoozer's alarm stops alerting")

	// The others still repeat, and the snoozer's does not until its snooze ends
	r.now = r.now.Add(4 * time.Minute)
	sent = r.check(true)
	require.Len(t, sent, 2)
	require.NotContains(t, sent, snoozer)
	require.Equal(t, notified(alarms.Repeated, alarms.UrgentLow, r.now), sent[other].Notifications)
	r.now = r.now.Add(26 * time.Minute)
	require.Equal(t, notified(alarms.Repeated, alarms.UrgentLow, r.now), r.check(true)[snoozer].Notifications)
}

// snooze snoozes an identity's alarm until a time, as the site and a socket's ack do
func (r *rig) snooze(identityID int64, alarm alarms.Type, until time.Time) {
	require.NoError(r.t, r.checker.controlDB.Live.Queries.SnoozeAlarm(context.TODO(), livedb.SnoozeAlarmParams{
		SiteID: r.site, IdentityID: identityID, Type: alarm, Until: until.UnixMilli()}))
}

// kept are the site's alarms the checker keeps between checks
func (r *rig) kept() []*livedb.Alarm {
	rows, err := r.checker.controlDB.Live.Queries.SiteAlarms(context.TODO(), livedb.SiteAlarmsParams{SiteID: r.site})
	require.NoError(r.t, err)
	return rows
}

func TestEachChangeToAnActiveAlarmIsPublished(t *testing.T) {
	r := newRig(t)
	identity := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(context.TODO(), controldb.UpdateSiteIdentitySettingsParams{ID: identity.ID, SiteID: r.site, AlarmUrgentLowEnabled: new(true)}))
	var changes []string
	r.checker.events.Subscribe(r.site, func(event events.Event) {
		if row, ok := event.Model.(*livedb.Alarm); ok {
			changes = append(changes, fmt.Sprintf("%s %s alerting:%v", event.Action, row.Type, row.Alerting))
		}
	})

	r.reading(50)
	r.check(false)
	r.now = r.now.Add(time.Minute)
	r.snooze(identity.ID, alarms.UrgentLow, r.now.Add(30*time.Minute))
	r.check(false)
	r.now = r.now.Add(time.Minute)
	r.reading(100)
	r.check(false)
	require.Equal(t, []string{"created urgent_low alerting:true", "updated urgent_low alerting:false", "deleted urgent_low alerting:false"}, changes,
		"it is created as it becomes active, updated as it stops alerting, and deleted as it stops being active")
	require.NotEmpty(t, r.kept(), "its snooze is kept after it clears")
}

func TestASnoozeDuringACheckIsKept(t *testing.T) {
	r := newRig(t)
	live := r.checker.controlDB.Live.Queries
	identity := r.identities()[0].ID
	require.NoError(t, r.control.UpdateSiteIdentitySettings(context.TODO(), controldb.UpdateSiteIdentitySettingsParams{ID: identity, SiteID: r.site, AlarmUrgentLowEnabled: new(true)}))
	r.reading(50)
	require.Len(t, r.check(false), 1)
	started := r.kept()[0]

	// A snooze lands after a check read the alarm, before it writes what it found from the alarm unsnoozed: that it
	// still alerts, as it sends
	r.now = r.now.Add(time.Minute)
	until := r.now.Add(30 * time.Minute).UnixMilli()
	r.snooze(identity, alarms.UrgentLow, r.now.Add(30*time.Minute))
	stale := livedb.UpsertAlarmParams(*started)
	require.NoError(t, live.UpsertAlarm(context.TODO(), stale))
	require.NoError(t, live.DeleteAlarm(context.TODO(), livedb.DeleteAlarmParams{SiteID: r.site, IdentityID: identity, Type: alarms.UrgentLow}))
	kept := *started
	kept.SnoozedUntil = until
	require.Equal(t, []*livedb.Alarm{&kept}, r.kept(), "the snooze is kept, and the alarm is not deleted")

	// The next check of changed sites finds it snoozed while alerting, so stops it alerting for the identity's sockets
	require.Equal(t, map[int64]alarms.Checked{identity: {IdentityID: identity, WasAlerting: []alarms.Type{alarms.UrgentLow}}}, r.check(false))
}

// No data stays on over a gap longer than the longest alarm, as the check reads the newest reading however old it is
func TestNoDataLastsOverALongGap(t *testing.T) {
	r := newRig(t)
	identity := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(context.TODO(), controldb.UpdateSiteIdentitySettingsParams{
		ID: identity.ID, SiteID: r.site, AlarmNoDataEnabled: new(true), AlarmNoDataMinutes: new(int64(20)),
	}))
	r.reading(100)
	r.check(false)
	r.now = r.now.Add(20 * time.Minute)
	require.Equal(t, notified(alarms.Started, alarms.NoData, r.now), r.check(true)[identity.ID].Notifications)

	r.now = r.now.Add(3 * time.Hour)
	sent := r.check(true)[identity.ID]
	require.Equal(t, notified(alarms.Repeated, alarms.NoData, r.now), sent.Notifications)
	require.Equal(t, []alarms.Type{alarms.NoData}, sent.Alerting)
}

// A high lasting the longest an alarm can wait starts it, however many readings that is, as from several sources
func TestALongHighAtManyReadingsAMinute(t *testing.T) {
	r := newRig(t)
	identity := r.identities()[0]
	require.NoError(t, r.control.UpdateSiteIdentitySettings(context.TODO(), controldb.UpdateSiteIdentitySettingsParams{
		ID: identity.ID, SiteID: r.site, AlarmHighEnabled: new(true), AlarmHigh: new(int64(180)), AlarmHighMinutes: new(int64(120)),
	}))
	// Ten readings a minute for 120 minutes
	for range 120*10 + 1 {
		r.reading(250)
		r.now = r.now.Add(6 * time.Second)
	}
	r.now = r.now.Add(-6 * time.Second)
	require.Equal(t, []alarms.Type{alarms.High}, r.check(false)[identity.ID].Alerting)
}
