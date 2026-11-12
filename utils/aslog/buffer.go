package aslog

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// TestOutput is where a test server writes its logs: stdout, unless the tests keep them to show with a failing test
var TestOutput io.Writer = os.Stdout

// Record is a log held to be written to the logs DB, with its site, request and job taken out of its attributes
type Record struct {
	Time       time.Time
	Level      string
	Message    string
	Attributes string
	SiteID     string
	RequestID  string
	JobID      string
}

// The most of a record kept, and of all records held at once: the logs DB's writes keep up with far less, so more is
// only a write that is slow or failing, which must not grow the server's memory
const (
	maxMessageBytes    = 512
	maxAttributesBytes = 2048
	maxBufferedBytes   = 1 << 20
	// FlushRecords is how many records are written to the logs DB at once
	FlushRecords = 1000
)

// Buffer holds the records logged since they were last taken. Past maxBufferedBytes new records are only counted.
type Buffer struct {
	mu      sync.Mutex
	records []Record
	bytes   int
	dropped int
	// Full is signalled when FlushRecords records are held, so they are written without waiting for the next tick
	Full chan struct{}
}

// Logs is the server's buffer, which every slog log is copied into for the logs DB to write
var Logs = &Buffer{Full: make(chan struct{}, 1)}

func (b *Buffer) add(record Record) {
	b.mu.Lock()
	defer b.mu.Unlock()
	size := record.size()
	if b.bytes+size > maxBufferedBytes {
		b.dropped++
		return
	}
	b.records, b.bytes = append(b.records, record), b.bytes+size
	if len(b.records) == FlushRecords {
		select {
		case b.Full <- struct{}{}:
		default:
		}
	}
}

// Take is up to FlushRecords of the oldest records held, and how many were dropped since the last take
func (b *Buffer) Take() ([]Record, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := min(len(b.records), FlushRecords)
	taken := append([]Record(nil), b.records[:n]...)
	b.records = append(b.records[:0], b.records[n:]...)
	for _, record := range taken {
		b.bytes -= record.size()
	}
	dropped := b.dropped
	b.dropped = 0
	return taken, dropped
}

// size is about what a record holds in memory, its text
func (r Record) size() int {
	return len(r.Message) + len(r.Attributes) + len(r.SiteID) + len(r.RequestID) + len(r.JobID)
}

// BufferHandler copies each log into a Buffer, beside the handlers writing it to standard out. Nothing here logs
// under a group, so a group's attributes are kept beside the others.
type BufferHandler struct {
	Buffer *Buffer
	attrs  []slog.Attr
}

func (h BufferHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h BufferHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return h
}

func (h BufferHandler) WithGroup(string) slog.Handler { return h }

func (h BufferHandler) Handle(_ context.Context, r slog.Record) error {
	record := Record{Time: r.Time, Level: level(r.Level), Message: truncate(r.Message, maxMessageBytes)}
	attributes := map[string]any{}
	add := func(attr slog.Attr) bool {
		if !record.setID(attr) && attr.Key != "ip" {
			attributes[attr.Key] = jsonValue(attr.Value)
		}
		return true
	}
	for _, attr := range h.attrs {
		add(attr)
	}
	r.Attrs(add)
	if len(attributes) > 0 {
		raw, err := json.Marshal(attributes)
		if err != nil || len(raw) > maxAttributesBytes {
			raw = []byte(`{"truncated":true}`)
		}
		record.Attributes = string(raw)
	}
	h.Buffer.add(record)
	return nil
}

// setID sets the record's site, request or job from an attribute of that key, saying whether it was one
func (r *Record) setID(attr slog.Attr) bool {
	switch attr.Key {
	case "site_id":
		r.SiteID = truncate(attr.Value.String(), 32)
	case "request_id":
		r.RequestID = truncate(attr.Value.String(), 64)
	case "job_id":
		r.JobID = truncate(attr.Value.String(), 32)
	default:
		return false
	}
	return true
}

// jsonValue is an attribute's value as JSON takes it, a group as an object, leaving out IP addresses, which the logs
// DB does not keep
func jsonValue(v slog.Value) any {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindGroup:
		group := map[string]any{}
		for _, attr := range v.Group() {
			if attr.Key != "ip" {
				group[attr.Key] = jsonValue(attr.Value)
			}
		}
		return group
	case slog.KindAny:
		if err, ok := v.Any().(error); ok {
			return err.Error()
		}
	}
	return v.Any()
}

// level is the logs table's level for a slog level, rounding one between them down
func level(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	case l >= slog.LevelInfo:
		return "info"
	}
	return "debug"
}

// truncate keeps s within n bytes, cutting it at a whole character
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n-3], "") + "..."
}

// IDs are the request and background job run ctx logs, "" for none
func IDs(ctx context.Context) (requestID, jobID string) {
	var record Record
	for _, attr := range attrsOf(ctx) {
		record.setID(attr)
	}
	return record.RequestID, record.JobID
}
