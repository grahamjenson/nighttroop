package truncate_sites

import (
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
)

// Spec is when every site is truncated: every 10 minutes, at fixed minutes, so a restart does not put it off as it
// would an @every interval
const Spec = "0 */10 * * * *"

// SetupCron truncates every site every 10 minutes, keeping each device's loop status whole only on its newest report,
// and one of its reports every 10 minutes from before the last hour
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("TruncateSites", Spec, controlDB.TruncateSites)
}
