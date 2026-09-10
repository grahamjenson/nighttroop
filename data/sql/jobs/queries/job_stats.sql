-- name: EnsureJobStatsDailyForGroup :one
INSERT INTO job_stats_daily (day_queue_job, errors, updated_at)
VALUES (?, '[]', CAST(unixepoch('subsec') * 1000 AS INTEGER))
ON CONFLICT(day_queue_job) DO UPDATE SET
    day_queue_job = day_queue_job -- Forces NOOP assignment to return conflicting row
RETURNING *;

-- name: UpdateJobStatsDaily :exec
UPDATE job_stats_daily
SET
    total_jobs = total_jobs + sqlc.arg(total_jobs),
    total_success = total_success + sqlc.arg(total_success),
    total_failed = total_failed + sqlc.arg(total_failed),
    duration_sum = duration_sum + sqlc.arg(duration_sum),
    duration_squared_sum = duration_squared_sum + sqlc.arg(duration_squared_sum),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),

    errors = (
        SELECT json_group_array(DISTINCT value)
        FROM (
            SELECT value FROM json_each(COALESCE(errors, '[]'))
            UNION ALL
            SELECT value FROM json_each(sqlc.arg(errors))
        )
    )
WHERE
    day_queue_job = sqlc.arg(day_queue_job)
;

-- name: AllJobStatsDaily :many
SELECT * FROM job_stats_daily;

