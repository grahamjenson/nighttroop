package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
	ErrorKindClient          ErrorKind = "client"
	ErrorKindConnector       ErrorKind = "connector"
	ErrorKindSiteBatchBackup ErrorKind = "site_batch_backup"
	ErrorKindControlBackup   ErrorKind = "control_backup"
)

// ErrorKinds are every error's kind, to pick from
var ErrorKinds = []string{
	string(ErrorKindServer), string(ErrorKindClient), string(ErrorKindConnector), string(ErrorKindSiteBatchBackup), string(ErrorKindControlBackup),
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

// SignedInByControlAttributes are the identity a control admin signed in to a site as, and the admin
type SignedInByControlAttributes struct {
	IdentityName string `json:"identity_name"`
	By           string `json:"by"`
}

// SiteDBDownloadedByControlEvent records the control admin by downloading a copy of the site's DB, so its admins can see
// someone outside the site has a copy of their data
func (l *LogsDB) SiteDBDownloadedByControlEvent(ctx context.Context, siteID string, by string) {
	l.event(ctx, siteID, events.SiteDBDownloadedByControl, struct {
		By string `json:"by"`
	}{By: by})
}

// SignedInByControlEvent records the control admin by signing in to the site as its identity, so its admins can see they did
func (l *LogsDB) SignedInByControlEvent(ctx context.Context, siteID string, identityName string, by string) {
	l.event(ctx, siteID, events.SignedInByControl, SignedInByControlAttributes{IdentityName: identityName, By: by})
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

// SummariesRegenerationRequestedEvent records a site's admin asking for its summaries to be built again
func (l *LogsDB) SummariesRegenerationRequestedEvent(ctx context.Context, siteID string, by string) {
	l.event(ctx, siteID, events.SummariesRegenerationRequested, SiteArchivedAttributes{By: by})
}

type SiteTruncatedAttributes struct {
	By           string `json:"by"`
	BeforeSizeKB *int64 `json:"before_size_kb"`
	AfterSizeKB  *int64 `json:"after_size_kb"`
}

// SiteTruncatedEvent records by, a control identity, a site admin or the size limit, truncating a site's DB, which
// started at started, and its size before and after
func (l *LogsDB) SiteTruncatedEvent(ctx context.Context, siteID string, by string, started time.Time, beforeSizeKB *int64, afterSizeKB *int64) {
	l.finishedEvent(ctx, siteID, "", events.SiteTruncated, started, SiteTruncatedAttributes{By: by, BeforeSizeKB: beforeSizeKB, AfterSizeKB: afterSizeKB})
}

type SiteDeletedAttributes struct {
	Reason string `json:"reason"`
}

func (l *LogsDB) SiteDeletedEvent(ctx context.Context, siteID string, reason string) {
	l.event(ctx, siteID, events.SiteDeleted, SiteDeletedAttributes{Reason: reason})
}

type SiteDBMissingAttributes struct {
	Restored bool `json:"restored"`
	// Reason is why a site that was not restored was not: s3_not_working, not_in_backup or restore_failed
	Reason string `json:"reason,omitempty"`
}

// SiteDBMissingEvent records a live site's DB file missing at start, and whether it was restored from its batch's
// backup, so the control admin can see a site left out
func (l *LogsDB) SiteDBMissingEvent(ctx context.Context, siteID string, restored bool, reason string) {
	l.event(ctx, siteID, events.SiteDBMissing, SiteDBMissingAttributes{Restored: restored, Reason: reason})
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

type ConnectorTurnedOffAttributes struct {
	ConnectorType consts.ConnectorType `json:"connector_type"`
}

// ConnectorTurnedOffEvent records a site's connector being turned off because its first sync failed
func (l *LogsDB) ConnectorTurnedOffEvent(ctx context.Context, siteID string, connectorType consts.ConnectorType) {
	l.event(ctx, siteID, events.ConnectorTurnedOff, ConnectorTurnedOffAttributes{ConnectorType: connectorType})
}

type ConnectorBackfilledAttributes struct {
	ConnectorType consts.ConnectorType `json:"connector_type"`
	From          int64                `json:"from"`
	To            int64                `json:"to"`
	Records       int                  `json:"records"`
	Error         *string              `json:"error,omitempty"`
}

// ConnectorBackfilledEvent records a site's backfill of the records from from to to, which started at started, the new
// records it added and why it stopped early if it did
func (l *LogsDB) ConnectorBackfilledEvent(ctx context.Context, siteID string, connectorType consts.ConnectorType, started time.Time, from, to int64, records int, err error) {
	attributes := ConnectorBackfilledAttributes{ConnectorType: connectorType, From: from, To: to, Records: records}
	if err != nil {
		message := truncate(err.Error())
		attributes.Error = &message
	}
	l.finishedEvent(ctx, siteID, "", events.ConnectorBackfilled, started, attributes)
}

type ConnectorRecordsAddedAttributes struct {
	ConnectorType consts.ConnectorType `json:"connector_type"`
	Records       int                  `json:"records"`
}

// ConnectorRecordsAddedEvent records a site's sync adding new records, and how many
func (l *LogsDB) ConnectorRecordsAddedEvent(ctx context.Context, siteID string, connectorType consts.ConnectorType, records int) {
	l.event(ctx, siteID, events.ConnectorRecordsAdded, ConnectorRecordsAddedAttributes{ConnectorType: connectorType, Records: records})
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
	Reason    string    `json:"reason,omitempty"`
	// RequestBody is what the request sent, its credentials redacted, to debug it by
	RequestBody map[string]any `json:"request_body,omitempty"`
}

// ServerError records a response with a 5xx status, with why when known, by its route's pattern as a path can carry
// a token
func (l *LogsDB) ServerError(ctx context.Context, siteID *string, attributes ServerErrorAttributes) {
	attributes.Kind = ErrorKindServer
	message := fmt.Sprintf("%d %s", attributes.Status, http.StatusText(attributes.Status))
	if attributes.Reason != "" {
		message += ": " + attributes.Reason
	}
	l.error(ctx, siteID, errors.New(message), attributes)
}

// ClientErrorAttributes are what a request refused for what it sent was, and who sent it: its identity, and the app
// it said it was in its User-Agent
type ClientErrorAttributes struct {
	Kind      ErrorKind `json:"kind"`
	Status    int       `json:"status"`
	Method    string    `json:"method"`
	Route     string    `json:"route"`
	RequestID string    `json:"request_id"`
	Reason    string    `json:"reason,omitempty"`
	Identity  string    `json:"identity,omitempty"`
	Role      string    `json:"role,omitempty"`
	Client    string    `json:"client,omitempty"`
	// RequestBody is what the request sent, its credentials redacted, to debug it by
	RequestBody map[string]any `json:"request_body,omitempty"`
}

// ClientError records a response refusing a request for what it sent, such as a 400, with why, by its route's
// pattern as a path can carry a token
func (l *LogsDB) ClientError(ctx context.Context, siteID *string, attributes ClientErrorAttributes) {
	attributes.Kind = ErrorKindClient
	message := fmt.Sprintf("%d %s", attributes.Status, http.StatusText(attributes.Status))
	if attributes.Reason != "" {
		message += ": " + attributes.Reason
	}
	l.error(ctx, siteID, errors.New(message), attributes)
}

type ConnectorErrorAttributes struct {
	Kind          ErrorKind            `json:"kind"`
	ConnectorType consts.ConnectorType `json:"connector_type"`
	// Record is the record the sync could not read, or the site refused, to debug it by
	Record any `json:"record,omitempty"`
}

func (l *LogsDB) ConnectorError(ctx context.Context, siteID string, connectorType consts.ConnectorType, err error) {
	l.error(ctx, &siteID, err, ConnectorErrorAttributes{Kind: ErrorKindConnector, ConnectorType: connectorType})
}

// connectorRecordBytes is how much of a refused record's JSON is kept on its error
const connectorRecordBytes = 8 * 1024

// ConnectorRecordError records a record a sync could not read, or the site refused, with the record, as an error of
// the site's connector
func (l *LogsDB) ConnectorRecordError(ctx context.Context, siteID string, connectorType consts.ConnectorType, err error, record any) {
	l.error(ctx, &siteID, err, ConnectorErrorAttributes{Kind: ErrorKindConnector, ConnectorType: connectorType, Record: RecordJSON(record)})
}

// RecordJSON is a record as JSON, up to connectorRecordBytes, or the start of its text when it is longer or is not
// JSON, so it fits on an error
func RecordJSON(record any) any {
	raw, err := json.Marshal(record)
	if bytes, ok := record.(json.RawMessage); ok && err != nil {
		raw = bytes
	}
	if err == nil && len(raw) <= connectorRecordBytes {
		return json.RawMessage(raw)
	}
	return map[string]any{"text": strings.ToValidUTF8(string(raw[:min(len(raw), connectorRecordBytes)]), ""), "truncated": len(raw) > connectorRecordBytes}
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
