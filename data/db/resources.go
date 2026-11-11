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
		CpuPeggedSeconds: reading.CPUPeggedSeconds, MemoryPercent: reading.MemoryPercent,
		MemoryTotalBytes: reading.MemoryTotalBytes, DiskPercent: reading.DiskPercent, DiskTotalBytes: reading.DiskTotalBytes})
	for _, level := range now {
		high := db.isHigh(level)
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
		if db.isHigh(level) {
			high = append(high, level)
		}
	}
	return high
}

// isHigh is a resource at or over its alert level, but not a disk with DiskAlertFreeGB or more still free
func (db *ControlDB) isHigh(level config.ResourceLevel) bool {
	if level.Resource == "disk" && level.Total != nil {
		free := *level.Total * (100 - level.Value) / 100
		if free >= float64(db.ServerConfig.DiskAlertFreeGB)*(1<<30) {
			return false
		}
	}
	return level.Value >= level.Limit
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
		total    *int64
	}{
		{"memory", reading.MemoryPercent, cfg.MemoryAlertPercent, reading.MemoryTotalBytes},
		{"disk", reading.DiskPercent, cfg.DiskAlertPercent, reading.DiskTotalBytes},
		{"cpu", &pegged, cfg.CPUPeggedAlertSeconds, nil},
	} {
		if l.value != nil {
			level := config.ResourceLevel{Resource: l.resource, Value: *l.value, Limit: float64(l.limit)}
			if l.total != nil {
				level.Total = new(float64(*l.total))
			}
			levels = append(levels, level)
		}
	}
	return levels
}
