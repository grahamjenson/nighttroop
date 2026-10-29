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
	"nighttroop.t1d.nz/utils/glucose"
	"nighttroop.t1d.nz/utils/settings"
)

// maxReadings is the most readings a check reads, only a cap: a reading every 4 seconds over the longest alarm
const maxReadings = 2000

type Checker struct {
	controlDB *db.ControlDB
	events    *events.SiteBus
	now       func() time.Time

	// pending are the sites whose readings or device statuses changed since they were last checked
	pendingLock sync.Mutex
	pending     map[string]bool

	// checkLock lets one check run at a time, as they share the identities' alarms kept in the live DB
	checkLock sync.Mutex
}

// SetupCron checks the sites whose readings changed every 5 seconds, and every site with alarms every minute
func SetupCron(c cron.CronScheduler, config *config.ServerConfig, controlDB *db.ControlDB) *Checker {
	checker := NewChecker(config, controlDB)
	c.AddFunc("CheckAlarmsOfNewReadings", "@every 5s", func() error { return checker.Check(context.Background(), false) })
	c.AddFunc("CheckAlarms", "@every 60s", func() error { return checker.Check(context.Background(), true) })
	return checker
}

// NewChecker is a checker that notes each site whose glucose readings or device statuses change, to check it next
func NewChecker(config *config.ServerConfig, controlDB *db.ControlDB) *Checker {
	checker := &Checker{
		controlDB: controlDB, events: config.Events, now: time.Now,
		pending: map[string]bool{},
	}
	// A handler must not block its publisher, so it only notes the site for the next check
	config.Events.SubscribeAll(func(event events.Event) {
		switch event.Model.(type) {
		case *sitedb.GlucoseReading, *sitedb.DeviceReport:
		default:
			return
		}
		checker.pendingLock.Lock()
		defer checker.pendingLock.Unlock()
		checker.pending[event.SiteID] = true
	})
	return checker
}

// Check checks the sites whose readings changed or whose alerting alarms were snoozed and, with all, every site with
// an alarm turned on or still kept
func (c *Checker) Check(ctx context.Context, all bool) error {
	c.pendingLock.Lock()
	sites := c.pending
	c.pending = map[string]bool{}
	c.pendingLock.Unlock()

	c.checkLock.Lock()
	defer c.checkLock.Unlock()
	live := c.controlDB.Live.Queries
	snoozed, err := live.SnoozedAlertingSiteIDs(ctx, livedb.SnoozedAlertingSiteIDsParams{Now: c.now().UnixMilli()})
	if err != nil {
		return err
	}
	for _, siteID := range snoozed {
		sites[siteID] = true
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
	var errs []error
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
	if err := giveDeviceStatuses(ctx, siteDB.Queries, turnedOn); err != nil {
		return err
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
	// The same glucose from more than one source is one reading, so it counts once
	cleaned := glucose.CleanReadings(rows)
	readings := make([]alarms.Reading, len(cleaned))
	for i, c := range cleaned {
		readings[i] = alarms.Reading{At: c.Row.OccurredAt, Mgdl: c.Row.Mgdl}
	}

	var errs []error
	for identityID, on := range turnedOn {
		last := kept[identityID]
		result := alarms.Check(on, last.state(), readings, now)
		checked := alarms.Checked{IdentityID: identityID, Notifications: result.Notifications, Alerting: types(result.Alerting), WasAlerting: last.alerting()}
		if len(checked.Notifications) > 0 || !slices.Equal(checked.Alerting, checked.WasAlerting) {
			c.events.Publish(siteID, events.Event{Name: events.Alarm, SiteID: siteID, Attributes: checked, StartedAt: time.UnixMilli(now)})
		}
		for _, t := range alarms.Types {
			s, keep := result.State[t]
			row := &livedb.Alarm{SiteID: siteID, IdentityID: identityID, Type: t, ActiveSince: s.ActiveSince, NotifiedAt: s.NotifiedAt,
				SnoozedUntil: s.SnoozedUntil, Alerting: slices.Contains(checked.Alerting, t)}
			if keep {
				errs = append(errs, live.UpsertAlarm(ctx, livedb.UpsertAlarmParams(*row)))
			} else if last[t] != nil {
				errs = append(errs, live.DeleteAlarm(ctx, livedb.DeleteAlarmParams{SiteID: siteID, IdentityID: identityID, Type: t, SnoozedUntil: last[t].SnoozedUntil}))
			}
			c.publishChange(siteID, last[t], row)
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

// alerting are the kept alarms that were alerting, in the order a check lists them
func (k keptAlarms) alerting() []alarms.Type {
	var alerting []alarms.Type
	for _, t := range alarms.Types {
		if row := k[t]; row != nil && row.Alerting {
			alerting = append(alerting, t)
		}
	}
	return alerting
}

// giveDeviceStatuses gives each device battery alarm the newest status of each device it checks, reading each device
// once however many alarms list it
func giveDeviceStatuses(ctx context.Context, q *sitedb.Queries, turnedOn map[int64][]alarms.Alarm) error {
	statuses := map[string]alarms.DeviceStatus{}
	read := map[string]bool{}
	for _, on := range turnedOn {
		for i, alarm := range on {
			battery, ok := alarm.(alarms.DeviceBatteryAlarm)
			if !ok {
				continue
			}
			for _, device := range battery.Devices {
				if read[device] {
					continue
				}
				read[device] = true
				reports, err := q.DeviceReportsForDevice(ctx, sitedb.DeviceReportsForDeviceParams{Device: &device, Limit: 1})
				if err != nil {
					return err
				}
				if len(reports) > 0 {
					r := reports[0]
					statuses[device] = alarms.DeviceStatus{At: r.OccurredAt, BatteryPercent: r.DeviceBatteryPercent, Charging: r.IsDeviceCharging != nil && *r.IsDeviceCharging}
				}
			}
			battery.Statuses = statuses
			on[i] = battery
		}
	}
	return nil
}

// forget deletes the kept alarms of a site's identities that are not to be checked
func (c *Checker) forget(ctx context.Context, kept map[int64]keptAlarms, checked map[int64][]alarms.Alarm) error {
	var errs []error
	for identityID, rows := range kept {
		if _, ok := checked[identityID]; ok {
			continue
		}
		for _, row := range rows {
			errs = append(errs, c.controlDB.Live.Queries.DeleteAlarm(ctx, livedb.DeleteAlarmParams{SiteID: row.SiteID, IdentityID: identityID, Type: row.Type, SnoozedUntil: row.SnoozedUntil}))
			c.publishChange(row.SiteID, row, &livedb.Alarm{})
		}
	}
	return errors.Join(errs...)
}

// publishChange publishes the change to an alarm as the alarms collection has it, which has only active alarms: it is
// created as it becomes active, updated as it starts or stops alerting, and deleted as it stops being active
func (c *Checker) publishChange(siteID string, last, row *livedb.Alarm) {
	wasActive, active := last != nil && last.ActiveSince != 0, row.ActiveSince != 0
	switch {
	case !wasActive && active:
		c.events.Publish(siteID, events.ModelChange(siteID, events.Created, row))
	case wasActive && !active:
		c.events.Publish(siteID, events.ModelChange(siteID, events.Deleted, last))
	case active && last.Alerting != row.Alerting:
		c.events.Publish(siteID, events.ModelChange(siteID, events.Updated, row))
	}
}

func types(of []alarms.Alarm) []alarms.Type {
	var types []alarms.Type
	for _, alarm := range of {
		types = append(types, alarm.Type())
	}
	return types
}
