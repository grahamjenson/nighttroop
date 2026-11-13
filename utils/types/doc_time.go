package types

import (
	"encoding/json"
	"strconv"
	"time"
)

// TimeWrite is which API versions write a time alias
type TimeWrite int

const (
	// WriteAlways writes it on API v1 and v3
	WriteAlways TimeWrite = iota
	// WriteV1 writes it on API v1 only, as one Nighttroop makes up for a client of v1, which a v3 client may refuse
	WriteV1
	// WriteNever only reads it
	WriteNever
)

// TimeAlias is one of the fields a document says when it happened in, e.g. date, created_at or dateString
type TimeAlias struct {
	Name string
	// Millis is in milliseconds, as date and mills are; the others are written as toISOString writes them, in UTC with
	// three fractional digits, and are read from a string in any layout StringToTime reads, or from milliseconds
	Millis bool
	Write  TimeWrite
}

// TimeAliases are the fields a collection's documents say when they happened in, in the order they are read. Every
// collection reads and writes its time with them, so they keep the same rules:
//
//   - The time is the first alias sent. One sent after it as a number of milliseconds in the same second gives it its
//     milliseconds, as xDrip+ sends a created_at in whole seconds beside a timestamp in ms.
//   - The UTC offset is the utcOffset sent, else the zone of the first alias sent as a string. A string in UTC, with
//     Z, says when, not where, as Loop writes every time, so it gives none, and the document takes the site's.
//   - An alias sent that is not a time refuses the document, unless it is one Nighttroop makes up (WriteV1), as a
//     treatment's timestamp is, which is read only when it is a time.
//   - Each alias is written from when it happened, and the created_at the server writes for every document is too.
type TimeAliases []TimeAlias

// sentTime is an alias's time, and whether it was sent as a number or a string
type sentTime struct {
	at       time.Time
	isNumber bool
}

// read is the alias's time in raw, nil when it is not sent
func (alias TimeAlias) read(raw json.RawMessage) (*sentTime, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	isNumber := raw[0] != '"'
	if alias.Millis {
		var ms DateInt
		if err := json.Unmarshal(raw, &ms); err != nil {
			return nil, err
		}
		return &sentTime{at: ms.Time(), isNumber: true}, nil
	}
	var sent DateString
	if err := json.Unmarshal(raw, &sent); err != nil {
		return nil, err
	}
	return &sentTime{at: sent.Time(), isNumber: isNumber}, nil
}

// Read is when the document in data happened, in milliseconds, nil when it says no time, and its UTC offset in
// minutes. A model reads them before its other fields, so a time that is not one is the error it is refused with.
func (aliases TimeAliases) Read(data []byte) (*DateInt, *FlexibleInt64, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, nil, err
	}
	var date *DateInt
	var offset *FlexibleInt64
	if sent := fields["utcOffset"]; len(sent) > 0 && string(sent) != "null" {
		offset = new(FlexibleInt64)
		if err := json.Unmarshal(sent, offset); err != nil {
			return nil, nil, err
		}
	}
	zoned := false
	for _, alias := range aliases {
		sent, err := alias.read(fields[alias.Name])
		if err != nil && alias.Write != WriteV1 {
			return nil, nil, err
		}
		if sent == nil {
			continue
		}
		ms := DateInt(sent.at.UnixMilli())
		if date == nil {
			date = &ms
		} else if sent.isNumber && int64(ms)/1000 == int64(*date)/1000 {
			date = &ms
		}
		if !sent.isNumber && !zoned {
			zoned = true
			if _, seconds := sent.at.Zone(); offset == nil && seconds != 0 {
				offset = new(FlexibleInt64(seconds / 60))
			}
		}
	}
	return date, offset, nil
}

// Written are the aliases written on the API version, with when it happened as each writes it
func (aliases TimeAliases) Written(ms int64, v3 bool) [][2]string {
	written := make([][2]string, 0, len(aliases))
	for _, alias := range aliases {
		if alias.Write == WriteNever || (v3 && alias.Write == WriteV1) {
			continue
		}
		value := strconv.Quote(ISOString(time.UnixMilli(ms)))
		if alias.Millis {
			value = strconv.FormatInt(ms, 10)
		}
		written = append(written, [2]string{alias.Name, value})
	}
	return written
}
