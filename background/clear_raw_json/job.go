package clear_raw_json

import (
	"nighttroop.t1d.nz/background/cron"
	"nighttroop.t1d.nz/data/db"
)

// Spec is when every site's old raw JSON is cleared: each hour at a fixed minute, so a restart does not put it off as
// it would an @every interval
const Spec = "0 17 * * * *"

// SetupCron clears the raw JSON every site keeps of rows older than TRUNCATE_RAW_JSON_AFTER_DAYS, each hour
func SetupCron(c cron.CronScheduler, controlDB *db.ControlDB) {
	c.AddFunc("ClearRawJSON", Spec, controlDB.ClearRawJSON)
}
