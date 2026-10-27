package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/config"
)

// LogsDB keeps logs, events and errors from around the server, each optionally about a site. It is not backed up,
// and Trim keeps only each table's newest rows.
type LogsDB struct {
	*logsdb.DB
	// events are the server's events, which each event recorded here is sent through
	events *events.SiteBus
}

// maxMessageLength and maxAttributesLength are the longest message and attributes JSON the tables keep
const (
	maxMessageLength    = 1024
	maxAttributesLength = 4096
)

func NewLogsDB(ctx context.Context, cfg *config.ServerConfig) (*LogsDB, error) {
	rawDB, err := logsdb.NewDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot open logs db: %w", err)
	}
	return &LogsDB{DB: rawDB, events: cfg.Events}, nil
}

// Recording is best effort: a log, event or error that cannot be written is only logged with slog, so recording
// never fails what is being recorded. Events and errors are recorded with their own methods, in logs_records.go.

// Log records a message at a level, with any attributes
func (l *LogsDB) Log(ctx context.Context, siteID *string, level LogLevel, message string, attributes map[string]any) {
	var attrs any
	if len(attributes) > 0 {
		attrs = attributes
	}
	l.record("log", l.Queries.InsertLog(ctx, logsdb.InsertLogParams{SiteID: siteID, Level: string(level), Message: truncate(message), Attributes: attributesJSON(attrs)}))
}

// publish sends an event through the server's events, and keeps it in the events table when it is persisted
func (l *LogsDB) publish(ctx context.Context, event events.Event) {
	if l.events != nil {
		l.events.Publish(event.SiteID, event)
	}
	if !event.IsPersisted() {
		return
	}
	var finished *int64
	if event.FinishedAt != nil {
		finished = new(event.FinishedAt.UnixMilli())
	}
	l.record("event", l.Queries.InsertEvent(ctx, logsdb.InsertEventParams{SiteID: orNil(event.SiteID), Batch: orNil(event.Batch), Name: string(event.Name), Attributes: attributesJSON(event.Attributes), StartedAt: event.StartedAt.UnixMilli(), FinishedAt: finished}))
}

// event publishes an event about a site, or no one site when siteID is "", that happened now
func (l *LogsDB) event(ctx context.Context, siteID string, name events.Name, attributes any) {
	l.publish(ctx, events.Event{Name: name, SiteID: siteID, Attributes: attributes, StartedAt: time.Now()})
}

// finishedEvent publishes an event, about a site or a batch of sites or neither, that started at started and finishes
// now, so it can be written in a defer
func (l *LogsDB) finishedEvent(ctx context.Context, siteID string, batch string, name events.Name, started time.Time, attributes any) {
	finished := time.Now()
	l.publish(ctx, events.Event{Name: name, SiteID: siteID, Batch: batch, Attributes: attributes, StartedAt: started, FinishedAt: &finished})
}

// orNil is nil for "", for a column that is NULL when not set
func orNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (l *LogsDB) error(ctx context.Context, siteID *string, err error, attributes any) {
	l.record("error", l.Queries.InsertError(ctx, logsdb.InsertErrorParams{SiteID: siteID, Message: truncate(err.Error()), Attributes: attributesJSON(attributes)}))
}

// Trim deletes all but each table's newest maxRows rows
func (l *LogsDB) Trim(ctx context.Context, maxRows int64) error {
	if _, err := l.Queries.TrimLogs(ctx, logsdb.TrimLogsParams{Keep: maxRows}); err != nil {
		return fmt.Errorf("cannot trim logs: %w", err)
	}
	if _, err := l.Queries.TrimEvents(ctx, logsdb.TrimEventsParams{Keep: maxRows}); err != nil {
		return fmt.Errorf("cannot trim events: %w", err)
	}
	if _, err := l.Queries.TrimErrors(ctx, logsdb.TrimErrorsParams{Keep: maxRows}); err != nil {
		return fmt.Errorf("cannot trim errors: %w", err)
	}
	return nil
}

func (l *LogsDB) record(what string, err error) {
	if err != nil {
		slog.Error("Cannot record "+what+" in the logs db", "error", err)
	}
}

// truncate keeps a message within the length the tables keep
func truncate(message string) string {
	if len(message) <= maxMessageLength {
		return message
	}
	return strings.ToValidUTF8(message[:maxMessageLength-3], "") + "..."
}

// attributesJSON is the attributes as JSON, nil when there are none, or a note when they are too long to keep
func attributesJSON(attributes any) *string {
	if attributes == nil {
		return nil
	}
	raw, err := json.Marshal(attributes)
	if err != nil || len(raw) > maxAttributesLength {
		raw = []byte(`{"truncated":true}`)
	}
	value := string(raw)
	return &value
}
