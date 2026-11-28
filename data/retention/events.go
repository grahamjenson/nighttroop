package retention

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"errors"
	"io"
)

// maxJSONBytes is the most a body of JSON is when it is read back, so bytes that say they are more are not
const maxJSONBytes = 1 << 20

// EventsVersion is the version of a day's events this code writes. Version 1 was the list of events alone; version 2
// is an object, with what was in force when the day started. Both are read.
const EventsVersion = 2

// errNotJSON is that bytes are not a day's events of a version this code reads
var errNotJSON = errors.New("retention: not a day's events of a known version")

// ShapeOf is the shape a cold day's bytes say they are, "" when they say none this code knows
func ShapeOf(packed []byte) Shape {
	if len(packed) < 2 {
		return ""
	}
	for shape, b := range shapeBytes {
		if b == packed[1] {
			return shape
		}
	}
	return ""
}

// EventsDay is a day's events. An event is the columns of its row its table's policy keeps, by their names, with "k",
// the table it is from.
type EventsDay struct {
	// InForce are the events from before the day that were still in force when it started, the oldest first: the
	// profile switch with no end, one with an end not yet reached, and a temp target not yet over. With them a day
	// is read alone, as the rows of the days before it may be gone.
	InForce []map[string]any `json:"in_force,omitempty"`
	// Events are the day's own, the oldest first
	Events []map[string]any `json:"events"`
}

// packJSON packs a value as bytes: the version, the shape, how it was packed, then its JSON, DEFLATE compressed when
// that is smaller
func packJSON(version byte, shape Shape, value any) ([]byte, error) {
	plain, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var compressed bytes.Buffer
	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	_, _ = writer.Write(plain)
	_ = writer.Close()
	if compressed.Len() < len(plain) {
		return append([]byte{version, shapeBytes[shape], deflated}, compressed.Bytes()...), nil
	}
	return append([]byte{version, shapeBytes[shape], 0}, plain...), nil
}

// readJSON is the JSON packJSON packed for the shape, with what its first bytes say it is
func readJSON(shape Shape, packed []byte) (Header, []byte, error) {
	if len(packed) < 3 || packed[1] != shapeBytes[shape] || packed[2]&^deflated != 0 {
		return Header{}, nil, errNotJSON
	}
	header, plain := Header{Version: int(packed[0]), Shape: shape, Deflated: packed[2]&deflated != 0}, packed[3:]
	if header.Deflated {
		var err error
		if plain, err = io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(plain)), maxJSONBytes)); err != nil {
			return Header{}, nil, errNotJSON
		}
	}
	return header, plain, nil
}

// PackEvents packs a day's events as bytes, as JSON
func PackEvents(day EventsDay) ([]byte, error) {
	if day.Events == nil {
		day.Events = []map[string]any{}
	}
	return packJSON(EventsVersion, EventsShape, day)
}

// ReadEvents is the day PackEvents packed, of either version, with what its first bytes say it is; Steps is how many
// events the day has of its own
func ReadEvents(packed []byte) (Header, EventsDay, error) {
	header, plain, err := readJSON(EventsShape, packed)
	var day EventsDay
	switch {
	case err != nil:
		return Header{}, day, err
	case header.Version == 1:
		err = json.Unmarshal(plain, &day.Events)
	case header.Version == EventsVersion:
		err = json.Unmarshal(plain, &day)
	default:
		err = errNotJSON
	}
	if err != nil {
		return Header{}, EventsDay{}, errNotJSON
	}
	header.Steps = len(day.Events)
	return header, day, nil
}
