package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
)

// LogsDB keeps logs, events and errors from around the server, each optionally about a site. It is not backed up,
// and Trim keeps only each table's newest rows.
type LogsDB struct {
	*logsdb.DB
	// events are the server's events, which each event recorded here is sent through
	events *events.SiteBus
	// stop ends writing aslog.Logs, and done is closed once it has written the last of them
	stop, done chan struct{}
	stopOnce   sync.Once
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
	l := &LogsDB{DB: rawDB, events: cfg.Events, stop: make(chan struct{}), done: make(chan struct{})}
	go l.writeLogs()
	return l, nil
}

// writeLogs writes the server's slog logs, from aslog.Logs, to the logs table every second, or once FlushRecords are
// held, until stopped
func (l *LogsDB) writeLogs() {
	defer close(l.done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-l.stop:
			l.flush(aslog.Logs)
			return
		case <-ticker.C:
		case <-aslog.Logs.Full:
		}
		l.flush(aslog.Logs)
	}
}

// flush writes buffer's held logs in one transaction, and how many were dropped while the buffer was full. A failure is
// written to standard error, not logged, so a broken DB does not fill the buffer with its own failures.
func (l *LogsDB) flush(buffer *aslog.Buffer) {
	records, dropped := buffer.Take()
	if dropped > 0 {
		records = append(records, aslog.Record{Time: time.Now(), Level: "warn", Message: fmt.Sprintf("Dropped %d logs while the buffer was full", dropped)})
	}
	if len(records) == 0 {
		return
	}
	ctx := context.Background()
	tx, err := l.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot write logs to the logs db:", err)
		return
	}
	defer tx.Rollback()
	q := l.Queries.WithTx(tx)
	for _, r := range records {
		err = q.InsertLog(ctx, logsdb.InsertLogParams{SiteID: orNil(r.SiteID), Level: r.Level, Message: r.Message, Attributes: orNil(r.Attributes),
			RequestID: orNil(r.RequestID), JobID: orNil(r.JobID), CreatedAt: r.Time.UnixMilli()})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot write logs to the logs db:", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot write logs to the logs db:", err)
	}
}

// Close writes the last of the held logs, then closes the DB
func (l *LogsDB) Close() error {
	l.stopOnce.Do(func() { close(l.stop) })
	<-l.done
	return l.DB.Close()
}

// Recording is best effort: a log, event or error that cannot be written is only logged with slog, so recording
// never fails what is being recorded. Events and errors are recorded with their own methods, in logs_records.go.

// Log records a message at a level, with any attributes
func (l *LogsDB) Log(ctx context.Context, siteID *string, level LogLevel, message string, attributes map[string]any) {
	var attrs any
	if len(attributes) > 0 {
		attrs = attributes
	}
	requestID, jobID := aslog.IDs(ctx)
	l.record("log", l.Queries.InsertLog(ctx, logsdb.InsertLogParams{SiteID: siteID, Level: string(level), Message: truncate(message), Attributes: attributesJSON(attrs),
		RequestID: orNil(requestID), JobID: orNil(jobID), CreatedAt: time.Now().UnixMilli()}))
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
	requestID, jobID := aslog.IDs(ctx)
	l.record("event", l.Queries.InsertEvent(ctx, logsdb.InsertEventParams{SiteID: orNil(event.SiteID), Batch: orNil(event.Batch), Name: string(event.Name), Attributes: attributesJSON(event.Attributes), StartedAt: event.StartedAt.UnixMilli(), FinishedAt: finished,
		RequestID: orNil(requestID), JobID: orNil(jobID)}))
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
	requestID, jobID := aslog.IDs(ctx)
	l.record("error", l.Queries.InsertError(ctx, logsdb.InsertErrorParams{SiteID: siteID, Message: truncate(err.Error()), Attributes: attributesJSON(attributes),
		RequestID: orNil(requestID), JobID: orNil(jobID)}))
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
