package dexcommodels

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
)

// EGV represents a Dexcom glucose reading.
type EGV struct {
	Value       int
	SystemTime  time.Time
	DisplayTime time.Time
	RawJson     string
}

func (e EGV) UTCOffsetInMinutes() int {
	_, offset := e.DisplayTime.Zone()
	return offset / 60
}

func EGVToCreate(e EGV, createdByIdentityName string) sitedb.CreateGlucoseReadingParams {
	return sitedb.CreateGlucoseReadingParams{
		ID:                    utils.NextID(),
		OccurredAt:            e.SystemTime.UnixMilli(),
		Mgdl:                  float64(e.Value),
		RawJsonData:           &e.RawJson,
		CreatedByIdentityName: createdByIdentityName,
		UtcOffsetMinutes:      ptr(int64(e.UTCOffsetInMinutes())),
	}
}

func (e *EGV) UnmarshalJSON(data []byte) error {
	var r dexcomEGVResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	systemTime, err := parseDexcomTime(r.ST)
	if err != nil {
		return err
	}
	displayTime, err := parseDexcomTime(r.DT)
	if err != nil {
		return err
	}

	e.Value = r.Value
	e.SystemTime = systemTime
	e.DisplayTime = displayTime
	e.RawJson = string(data)

	return nil
}

type dexcomEGVResponse struct {
	WT    string `json:"WT"`
	ST    string `json:"ST"`
	DT    string `json:"DT"`
	Value int    `json:"Value"`
}

// parseDexcomTime parses Dexcom's weird /Date(1234567890)/ format.
var dateRe = regexp.MustCompile(`^Date\((\d+)(?:([+-]\d{4}))?\)$`)

// ParseDate parses:
//
//	Date(1780533499468)
//	Date(1780533499468+1200)
//	Date(1780533499468-0500)
func parseDexcomTime(s string) (time.Time, error) {
	matches := dateRe.FindStringSubmatch(s)
	if matches == nil {
		return time.Time{}, fmt.Errorf("invalid date format: %q", s)
	}

	ms, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	t := time.UnixMilli(ms)

	// No timezone offset supplied.
	if matches[2] == "" {
		return t.UTC(), nil
	}

	offsetStr := matches[2]

	sign := 1
	if offsetStr[0] == '-' {
		sign = -1
	}

	hours, err := strconv.Atoi(offsetStr[1:3])
	if err != nil {
		return time.Time{}, err
	}

	minutes, err := strconv.Atoi(offsetStr[3:5])
	if err != nil {
		return time.Time{}, err
	}

	offsetSeconds := sign * ((hours * 3600) + (minutes * 60))

	loc := time.FixedZone(offsetStr, offsetSeconds)

	// Interpret the epoch instant in the supplied timezone.
	return t.In(loc), nil
}

func ptr[T any](v T) *T {
	return &v
}
