-- name: UpsertSummaryDaily :one
INSERT INTO summary_daily (date, timezone, start_at, end_at, time_in_range, total_daily_dose, total_carbs, glucose_mgdl_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(date) DO UPDATE SET
    timezone = excluded.timezone,
    start_at = excluded.start_at,
    end_at = excluded.end_at,
    time_in_range = excluded.time_in_range,
    total_daily_dose = excluded.total_daily_dose,
    total_carbs = excluded.total_carbs,
    glucose_mgdl_json = excluded.glucose_mgdl_json,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: GetSummaryDaily :one
SELECT * FROM summary_daily WHERE date = ?;

-- name: SearchSummaryDaily :many
SELECT * FROM summary_daily
WHERE
    (CAST(sqlc.narg('DateGTE') AS TEXT) IS NULL OR date >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS TEXT) IS NULL OR date <= sqlc.narg('DateLTE'))
ORDER BY date DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: HardDeleteSummaryDaily :execresult
DELETE FROM summary_daily WHERE date = ?;

-- name: DeleteAllSummaryDaily :execresult
DELETE FROM summary_daily;
