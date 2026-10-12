
------
-- CREATE
------

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


------
-- UPDATE
------


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


-- This grabs jobs that are ready 
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


------
-- READ
------

------------
-- Job Stats
------------

-- name: GetJobsToBeSummarised :many
SELECT * FROM jobs
WHERE
    job_status IN ('completed', 'failed')
    AND remaining_attempts = 0
    AND created_at <= (CAST(unixepoch('subsec') * 1000 AS INTEGER) - CAST(sqlc.arg('SecondsAgo') AS INTEGER)); --jobs created at more than 1 minute ago



------
-- DELETE
------

-- name: DeleteJobsByIds :exec
DELETE FROM jobs
WHERE id IN (sqlc.slice('ids'));



------
-- CRON
------

-- name: InsertSuccessfulCronJob :exec
INSERT INTO
    jobs (
        queue,
        job,
        consumer_fetched_at, -- started at
        job_status,
        finished_at, -- finished at
        created_at,
        updated_at,
        remaining_attempts
    )
VALUES
    (
        "cron",
        ?,
        ?,
        'completed',
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        0
    );

-- name: InsertFailedCronJob :exec
INSERT INTO
    jobs (
        queue,
        job,
        consumer_fetched_at, -- started at
        errors,
        job_status,
        finished_at, -- finished at
        created_at,
        updated_at,
        remaining_attempts
    )
VALUES
    (
        "cron",
        ?,
        ?,
        ?,
        'failed',
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        CAST(unixepoch('subsec') * 1000 AS INTEGER),
        0
    );