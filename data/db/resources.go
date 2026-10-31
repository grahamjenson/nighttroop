package db

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/utils/config"
)

// KeepResourceReadings is how long checks of the server's resources are kept, for the control site's graphs
const KeepResourceReadings = 3 * 24 * time.Hour

// RecordResources keeps a check of the server's resources, forgetting checks older than a few days, and records an
// event for each resource that went over its alert level since the last check, or came back under it
func (db *ControlDB) RecordResources(ctx context.Context, reading livedb.CreateResourceReadingParams) error {
	previous, err := db.Live.Queries.LatestResourceReading(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err := db.Live.Queries.CreateResourceReading(ctx, reading); err != nil {
		return err
	}
	before := time.UnixMilli(reading.At).Add(-KeepResourceReadings).UnixMilli()
	if err := db.Live.Queries.DeleteResourceReadingsBefore(ctx, livedb.DeleteResourceReadingsBeforeParams{Before: before}); err != nil {
		return err
	}

	wasHigh := db.highResources(previous)
	now := db.resourceLevels(&livedb.ResourceReading{At: reading.At, CpuPercent: reading.CPUPercent,
		CpuPeggedSeconds: reading.CPUPeggedSeconds, MemoryPercent: reading.MemoryPercent, DiskPercent: reading.DiskPercent})
	for _, level := range now {
		high := level.Value >= level.Limit
		was := slices.ContainsFunc(wasHigh, func(l config.ResourceLevel) bool { return l.Resource == level.Resource })
		switch {
		case high && !was:
			db.Logs.ResourceHighEvent(ctx, level)
		case !high && was:
			db.Logs.ResourceRecoveredEvent(ctx, level)
		}
	}
	return nil
}

// HighResources are the resources at or over their alert levels at the last check, for the control site's banner
func (db *ControlDB) HighResources(ctx context.Context) []config.ResourceLevel {
	latest, err := db.Live.Queries.LatestResourceReading(ctx)
	if err != nil {
		return nil
	}
	return db.highResources(latest)
}

func (db *ControlDB) highResources(reading *livedb.ResourceReading) []config.ResourceLevel {
	high := []config.ResourceLevel{}
	for _, level := range db.resourceLevels(reading) {
		if level.Value >= level.Limit {
			high = append(high, level)
		}
	}
	return high
}

// resourceLevels are a check's resources that could be read, with their alert levels
func (db *ControlDB) resourceLevels(reading *livedb.ResourceReading) []config.ResourceLevel {
	if reading == nil {
		return nil
	}
	cfg := db.ServerConfig
	pegged := float64(reading.CpuPeggedSeconds)
	levels := []config.ResourceLevel{}
	for _, l := range []struct {
		resource string
		value    *float64
		limit    int
	}{
		{"memory", reading.MemoryPercent, cfg.MemoryAlertPercent},
		{"disk", reading.DiskPercent, cfg.DiskAlertPercent},
		{"cpu", &pegged, cfg.CPUPeggedAlertSeconds},
	} {
		if l.value != nil {
			levels = append(levels, config.ResourceLevel{Resource: l.resource, Value: *l.value, Limit: float64(l.limit)})
		}
	}
	return levels
}
