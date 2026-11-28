package retention

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"errors"
	"io"
)

// maxEventsBytes is the most a day's events are when they are read back, so bytes that say they are more are not
const maxEventsBytes = 1 << 20

// errNotEvents is that bytes are not a day's events of a version this code reads
var errNotEvents = errors.New("retention: not a day's events of a known version")

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

// PackEvents packs a day's events as bytes: the version, the shape, how they were packed, then the events as a JSON
// list, DEFLATE compressed when that is smaller. An event is the columns of its row its table's policy keeps, by
// their names, with "k", the table it is from.
func PackEvents(events []map[string]any) ([]byte, error) {
	plain, err := json.Marshal(events)
	if err != nil {
		return nil, err
	}
	var compressed bytes.Buffer
	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	_, _ = writer.Write(plain)
	_ = writer.Close()
	if compressed.Len() < len(plain) {
		return append([]byte{Version, shapeBytes[EventsShape], deflated}, compressed.Bytes()...), nil
	}
	return append([]byte{Version, shapeBytes[EventsShape], 0}, plain...), nil
}

// ReadEvents is the events PackEvents packed, with what their first bytes say they are; Steps is how many there are
func ReadEvents(packed []byte) (Header, []map[string]any, error) {
	if len(packed) < 3 || packed[0] != Version || packed[1] != shapeBytes[EventsShape] || packed[2]&^deflated != 0 {
		return Header{}, nil, errNotEvents
	}
	header, plain := Header{Version: Version, Shape: EventsShape, Deflated: packed[2]&deflated != 0}, packed[3:]
	if header.Deflated {
		var err error
		if plain, err = io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(plain)), maxEventsBytes)); err != nil {
			return Header{}, nil, errNotEvents
		}
	}
	var events []map[string]any
	if err := json.Unmarshal(plain, &events); err != nil {
		return Header{}, nil, errNotEvents
	}
	header.Steps = len(events)
	return header, events, nil
}
