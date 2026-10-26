package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/config"
)

// LogsDB keeps logs, events and errors from around the server, each optionally about a site. It is not backed up,
// and Trim keeps only each table's newest rows.
type LogsDB struct {
	*logsdb.DB
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
	return &LogsDB{DB: rawDB}, nil
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

// event records an event that happened now, with no finish time
func (l *LogsDB) event(ctx context.Context, siteID *string, name EventName, attributes any) {
	l.record("event", l.Queries.InsertEvent(ctx, logsdb.InsertEventParams{SiteID: siteID, Name: string(name), Attributes: attributesJSON(attributes), StartedAt: time.Now().UnixMilli()}))
}

// finishedEvent records an event, about a site or a batch of sites or neither, that started at started and finishes
// now, so it can be written in a defer
func (l *LogsDB) finishedEvent(ctx context.Context, siteID *string, batch *string, name EventName, started time.Time, attributes any) {
	finished := time.Now().UnixMilli()
	l.record("event", l.Queries.InsertEvent(ctx, logsdb.InsertEventParams{SiteID: siteID, Batch: batch, Name: string(name), Attributes: attributesJSON(attributes), StartedAt: started.UnixMilli(), FinishedAt: &finished}))
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
