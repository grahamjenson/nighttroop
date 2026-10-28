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
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/settings"
)

// maxReadings is the most readings a check reads, only a cap: a reading every 4 seconds over the longest alarm
const maxReadings = 2000

// key is an identity's alarms, kept by its site and ID
type key struct {
	siteID     string
	identityID int64
}

type Checker struct {
	controlDB *db.ControlDB
	events    *events.SiteBus
	now       func() time.Time

	// pending are the sites whose readings changed, or that were snoozed, since they were last checked
	pendingLock sync.Mutex
	pending     map[string]bool
	snoozes     []snooze

	// checkLock lets one check run at a time, as they share the identities' states and what they last sent
	checkLock sync.Mutex
	states    map[key]alarms.State
	sent      map[key]alarms.Checked
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
		pending: map[string]bool{}, states: map[key]alarms.State{}, sent: map[key]alarms.Checked{},
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
	for _, s := range snoozes {
		k := key{s.siteID, s.IdentityID}
		for _, t := range s.Types {
			c.states[k] = c.states[k].Snooze(t, s.Until)
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
		for k := range c.states {
			sites[k.siteID] = true
		}
	}
	var errs []error
	for siteID := range sites {
		errs = append(errs, c.checkSite(ctx, siteID, c.now().UnixMilli()))
	}
	return errors.Join(errs...)
}

// checkSite checks each of a site's enabled identities' alarms, forgetting those of any identity gone
func (c *Checker) checkSite(ctx context.Context, siteID string, now int64) error {
	siteDB, site, err := c.controlDB.GetSite(siteID)
	if err != nil || site == nil || siteDB == nil {
		c.forget(siteID, nil)
		return err
	}
	identities, err := c.controlDB.Queries.AllSiteIdentities(ctx, controldb.AllSiteIdentitiesParams{SiteID: siteID})
	if err != nil {
		return err
	}

	turnedOn := map[int64][]alarms.Alarm{}
	for _, identity := range identities {
		k := key{siteID, identity.ID}
		on := alarms.FromSettings(settings.NewCombinedSettingsMap(site, identity))
		if identity.Enabled && (len(on) > 0 || c.states[k] != nil) {
			turnedOn[identity.ID] = on
		}
	}
	c.forget(siteID, turnedOn)
	if len(turnedOn) == 0 {
		return nil
	}

	// The newest reading, however old, as no data is measured from it, then those back to the longest an alarm can
	// wait before it, as an alarm lasts by its readings, not by the time now
	rows, err := siteDB.Queries.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{TypeEq: new("sgv"), Limit: 1})
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		since := rows[0].OccurredAt - settings.MaxAlarmMinutes*time.Minute.Milliseconds()
		if rows, err = siteDB.Queries.SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{TypeEq: new("sgv"), DateGTE: &since, Limit: maxReadings}); err != nil {
			return err
		}
	}
	readings := make([]alarms.Reading, len(rows))
	for i, row := range rows {
		readings[i] = alarms.Reading{At: row.OccurredAt, Mgdl: row.Mgdl}
	}

	for identityID, on := range turnedOn {
		k := key{siteID, identityID}
		result := alarms.Check(on, c.states[k], readings, now)
		checked := alarms.Checked{IdentityID: identityID, Notifications: result.Notifications, Active: types(result.Active), Alerting: types(result.Alerting)}
		last := c.sent[k]
		if len(checked.Notifications) > 0 || !slices.Equal(checked.Active, last.Active) || !slices.Equal(checked.Alerting, last.Alerting) {
			c.events.Publish(siteID, events.Event{Name: events.Alarm, SiteID: siteID, Attributes: checked, StartedAt: time.UnixMilli(now)})
		}
		c.states[k], c.sent[k] = result.State, checked
		if len(result.State) == 0 {
			delete(c.states, k)
			delete(c.sent, k)
		}
	}
	return nil
}

// forget drops the states of a site's identities that are not to be checked
func (c *Checker) forget(siteID string, checked map[int64][]alarms.Alarm) {
	for k := range c.states {
		if _, ok := checked[k.identityID]; k.siteID == siteID && !ok {
			delete(c.states, k)
			delete(c.sent, k)
		}
	}
}

func types(of []alarms.Alarm) []alarms.Type {
	var types []alarms.Type
	for _, alarm := range of {
		types = append(types, alarm.Type())
	}
	return types
}
