// Package thin is how a site's device reports are thinned as they age: of each device's reports, all from the last
// KeepAll are kept, then only its newest in each Size. The site's truncation (ThinOldDeviceReports) deletes the others,
// after their numbers are packaged, and all of them once they are older than the site keeps device reports.
package thin

import "time"

const (
	// KeepAll is how long every report is kept
	KeepAll = time.Hour
	// Size is the bucket in which a device keeps its newest report once it is older than KeepAll. The buckets are
	// numbered from the epoch.
	Size = 10 * time.Minute
)

// Bucket is the size, in ms, of the bucket in which a report that occurred at at is kept only as its device's newest,
// at now, both in ms, or 0 while every report is kept
func Bucket(at, now int64) int64 {
	if at >= now-KeepAll.Milliseconds() {
		return 0
	}
	return Size.Milliseconds()
}
