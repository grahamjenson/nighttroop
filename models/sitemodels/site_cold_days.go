package sitemodels

import (
	"math"
	"time"

	"nighttroop.t1d.nz/data/retention"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// ColdDays are a page of the site's cold storage: its cold days that match a filter, the newest first or the oldest,
// with how many match and the bytes they take, and the kinds of data the site has cold days of, to filter by
type ColdDays struct {
	Days       []*ColdDay `json:"days"`
	Total      int64      `json:"total"`
	TotalBytes int64      `json:"total_bytes"`
	Kinds      []string   `json:"kinds"`
}

// ColdStorageFilter is which cold days a page of cold storage lists: those of Kind, after DateGT and before DateLT,
// each a UTC day as 2006-01-02, when given, the oldest first if Oldest, Limit of them after Skip
type ColdStorageFilter struct {
	Kind   *string
	DateGT *string
	DateLT *string
	Oldest bool
	Skip   int64
	Limit  int64
}

// ColdDay is one kind of data of one UTC day as the site keeps it once its rows are old: what its packed bytes say
// they are, and what is in them. It is only read: a cold day is written by packaging the day's rows.
type ColdDay struct {
	Kind string `json:"kind"`
	Date string `json:"date"`
	// Bytes is the size of its packed body, and PackagedAt when it was last packaged
	Bytes    int64 `json:"bytes"`
	PackagedAt int64 `json:"packaged_at"`
	// Version and Shape are what its first bytes say it is. SecondOrder is whether each value is kept as the change of
	// its change, and Deflated whether its values are DEFLATE compressed.
	Version     int    `json:"version"`
	Shape       string `json:"shape"`
	SecondOrder bool   `json:"second_order"`
	Deflated    bool   `json:"deflated"`
	// StepSeconds is its grid's step, Scale what each value was multiplied by to be kept as a whole number, Steps how
	// many steps the day has, and Values how many of them have a value
	StepSeconds int64   `json:"step_seconds"`
	Scale       float64 `json:"scale"`
	Steps       int     `json:"steps"`
	Values      int     `json:"values"`
	// Min and Max are its least and greatest values, nil with none
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
	// Error is why its bytes could not be read, as a day of a version or shape this server does not know
	Error *string `json:"error"`
	// Points are its values, each with the time its step starts, for one day of a series opened, and nil in a list
	Points []ColdPoint `json:"points,omitempty"`
	// Events are a day's events, the oldest first, each the columns kept of its row, with "k", its table, for one day
	// of events opened, and nil in a list. For a day of events, Steps and Values are how many there are.
	Events []map[string]any `json:"events,omitempty"`
	// InForce are the events from before a day of events that were still in force when it started: a profile switch,
	// a temp target
	InForce []map[string]any `json:"in_force,omitempty"`
}

// ColdPoint is one step of a series that has a value: the time the step starts, in ms, and the value, as it is kept
type ColdPoint struct {
	At    int64   `json:"at"`
	Value float64 `json:"value"`
}

// ColdDayFromDB is a cold day as it is shown, with its points when asked for
func ColdDayFromDB(row *sitedb.ColdDay, points bool) *ColdDay {
	day := &ColdDay{Kind: row.Kind, Date: row.Date, Bytes: int64(len(row.Body)), PackagedAt: row.UpdatedAt}
	if retention.ShapeOf(row.Body) == retention.EventsShape {
		header, events, err := retention.ReadEvents(row.Body)
		if err != nil {
			day.Error = new(err.Error())
			return day
		}
		day.Version, day.Shape, day.Deflated, day.Steps, day.Values = header.Version, string(header.Shape), header.Deflated, len(events.Events), len(events.Events)
		if points {
			day.Events, day.InForce = events.Events, events.InForce
		}
		return day
	}
	header, values, err := retention.ReadGrid(row.Body)
	if err != nil {
		day.Error = new(err.Error())
		return day
	}
	day.Version, day.Shape, day.SecondOrder, day.Deflated = header.Version, string(header.Shape), header.SecondOrder, header.Deflated
	day.StepSeconds, day.Scale, day.Steps = int64(header.Step/time.Second), header.Scale, header.Steps
	start, _ := time.Parse(time.DateOnly, row.Date)
	for i, value := range values {
		if math.IsNaN(value) {
			continue
		}
		day.Values++
		if day.Min == nil || value < *day.Min {
			day.Min = new(value)
		}
		if day.Max == nil || value > *day.Max {
			day.Max = new(value)
		}
		if points {
			day.Points = append(day.Points, ColdPoint{At: start.Add(time.Duration(i) * header.Step).UnixMilli(), Value: value})
		}
	}
	return day
}
