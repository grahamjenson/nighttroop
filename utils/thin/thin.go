// Package thin is how a site's device reports are thinned as they age: of each device's reports, all from the last
// KeepAll are kept, then only its newest in each bucket, a bucket larger the older the report. The site's truncation
// (ThinOldDeviceReports) deletes the others, after their numbers are packaged. Bucket is the rule, for its tests.
package thin

import "time"

const (
	// KeepAll is how long every report is kept
	KeepAll = time.Hour
	// Week and Fortnight are the ages after which a report's bucket is Medium, then Large; before them it is Small
	Week, Fortnight = 7 * 24 * time.Hour, 14 * 24 * time.Hour
	// Small, Medium and Large are the buckets in which a device keeps its newest report. They are numbered from the
	// epoch, so each Large's newest is also its Medium's and Small's.
	Small, Medium, Large = 10 * time.Minute, 20 * time.Minute, time.Hour
)

// Bucket is the size, in ms, of the bucket in which a report that occurred at at is kept only as its device's newest,
// at now, both in ms, or 0 while every report is kept
func Bucket(at, now int64) int64 {
	switch {
	case at >= now-KeepAll.Milliseconds():
		return 0
	case at < now-Fortnight.Milliseconds():
		return Large.Milliseconds()
	case at < now-Week.Milliseconds():
		return Medium.Milliseconds()
	}
	return Small.Milliseconds()
}
