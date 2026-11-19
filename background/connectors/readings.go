package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/sitedb"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

// copyReadings fetches a connector's readings and keeps them in the site, returning how many were new. name is the
// connector's name in its errors, and the fetch tells unreadable of each reading it cannot read. A reading that cannot
// be read or kept is skipped, and only the first of a sync is recorded.
func copyReadings[T any](ctx context.Context, jobInfo *ConnectorJobInfo, name string, unreadable *func(json.RawMessage, error), fetch func(context.Context) ([]T, error), toCreate func(T, string) sitedb.CreateGlucoseReadingParams) (int, error) {
	refused := false
	refuse := func(err error, record any) {
		if !refused {
			refused = true
			recordRefused(ctx, jobInfo, err, record)
		}
	}
	*unreadable = func(raw json.RawMessage, err error) {
		refuse(fmt.Errorf("%s: cannot read a reading: %w", name, err), raw)
	}
	readings, err := fetch(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s Error: %w", name, err)
	}
	createdBy := string(jobInfo.Type) + "_connector"
	added, err := keepReadings(ctx, jobInfo.Site, readings, func(r T) sitedb.CreateGlucoseReadingParams {
		return toCreate(r, createdBy)
	}, func(err error, r T) { refuse(fmt.Errorf("%s: cannot keep a reading: %w", name, err), r) })
	if err != nil {
		return 0, fmt.Errorf("%s Error: %w", name, err)
	}
	if added > 0 {
		slog.InfoContext(ctx, "Added new readings from a connector", "connector", name, "added", added)
	}
	return added, nil
}

// keepReadings stores a sync's readings in one transaction, so a sync takes the site's write lock once, not once for
// each reading, though the site has most of them already: Libre sends its last 12 hours each time, and a reading's ID
// is its time. A reading the site has is left out, and refuse is told of any other it cannot keep. Each kept is
// published once committed, so no one sees a reading that is not stored. It is how many were kept.
func keepReadings[T any](ctx context.Context, site *ConnectorSite, readings []T, toCreate func(T) sitedb.CreateGlucoseReadingParams, refuse func(error, T)) (int, error) {
	if len(readings) == 0 {
		return 0, nil
	}
	tx, err := site.SiteDB.SQLiteDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	q := site.SiteDB.QueriesWithTx(tx)
	kept := []*sitedb.GlucoseReading{}
	for _, reading := range readings {
		// A unique constraint fails only its own statement, so the transaction goes on
		row, err := q.CreateGlucoseReading(ctx, toCreate(reading))
		if err == nil {
			kept = append(kept, row)
		} else if !utilerrors.IsSQLDuplicateID(err) {
			refuse(err, reading)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	for _, row := range kept {
		site.published(row)
	}
	return len(kept), nil
}
