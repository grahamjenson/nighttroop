package clear_raw_json

import (
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
)

// Spec is when every site is truncated: each hour at a fixed minute, so a restart does not put it off as it would an
// @every interval
const Spec = "0 17 * * * *"

// SetupCron truncates every site each hour, which does nothing for now, as the raw JSON it cleared is no longer kept
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("TruncateSites", Spec, controlDB.TruncateSites)
}
