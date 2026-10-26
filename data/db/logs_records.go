package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nighttroop.t1d.nz/utils/consts"
)

// Every event and error the server records has its own method here, with typed attributes, so what is in the logs
// DB is known. A site ID must come from the site context, never from a request. An event that takes time is recorded
// with when it started, and finishes when it is recorded, so it can be written in a defer.

// EventName is what happened in an event
type EventName string

const (
	EventSiteCreated       EventName = "site_created"
	EventSiteDisabled      EventName = "site_disabled"
	EventSiteDeleted       EventName = "site_deleted"
	EventSiteArchived      EventName = "site_archived"
	EventSiteTruncated     EventName = "site_truncated"
	EventSiteBatchBackedUp EventName = "site_batch_backed_up"
	EventControlBackedUp   EventName = "control_backed_up"
	EventSyncTurnedOff     EventName = "sync_turned_off"
	EventSyncBackfilled    EventName = "sync_backfilled"
)

// ErrorKind is what failed in an error, the kind in its attributes
type ErrorKind string

const (
	ErrorKindServer          ErrorKind = "server"
	ErrorKindSync            ErrorKind = "sync"
	ErrorKindSiteBatchBackup ErrorKind = "site_batch_backup"
	ErrorKindControlBackup   ErrorKind = "control_backup"
)

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
	l.event(ctx, &siteID, EventSiteCreated, nil)
}

type SiteDisabledAttributes struct {
	By string `json:"by"`
}

// SiteDisabledEvent records a control identity disabling a site
func (l *LogsDB) SiteDisabledEvent(ctx context.Context, siteID string, by string) {
	l.event(ctx, &siteID, EventSiteDisabled, SiteDisabledAttributes{By: by})
}

type SiteArchivedAttributes struct {
	By string `json:"by"`
}

// SiteArchivedEvent records a control identity archiving a site, which started at started
func (l *LogsDB) SiteArchivedEvent(ctx context.Context, siteID string, by string, started time.Time) {
	l.finishedEvent(ctx, &siteID, EventSiteArchived, started, SiteArchivedAttributes{By: by})
}

type SiteTruncatedAttributes struct {
	By           string `json:"by"`
	BeforeSizeKB *int64 `json:"before_size_kb"`
	AfterSizeKB  *int64 `json:"after_size_kb"`
}

// SiteTruncatedEvent records a control identity truncating a site's DB, which started at started, and its size before
// and after
func (l *LogsDB) SiteTruncatedEvent(ctx context.Context, siteID string, by string, started time.Time, beforeSizeKB *int64, afterSizeKB *int64) {
	l.finishedEvent(ctx, &siteID, EventSiteTruncated, started, SiteTruncatedAttributes{By: by, BeforeSizeKB: beforeSizeKB, AfterSizeKB: afterSizeKB})
}

type SiteDeletedAttributes struct {
	Reason string `json:"reason"`
}

func (l *LogsDB) SiteDeletedEvent(ctx context.Context, siteID string, reason string) {
	l.event(ctx, &siteID, EventSiteDeleted, SiteDeletedAttributes{Reason: reason})
}

type SiteBatchBackedUpAttributes struct {
	Batch     string `json:"batch"`
	SiteCount int    `json:"site_count"`
	Uploaded  bool   `json:"uploaded"`
}

// SiteBatchBackedUpEvent records the sites whose IDs start with batch being backed up from started, and uploaded when
// S3 works
func (l *LogsDB) SiteBatchBackedUpEvent(ctx context.Context, batch string, siteCount int, started time.Time, uploaded bool) {
	l.finishedEvent(ctx, nil, EventSiteBatchBackedUp, started, SiteBatchBackedUpAttributes{Batch: batch, SiteCount: siteCount, Uploaded: uploaded})
}

type ControlBackedUpAttributes struct {
	Uploaded bool `json:"uploaded"`
}

// ControlBackedUpEvent records the control DB being backed up from started, and uploaded when S3 works
func (l *LogsDB) ControlBackedUpEvent(ctx context.Context, started time.Time, uploaded bool) {
	l.finishedEvent(ctx, nil, EventControlBackedUp, started, ControlBackedUpAttributes{Uploaded: uploaded})
}

type SyncTurnedOffAttributes struct {
	SyncType consts.SyncSourceType `json:"sync_type"`
}

// SyncTurnedOffEvent records a site's sync source being turned off because its first sync failed
func (l *LogsDB) SyncTurnedOffEvent(ctx context.Context, siteID string, syncType consts.SyncSourceType) {
	l.event(ctx, &siteID, EventSyncTurnedOff, SyncTurnedOffAttributes{SyncType: syncType})
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
	l.finishedEvent(ctx, &siteID, EventSyncBackfilled, started, attributes)
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
