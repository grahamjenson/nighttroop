package connectors

import (
	"context"

	"nighttroop.t1d.nz/data/sql/sitedb"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

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
