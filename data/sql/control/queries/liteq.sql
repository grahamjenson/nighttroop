-- name: QueueJobIgnoreDupe :exec
INSERT INTO
    jobs (
        queue,
        job,
        execute_after,
        job_status,
        created_at,
        updated_at,
        remaining_attempts,
        deduping_key
    )
VALUES
    (
        ?,
        ?,
        ?,
        'queued',
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        ?,
        ?
    ) ON CONFLICT (deduping_key)
WHERE
    deduping_key != ''
    AND job_status IN ('queued', 'fetched') DO NOTHING;

-- name: QueueJobReplaceDupe :exec
INSERT INTO
    jobs (
        queue,
        job,
        execute_after,
        job_status,
        created_at,
        updated_at,
        remaining_attempts,
        deduping_key
    )
VALUES
    (
        ?,
        ?,
        ?,
        'queued',
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        ?,
        ?
    ) ON CONFLICT (deduping_key)
WHERE
    deduping_key != ''
    AND job_status IN ('queued', 'fetched') DO
UPDATE
SET
    job = EXCLUDED.job,
    execute_after = EXCLUDED.execute_after,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    remaining_attempts = EXCLUDED.remaining_attempts;

-- name: CompleteJob :exec
UPDATE
    jobs
SET
    job_status = 'completed',
    finished_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    remaining_attempts = 0
WHERE
    id = ?;

-- name: FailJob :exec
UPDATE
    jobs
SET
    job_status = CASE
        WHEN remaining_attempts <= 1 THEN 'failed'
        ELSE 'queued'
    END,
    finished_at = CASE
        WHEN remaining_attempts <= 1 THEN CAST(unixepoch('subsec') * 1000 AS INTEGER)
        ELSE 0
    END,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    consumer_fetched_at = CASE
        WHEN remaining_attempts <= 1 THEN consumer_fetched_at
        ELSE 0
    END,
    remaining_attempts = MAX(remaining_attempts - 1, 0),
    errors = ?
WHERE
    id = ?;

-- name: MarkJobsForConsumer :many
UPDATE
    jobs
SET
    consumer_fetched_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    job_status = 'fetched'
WHERE
    jobs.job_status = 'queued'
    AND jobs.remaining_attempts > 0
    AND jobs.id IN (
        SELECT
            id
        FROM
            jobs js
        WHERE
            js.queue = ?
            AND js.job_status = 'queued'
            AND js.execute_after <= ?
            AND js.remaining_attempts > 0
        ORDER BY
            execute_after ASC
        LIMIT
            ?
    ) RETURNING *;

-- name: ResetJobs :execrows
UPDATE
    jobs
SET
    job_status = CASE
        WHEN remaining_attempts <= 1 THEN 'failed'
        ELSE 'queued'
    END,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    consumer_fetched_at = 0,
    remaining_attempts = MAX(remaining_attempts - 1, 0),
    errors = json_insert(errors, '$[#]', 'visibility timeout expired')
WHERE
    job_status = 'fetched'
    AND queue = ?
    AND consumer_fetched_at < ?;

-- name: FindJob :one
SELECT
    *
FROM
    jobs
WHERE
    id = ?;


------------
-- Job Stats
------------


-- name: DeleteJobsByIds :exec
DELETE FROM jobs
WHERE id IN (sqlc.slice('ids'));


-- name: GetJobsToBeSummarised :many
SELECT * FROM jobs
WHERE
    job_status IN ('completed', 'failed')
    AND remaining_attempts = 0
    AND created_at < (CAST(unixepoch('subsec') * 1000 AS INTEGER) - 60); --jobs created at more than 1 minute ago


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




