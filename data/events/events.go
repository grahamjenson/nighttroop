package events

import (
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"
)

// Every event on the server is an Event, sent through one SiteBus under its site's ID, or "" when it is about no one
// site: a change to a site's models, such as glucose_reading.created, or something done to a site or the server,
// such as site_created or sync_backfilled. Only the Persisted ones are also kept, in the logs DB's events table.

// Name is what happened in an event
type Name string

const (
	SiteCreated       Name = "site_created"
	SiteDisabled      Name = "site_disabled"
	SiteDeleted       Name = "site_deleted"
	SiteArchived      Name = "site_archived"
	SiteHydrated      Name = "site_hydrated"
	SiteTruncated     Name = "site_truncated"
	SiteBatchBackedUp Name = "site_batch_backed_up"
	ControlBackedUp   Name = "control_backed_up"
	SyncTurnedOff     Name = "sync_turned_off"
	// SiteDBMissing is a live site's DB file missing at start, restored from its batch's backup or left out
	SiteDBMissing Name = "site_db_missing"
	// SiteAdminRecovered is the site's backup key being used to give its admin new credentials
	SiteAdminRecovered Name = "site_admin_recovered"
	SyncBackfilled     Name = "sync_backfilled"
	SyncAdded          Name = "sync_added"
	// SiteTimezoneChanged is a site's own timezone changing, its Attributes the timezone it was and is
	SiteTimezoneChanged Name = "site_timezone_changed"
	// SiteWeekStartChanged is a site's own week start day changing, its Attributes the day it was and is
	SiteWeekStartChanged Name = "site_week_start_changed"
	// Alarm is an identity's alarms after a check that changed them, its Attributes an alarms.Checked
	Alarm Name = "alarm"
	// AlarmCreated, AlarmSnoozed and AlarmDeleted are an identity's alarm becoming active, being snoozed, and stopping
	// being active, kept so when each happened can be seen, as the live DB's alarms are not
	AlarmCreated Name = "alarm_created"
	AlarmSnoozed Name = "alarm_snoozed"
	AlarmDeleted Name = "alarm_deleted"
	// IdentityCredentialsChanged is an identity getting a new password or TOTP, which refuses its JWTs issued before
	IdentityCredentialsChanged Name = "identity_credentials_changed"
	// SignedIn and SignInFailed are a sign in to a site, or to the control site, and one refused, with where from
	SignedIn     Name = "signed_in"
	SignInFailed Name = "sign_in_failed"
	// SignedInByControl is a control admin signing in to a site as one of its identities, from the site's details
	SignedInByControl Name = "signed_in_by_control"
	// ResourceHigh and ResourceRecovered are one of the server's resources going over its alert level and coming
	// back under it, their Attributes a config.ResourceLevel
	ResourceHigh      Name = "resource_high"
	ResourceRecovered Name = "resource_recovered"
)

// Persisted are the events kept in the logs DB. A model change is not, as the site's DB row is its record.
var Persisted = []Name{SiteCreated, SiteDisabled, SiteDeleted, SiteArchived, SiteHydrated, SiteTruncated, SiteDBMissing, SiteBatchBackedUp, ControlBackedUp,
	SiteAdminRecovered, IdentityCredentialsChanged, SyncTurnedOff, SyncBackfilled, SyncAdded, SiteTimezoneChanged, SiteWeekStartChanged, AlarmCreated, AlarmSnoozed, AlarmDeleted, SignedIn, SignInFailed, SignedInByControl, ResourceHigh, ResourceRecovered}

// PersistedNames are the persisted events' names, to pick from
func PersistedNames() []string {
	names := make([]string, len(Persisted))
	for i, name := range Persisted {
		names[i] = string(name)
	}
	return names
}

// Action is what happened to a model in a model change
type Action string

const (
	Created Action = "created"
	Updated Action = "updated"
	Deleted Action = "deleted"
)

// Event is something that happened on the server
type Event struct {
	Name   Name
	SiteID string
	// Batch is the batch of sites an event is about, such as a backup of the sites whose IDs start with it
	Batch string
	// Attributes are what else is known about it, kept as JSON when it is persisted
	Attributes any
	StartedAt  time.Time
	// FinishedAt is when an event that took time finished
	FinishedAt *time.Time
	// Action and Model are a model change's, published once it is committed. Model is the sitedb row as stored, e.g.
	// *sitedb.GlucoseReading or *sitedb.EventGroupsBundle; a deleted model is the row as it was before.
	Action Action
	Model  any
}

// IsPersisted is whether the event is kept in the logs DB
func (e Event) IsPersisted() bool {
	return slices.Contains(Persisted, e.Name)
}

// ModelChange is the event of a committed change to one of a site's models, named for the model and the action, as
// glucose_reading.created
func ModelChange(siteID string, action Action, model any) Event {
	name := "model"
	if t := reflect.TypeOf(model); t != nil {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		name = snakeCase(t.Name())
	}
	return Event{Name: Name(name + "." + string(action)), SiteID: siteID, StartedAt: time.Now(), Action: action, Model: model}
}

// snakeCase is a Go type's name as words joined by underscores, as glucose_reading for GlucoseReading
func snakeCase(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// SiteBus is every event, subscribed to by site ID
type SiteBus = Bus[string, Event]

func NewSiteBus() *SiteBus { return NewBus[string, Event]() }
