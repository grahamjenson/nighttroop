-- name: UpsertSummaryPeriod :one
INSERT INTO summary_period (period, date, timezone, start_at, end_at, days, time_in_range, total_daily_dose, total_daily_carbs, glucose_mgdl_json, agp_mgdl_json,
    loop_iob_json, loop_iob_basal_json, loop_iob_treatment_json, loop_cob_json, glucose_readings, meter_readings, cgm_slots,
    glucose_mean_mgdl, glucose_sd_mgdl, time_very_low, time_low, time_high, time_very_high, total_daily_basal)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(period, date) DO UPDATE SET
    timezone = excluded.timezone,
    start_at = excluded.start_at,
    end_at = excluded.end_at,
    days = excluded.days,
    time_in_range = excluded.time_in_range,
    total_daily_dose = excluded.total_daily_dose,
    total_daily_carbs = excluded.total_daily_carbs,
    glucose_mgdl_json = excluded.glucose_mgdl_json,
    agp_mgdl_json = excluded.agp_mgdl_json,
    loop_iob_json = excluded.loop_iob_json,
    loop_iob_basal_json = excluded.loop_iob_basal_json,
    loop_iob_treatment_json = excluded.loop_iob_treatment_json,
    loop_cob_json = excluded.loop_cob_json,
    glucose_readings = excluded.glucose_readings,
    meter_readings = excluded.meter_readings,
    cgm_slots = excluded.cgm_slots,
    glucose_mean_mgdl = excluded.glucose_mean_mgdl,
    glucose_sd_mgdl = excluded.glucose_sd_mgdl,
    time_very_low = excluded.time_very_low,
    time_low = excluded.time_low,
    time_high = excluded.time_high,
    time_very_high = excluded.time_very_high,
    total_daily_basal = excluded.total_daily_basal,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: GetSummaryPeriod :one
SELECT * FROM summary_period WHERE period = ? AND date = ?;

-- name: SearchSummaryPeriods :many
SELECT * FROM summary_period
WHERE
    period = sqlc.arg('Period')
    AND (CAST(sqlc.narg('DateGTE') AS TEXT) IS NULL OR date >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS TEXT) IS NULL OR date <= sqlc.narg('DateLTE'))
ORDER BY date DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: HardDeleteSummaryPeriod :execresult
DELETE FROM summary_period WHERE period = ? AND date = ?;

-- name: DeleteSummaryPeriods :execresult
-- Deletes the days, weeks or months, or with no period all of them
DELETE FROM summary_period WHERE CAST(sqlc.narg('Period') AS TEXT) IS NULL OR period = sqlc.narg('Period');

-- name: MarkSummariesStale :exec
-- The summaries a record at At was deleted or moved from; none is longer than 32 days, however old At is
UPDATE summary_period SET stale_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE stale_at IS NULL AND end_at > @at AND end_at <= @at + 2764800000 AND start_at <= @at;

-- name: StaleSummaryDays :many
-- The days marked to be built again, by when each starts
SELECT start_at FROM summary_period WHERE period = 'day' AND stale_at IS NOT NULL;

-- name: ClearStaleSummaries :exec
-- The summaries marked stale by Before, as those just built, are up to date; one marked since is built again later
UPDATE summary_period SET stale_at = NULL WHERE stale_at <= @before;

-- name: SummaryDataWrittenSince :many
-- When the data a day's summary is built from that was written, changed or deleted after Since, and happened before
-- Before, happened, by the 15 minutes each is in, as every timezone's midnight starts one
SELECT DISTINCT CAST(written.occurred_at / 900000 * 900000 AS INTEGER) AS occurred_at FROM (
    SELECT glucose_readings.occurred_at FROM glucose_readings
        WHERE glucose_readings.updated_at > @since AND glucose_readings.occurred_at < @before
    UNION ALL SELECT insulin_doses.occurred_at FROM insulin_doses
        WHERE insulin_doses.updated_at > @since AND insulin_doses.occurred_at < @before
    UNION ALL SELECT food_intakes.occurred_at FROM food_intakes
        WHERE food_intakes.updated_at > @since AND food_intakes.occurred_at < @before
    UNION ALL SELECT temp_basals.occurred_at FROM temp_basals
        WHERE temp_basals.updated_at > @since AND temp_basals.occurred_at < @before
    UNION ALL SELECT temp_targets.occurred_at FROM temp_targets
        WHERE temp_targets.updated_at > @since AND temp_targets.occurred_at < @before
    UNION ALL SELECT profile_switches.occurred_at FROM profile_switches
        WHERE profile_switches.updated_at > @since AND profile_switches.occurred_at < @before
    UNION ALL SELECT event_groups.occurred_at FROM event_groups
        WHERE event_groups.updated_at > @since AND event_groups.occurred_at < @before
    UNION ALL SELECT device_reports.occurred_at FROM device_reports
        WHERE device_reports.updated_at > @since AND device_reports.occurred_at < @before
) AS written;
