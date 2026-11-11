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

-- name: StaleSummaryDays :many
-- The days whose data was written since they were built, by when each starts
SELECT start_at FROM summary_period WHERE period = 'day' AND stale_at IS NOT NULL;

-- name: ClearStaleSummaries :exec
-- The summaries marked stale by Before, as those just built, are up to date; one marked since is built again later
UPDATE summary_period SET stale_at = NULL WHERE stale_at <= @before;

-- name: UnbuiltGlucoseTimes :many
-- The glucose written since Since, before Before, in no day's summary yet, as a backfill writes to days never built,
-- by the 15 minutes each is in
SELECT DISTINCT CAST(glucose_readings.occurred_at / 900000 * 900000 AS INTEGER) AS occurred_at FROM glucose_readings
WHERE glucose_readings.updated_at >= @since AND glucose_readings.occurred_at < @before
    AND glucose_readings.deleted_at IS NULL AND NOT EXISTS (
        SELECT 1 FROM summary_period WHERE summary_period.period = 'day'
            AND summary_period.end_at > glucose_readings.occurred_at AND summary_period.start_at <= glucose_readings.occurred_at
    );
