package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/config"
)

// LogsDB keeps logs, events and errors from around the server, each optionally about a site. It is not backed up,
// and Trim keeps only each table's newest rows.
type LogsDB struct {
	*logsdb.DB
}

// The levels a log can be
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

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
// never fails what is being recorded. The site ID must come from the site context, never from a request.

// Log records a message at a level, e.g. LogLevelInfo
func (l *LogsDB) Log(ctx context.Context, siteID *string, level string, message string, attributes map[string]any) {
	l.record("log", l.Queries.InsertLog(ctx, logsdb.InsertLogParams{SiteID: siteID, Level: level, Message: truncate(message), Attributes: attributesJSON(attributes)}))
}

// Event records something that happened, e.g. site_created
func (l *LogsDB) Event(ctx context.Context, siteID *string, name string, attributes map[string]any) {
	l.record("event", l.Queries.InsertEvent(ctx, logsdb.InsertEventParams{SiteID: siteID, Name: name, Attributes: attributesJSON(attributes)}))
}

// Error records an error
func (l *LogsDB) Error(ctx context.Context, siteID *string, err error, attributes map[string]any) {
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
func attributesJSON(attributes map[string]any) *string {
	if len(attributes) == 0 {
		return nil
	}
	raw, err := json.Marshal(attributes)
	if err != nil || len(raw) > maxAttributesLength {
		raw = []byte(`{"truncated":true}`)
	}
	value := string(raw)
	return &value
}
