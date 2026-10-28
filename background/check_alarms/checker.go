// Package check_alarms checks each identity's alarms when a site's glucose readings change, and periodically for the
// alarms that wait on time, and sends what it finds through the server's events as alarm events
package check_alarms

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/settings"
)

// maxReadings is how many of the newest readings a check reads, enough to cover the longest alarm, 120 minutes, at a
// reading a minute from more than one source
const maxReadings = 500

type Checker struct {
	controlDB *db.ControlDB
	events    *events.SiteBus
	now       func() time.Time

	// pending are the sites whose readings changed, or that were snoozed, since they were last checked
	pendingLock sync.Mutex
	pending     map[string]bool
	snoozes     []snooze

	// checkLock lets one check run at a time, as they share the identities' alarms kept in the live DB
	checkLock sync.Mutex
}

// snooze is an identity's snooze, kept for the next check
type snooze struct {
	siteID string
	alarms.Snoozed
}

// SetupCron checks the sites whose readings changed every 5 seconds, and every site with alarms every minute
func SetupCron(c cron.CronScheduler, config *config.ServerConfig, controlDB *db.ControlDB) *Checker {
	checker := NewChecker(config, controlDB)
	c.AddFunc("CheckAlarmsOfNewReadings", "@every 5s", func() error { return checker.Check(context.Background(), false) })
	c.AddFunc("CheckAlarms", "@every 60s", func() error { return checker.Check(context.Background(), true) })
	return checker
}

// NewChecker is a checker that notes each site whose glucose readings change, to check it next
func NewChecker(config *config.ServerConfig, controlDB *db.ControlDB) *Checker {
	checker := &Checker{
		controlDB: controlDB, events: config.Events, now: time.Now,
		pending: map[string]bool{},
	}
	// A handler must not block its publisher, so it only notes the site, and any snooze, for the next check
	config.Events.SubscribeAll(func(event events.Event) {
		_, reading := event.Model.(*sitedb.GlucoseReading)
		snoozed, isSnooze := event.Attributes.(alarms.Snoozed)
		if !reading && !(event.Name == events.AlarmSnoozed && isSnooze) {
			return
		}
		checker.pendingLock.Lock()
		defer checker.pendingLock.Unlock()
		checker.pending[event.SiteID] = true
		if isSnooze {
			checker.snoozes = append(checker.snoozes, snooze{event.SiteID, snoozed})
		}
	})
	return checker
}

// Check checks the sites whose readings changed and, with all, every site with an alarm turned on or still kept
func (c *Checker) Check(ctx context.Context, all bool) error {
	c.pendingLock.Lock()
	sites, snoozes := c.pending, c.snoozes
	c.pending, c.snoozes = map[string]bool{}, nil
	c.pendingLock.Unlock()

	c.checkLock.Lock()
	defer c.checkLock.Unlock()
	live := c.controlDB.Live.Queries
	var errs []error
	for _, s := range snoozes {
		for _, t := range s.Types {
			errs = append(errs, live.SnoozeAlarm(ctx, livedb.SnoozeAlarmParams{SiteID: s.siteID, IdentityID: s.IdentityID, Type: t, Until: s.Until}))
		}
	}
	if all {
		withAlarms, err := c.controlDB.Queries.GetSiteIDsWithAlarms(ctx)
		if err != nil {
			return err
		}
		for _, siteID := range withAlarms {
			sites[siteID] = true
		}
		// A site whose alarms were all turned off is checked once more, to clear them
		kept, err := live.AlarmSiteIDs(ctx)
		if err != nil {
			return err
		}
		for _, siteID := range kept {
			sites[siteID] = true
		}
	}
	for siteID := range sites {
		errs = append(errs, c.checkSite(ctx, siteID, c.now().UnixMilli()))
	}
	return errors.Join(errs...)
}

// checkSite checks each of a site's enabled identities' alarms, forgetting those of any identity gone
func (c *Checker) checkSite(ctx context.Context, siteID string, now int64) error {
	live := c.controlDB.Live.Queries
	keptRows, err := live.SiteAlarms(ctx, livedb.SiteAlarmsParams{SiteID: siteID})
	if err != nil {
		return err
	}
	kept := map[int64]keptAlarms{}
	for _, row := range keptRows {
		if kept[row.IdentityID] == nil {
			kept[row.IdentityID] = keptAlarms{}
		}
		kept[row.IdentityID][row.Type] = row
	}

	siteDB, site, err := c.controlDB.GetSite(siteID)
	if err != nil || site == nil || siteDB == nil {
		return errors.Join(err, c.forget(ctx, kept, nil))
	}
	identities, err := c.controlDB.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteID})
	if err != nil {
		return err
	}

	turnedOn := map[int64][]alarms.Alarm{}
	for _, identity := range identities {
		on := alarms.FromSettings(settings.NewCombinedSettingsMap(site, identity))
		if identity.Enabled && (len(on) > 0 || kept[identity.ID] != nil) {
			turnedOn[identity.ID] = on
		}
	}
	if err := c.forget(ctx, kept, turnedOn); err != nil || len(turnedOn) == 0 {
		return err
	}

	rows, err := siteDB.Queries.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{TypeEq: new("sgv"), Limit: maxReadings})
	if err != nil {
		return err
	}
	readings := make([]alarms.Reading, len(rows))
	for i, row := range rows {
		readings[i] = alarms.Reading{At: row.OccurredAt, Mgdl: row.Mgdl}
	}

	var errs []error
	for identityID, on := range turnedOn {
		last := kept[identityID]
		result := alarms.Check(on, last.state(), readings, now)
		checked := alarms.Checked{IdentityID: identityID, Notifications: result.Notifications, Active: types(result.Active), Alerting: types(result.Alerting)}
		lastActive, lastAlerting := last.types()
		if len(checked.Notifications) > 0 || !slices.Equal(checked.Active, lastActive) || !slices.Equal(checked.Alerting, lastAlerting) {
			c.events.Publish(siteID, events.Event{Name: events.Alarm, SiteID: siteID, Attributes: checked, StartedAt: time.UnixMilli(now)})
		}
		for _, t := range alarms.Types {
			s, keep := result.State[t]
			if keep {
				errs = append(errs, live.UpsertAlarm(ctx, livedb.UpsertAlarmParams{SiteID: siteID, IdentityID: identityID, Type: t,
					ActiveSince: s.ActiveSince, NotifiedAt: s.NotifiedAt, SnoozedUntil: s.SnoozedUntil, Alerting: slices.Contains(checked.Alerting, t)}))
			} else if last[t] != nil {
				errs = append(errs, live.DeleteAlarm(ctx, livedb.DeleteAlarmParams{SiteID: siteID, IdentityID: identityID, Type: t}))
			}
		}
	}
	return errors.Join(errs...)
}

// keptAlarms are an identity's alarms kept in the live DB from its last check, by type
type keptAlarms map[alarms.Type]*livedb.Alarm

// state is the kept alarms as the state a check is given
func (k keptAlarms) state() alarms.State {
	state := alarms.State{}
	for t, row := range k {
		state[t] = alarms.AlarmState{ActiveSince: row.ActiveSince, NotifiedAt: row.NotifiedAt, SnoozedUntil: row.SnoozedUntil}
	}
	return state
}

// types are the kept alarms that were active and alerting, in the order a check lists them
func (k keptAlarms) types() (active, alerting []alarms.Type) {
	for _, t := range alarms.Types {
		if row := k[t]; row != nil && row.ActiveSince != 0 {
			active = append(active, t)
			if row.Alerting {
				alerting = append(alerting, t)
			}
		}
	}
	return active, alerting
}

// forget deletes the kept alarms of a site's identities that are not to be checked
func (c *Checker) forget(ctx context.Context, kept map[int64]keptAlarms, checked map[int64][]alarms.Alarm) error {
	var errs []error
	for identityID, rows := range kept {
		if _, ok := checked[identityID]; ok {
			continue
		}
		for _, row := range rows {
			errs = append(errs, c.controlDB.Live.Queries.DeleteAlarm(ctx, livedb.DeleteAlarmParams{SiteID: row.SiteID, IdentityID: identityID, Type: row.Type}))
		}
	}
	return errors.Join(errs...)
}

func types(of []alarms.Alarm) []alarms.Type {
	var types []alarms.Type
	for _, alarm := range of {
		types = append(types, alarm.Type())
	}
	return types
}
