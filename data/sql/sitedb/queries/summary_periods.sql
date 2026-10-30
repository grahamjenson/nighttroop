-- name: UpsertSummaryPeriod :one
INSERT INTO summary_periods (period, date, timezone, start_at, end_at, days, time_in_range, mean_daily_dose, mean_daily_carbs)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(period, date) DO UPDATE SET
    timezone = excluded.timezone,
    start_at = excluded.start_at,
    end_at = excluded.end_at,
    days = excluded.days,
    time_in_range = excluded.time_in_range,
    mean_daily_dose = excluded.mean_daily_dose,
    mean_daily_carbs = excluded.mean_daily_carbs,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: GetSummaryPeriod :one
SELECT * FROM summary_periods WHERE period = ? AND date = ?;

-- name: DeleteAllSummaryPeriods :execresult
DELETE FROM summary_periods;
