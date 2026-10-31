package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/geoip"
)

// Every event and error the server records has its own method here, with typed attributes, so what is in the logs
// DB is known. A site ID must come from the site context, never from a request. An event that takes time is recorded
// with when it started, and finishes when it is recorded, so it can be written in a defer.

// ErrorKind is what failed in an error, the kind in its attributes
type ErrorKind string

const (
	ErrorKindServer          ErrorKind = "server"
	ErrorKindSync            ErrorKind = "sync"
	ErrorKindSiteBatchBackup ErrorKind = "site_batch_backup"
	ErrorKindControlBackup   ErrorKind = "control_backup"
)

// ErrorKinds are every error's kind, to pick from
var ErrorKinds = []string{
	string(ErrorKindServer), string(ErrorKindSync), string(ErrorKindSiteBatchBackup), string(ErrorKindControlBackup),
}

// LogLevel is how much a log matters
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

///////
// EVENTS
///////

func (l *LogsDB) SiteCreatedEvent(ctx context.Context, siteID string) {
	l.event(ctx, siteID, events.SiteCreated, nil)
}

// SiteAdminRecoveredEvent records the site's backup key giving its admin new credentials, which its admins can see
func (l *LogsDB) SiteAdminRecoveredEvent(ctx context.Context, siteID string) {
	l.event(ctx, siteID, events.SiteAdminRecovered, nil)
}

type IdentityCredentialsChangedAttributes struct {
	IdentityName string `json:"identity_name"`
	// Credentials are what changed: password, totp, or password_and_totp
	Credentials string `json:"credentials"`
	By          string `json:"by"`
}

// IdentityCredentialsChangedEvent records an identity getting new credentials, by a control identity or the backup key,
// which signs out whoever had the old ones
func (l *LogsDB) IdentityCredentialsChangedEvent(ctx context.Context, siteID string, identityName string, credentials string, by string) {
	l.event(ctx, siteID, events.IdentityCredentialsChanged, IdentityCredentialsChangedAttributes{IdentityName: identityName, Credentials: credentials, By: by})
}

type SignInAttributes struct {
	IdentityName string `json:"identity_name,omitempty"`
	// From is where the sign in is estimated to be from, by its IP address
	From geoip.Location `json:"from"`
}

// SignedInEvent records an identity signing in to its site, or with no siteID a control identity to the control site
func (l *LogsDB) SignedInEvent(ctx context.Context, siteID string, identityName string, from geoip.Location) {
	l.event(ctx, siteID, events.SignedIn, SignInAttributes{IdentityName: identityName, From: from})
}

// SignInFailedEvent records a refused sign in, without the name tried, which may not be an identity's
func (l *LogsDB) SignInFailedEvent(ctx context.Context, siteID string, from geoip.Location) {
	l.event(ctx, siteID, events.SignInFailed, SignInAttributes{From: from})
}

// ResourceHighEvent records one of the server's resources going over its alert level, an event of no one site
func (l *LogsDB) ResourceHighEvent(ctx context.Context, level config.ResourceLevel) {
	l.event(ctx, "", events.ResourceHigh, level)
}

// ResourceRecoveredEvent records one of the server's resources coming back under its alert level
func (l *LogsDB) ResourceRecoveredEvent(ctx context.Context, level config.ResourceLevel) {
	l.event(ctx, "", events.ResourceRecovered, level)
}

type SiteDisabledAttributes struct {
	By string `json:"by"`
}

// SiteDisabledEvent records a control identity disabling a site
func (l *LogsDB) SiteDisabledEvent(ctx context.Context, siteID string, by string) {
	l.event(ctx, siteID, events.SiteDisabled, SiteDisabledAttributes{By: by})
}

type SiteArchivedAttributes struct {
	By string `json:"by"`
}

// SiteArchivedEvent records a control identity archiving a site, which started at started
func (l *LogsDB) SiteArchivedEvent(ctx context.Context, siteID string, by string, started time.Time) {
	l.finishedEvent(ctx, siteID, "", events.SiteArchived, started, SiteArchivedAttributes{By: by})
}

// SiteHydratedEvent records a control identity bringing an archived site back, which started at started
func (l *LogsDB) SiteHydratedEvent(ctx context.Context, siteID string, by string, started time.Time) {
	l.finishedEvent(ctx, siteID, "", events.SiteHydrated, started, SiteArchivedAttributes{By: by})
}

type SiteTimezoneChangedAttributes struct {
	By   string `json:"by"`
	From string `json:"from"`
	To   string `json:"to"`
}

// SiteTimezoneChangedEvent records a site's admin changing its own timezone, from one to another
func (l *LogsDB) SiteTimezoneChangedEvent(ctx context.Context, siteID string, by string, from string, to string) {
	l.event(ctx, siteID, events.SiteTimezoneChanged, SiteTimezoneChangedAttributes{By: by, From: from, To: to})
}

// SiteWeekStartChangedEvent records a site's admin changing its own week start day, from one to another
func (l *LogsDB) SiteWeekStartChangedEvent(ctx context.Context, siteID string, by string, from string, to string) {
	l.event(ctx, siteID, events.SiteWeekStartChanged, SiteTimezoneChangedAttributes{By: by, From: from, To: to})
}

type SiteTruncatedAttributes struct {
	By           string `json:"by"`
	BeforeSizeKB *int64 `json:"before_size_kb"`
	AfterSizeKB  *int64 `json:"after_size_kb"`
}

// SiteTruncatedEvent records a control identity truncating a site's DB, which started at started, and its size before
// and after
func (l *LogsDB) SiteTruncatedEvent(ctx context.Context, siteID string, by string, started time.Time, beforeSizeKB *int64, afterSizeKB *int64) {
	l.finishedEvent(ctx, siteID, "", events.SiteTruncated, started, SiteTruncatedAttributes{By: by, BeforeSizeKB: beforeSizeKB, AfterSizeKB: afterSizeKB})
}

type SiteDeletedAttributes struct {
	Reason string `json:"reason"`
}

func (l *LogsDB) SiteDeletedEvent(ctx context.Context, siteID string, reason string) {
	l.event(ctx, siteID, events.SiteDeleted, SiteDeletedAttributes{Reason: reason})
}

type SiteBatchBackedUpAttributes struct {
	Batch     string `json:"batch"`
	SiteCount int    `json:"site_count"`
	Uploaded  bool   `json:"uploaded"`
}

// SiteBatchBackedUpEvent records the sites whose IDs start with batch being backed up from started, and uploaded when
// S3 works
func (l *LogsDB) SiteBatchBackedUpEvent(ctx context.Context, batch string, siteCount int, started time.Time, uploaded bool) {
	l.finishedEvent(ctx, "", batch, events.SiteBatchBackedUp, started, SiteBatchBackedUpAttributes{Batch: batch, SiteCount: siteCount, Uploaded: uploaded})
}

type ControlBackedUpAttributes struct {
	Uploaded bool `json:"uploaded"`
}

// ControlBackedUpEvent records the control DB being backed up from started, and uploaded when S3 works
func (l *LogsDB) ControlBackedUpEvent(ctx context.Context, started time.Time, uploaded bool) {
	l.finishedEvent(ctx, "", "", events.ControlBackedUp, started, ControlBackedUpAttributes{Uploaded: uploaded})
}

type SyncTurnedOffAttributes struct {
	SyncType consts.SyncSourceType `json:"sync_type"`
}

// SyncTurnedOffEvent records a site's sync source being turned off because its first sync failed
func (l *LogsDB) SyncTurnedOffEvent(ctx context.Context, siteID string, syncType consts.SyncSourceType) {
	l.event(ctx, siteID, events.SyncTurnedOff, SyncTurnedOffAttributes{SyncType: syncType})
}

type SyncBackfilledAttributes struct {
	SyncType consts.SyncSourceType `json:"sync_type"`
	From     int64                 `json:"from"`
	To       int64                 `json:"to"`
	Records  int                   `json:"records"`
	Error    *string               `json:"error,omitempty"`
}

// SyncBackfilledEvent records a site's backfill of the records from from to to, which started at started, the new
// records it added and why it stopped early if it did
func (l *LogsDB) SyncBackfilledEvent(ctx context.Context, siteID string, syncType consts.SyncSourceType, started time.Time, from, to int64, records int, err error) {
	attributes := SyncBackfilledAttributes{SyncType: syncType, From: from, To: to, Records: records}
	if err != nil {
		message := truncate(err.Error())
		attributes.Error = &message
	}
	l.finishedEvent(ctx, siteID, "", events.SyncBackfilled, started, attributes)
}

type SyncAddedAttributes struct {
	SyncType consts.SyncSourceType `json:"sync_type"`
	Records  int                   `json:"records"`
}

// SyncAddedEvent records a site's sync adding new records, and how many
func (l *LogsDB) SyncAddedEvent(ctx context.Context, siteID string, syncType consts.SyncSourceType, records int) {
	l.event(ctx, siteID, events.SyncAdded, SyncAddedAttributes{SyncType: syncType, Records: records})
}

type AlarmAttributes struct {
	IdentityID int64  `json:"identity_id"`
	Type       string `json:"type"`
}

type AlarmSnoozedAttributes struct {
	IdentityID int64 `json:"identity_id"`
	// Types are the alarms snoozed, one from the site, and those of a level and group from a Nightscout ack
	Types []string `json:"types"`
	// SnoozedUntil is when the snooze ends, in milliseconds, and By where it came from: site or nightscout
	SnoozedUntil int64  `json:"snoozed_until"`
	By           string `json:"by"`
}

// AlarmCreatedEvent records an identity's alarm becoming active
func (l *LogsDB) AlarmCreatedEvent(ctx context.Context, siteID string, identityID int64, alarmType string) {
	l.event(ctx, siteID, events.AlarmCreated, AlarmAttributes{IdentityID: identityID, Type: alarmType})
}

// AlarmSnoozedEvent records an identity snoozing some of its alarms until a time in milliseconds, from the site or
// Nightscout
func (l *LogsDB) AlarmSnoozedEvent(ctx context.Context, siteID string, identityID int64, types []string, until int64, by string) {
	l.event(ctx, siteID, events.AlarmSnoozed, AlarmSnoozedAttributes{IdentityID: identityID, Types: types, SnoozedUntil: until, By: by})
}

// AlarmDeletedEvent records an identity's alarm stopping being active
func (l *LogsDB) AlarmDeletedEvent(ctx context.Context, siteID string, identityID int64, alarmType string) {
	l.event(ctx, siteID, events.AlarmDeleted, AlarmAttributes{IdentityID: identityID, Type: alarmType})
}

///////
// ERRORS
///////

type ServerErrorAttributes struct {
	Kind      ErrorKind `json:"kind"`
	Status    int       `json:"status"`
	Method    string    `json:"method"`
	Route     string    `json:"route"`
	RequestID string    `json:"request_id"`
}

// ServerError records a response with a 5xx status, by its route's pattern as a path can carry a token
func (l *LogsDB) ServerError(ctx context.Context, siteID *string, status int, method string, route string, requestID string) {
	l.error(ctx, siteID, fmt.Errorf("%d %s", status, http.StatusText(status)), ServerErrorAttributes{Kind: ErrorKindServer, Status: status, Method: method, Route: route, RequestID: requestID})
}

type SyncErrorAttributes struct {
	Kind     ErrorKind             `json:"kind"`
	SyncType consts.SyncSourceType `json:"sync_type"`
}

func (l *LogsDB) SyncError(ctx context.Context, siteID string, syncType consts.SyncSourceType, err error) {
	l.error(ctx, &siteID, err, SyncErrorAttributes{Kind: ErrorKindSync, SyncType: syncType})
}

type SiteBatchBackupErrorAttributes struct {
	Kind  ErrorKind `json:"kind"`
	Batch string    `json:"batch"`
}

func (l *LogsDB) SiteBatchBackupError(ctx context.Context, batch string, err error) {
	l.error(ctx, nil, err, SiteBatchBackupErrorAttributes{Kind: ErrorKindSiteBatchBackup, Batch: batch})
}

type ControlBackupErrorAttributes struct {
	Kind ErrorKind `json:"kind"`
}

func (l *LogsDB) ControlBackupError(ctx context.Context, err error) {
	l.error(ctx, nil, err, ControlBackupErrorAttributes{Kind: ErrorKindControlBackup})
}
