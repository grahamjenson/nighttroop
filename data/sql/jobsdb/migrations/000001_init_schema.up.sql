-- LITEQ Jobs 
CREATE TABLE jobs (
    id INTEGER NOT NULL,
    queue TEXT NOT NULL CHECK(length(queue) <= 64),
    job TEXT NOT NULL CHECK(length(job) <= 256),
    job_status TEXT NOT NULL DEFAULT 'queued' CHECK(job_status IN ('queued', 'fetched', 'completed', 'failed')),
    execute_after INTEGER NOT NULL DEFAULT 0,
    remaining_attempts INTEGER NOT NULL DEFAULT 1,
    consumer_fetched_at INTEGER NOT NULL DEFAULT 0,
    finished_at INTEGER NOT NULL DEFAULT 0,
    deduping_key TEXT NOT NULL DEFAULT '' CHECK(length(deduping_key) <= 256),
    errors TEXT NOT NULL DEFAULT "[]" CHECK(length(errors) <= 51200 AND json_valid(errors)),
    run_id TEXT CHECK(length(run_id) <= 32), -- a failed run's job_id, which its logs have
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (id)
) STRICT;

CREATE INDEX todo ON jobs (queue, job_status, execute_after)
WHERE
    job_status = 'queued'
    OR job_status = 'fetched';

CREATE UNIQUE INDEX dedupe ON jobs (deduping_key)
WHERE
    deduping_key != ''
    AND job_status IN ('queued', 'fetched');


--- For reporting

CREATE TABLE job_stats_daily (
    day_queue_job TEXT NOT NULL, -- YYYY-MM-DD::queue::job

    total_jobs INTEGER NOT NULL DEFAULT 0,
    total_success INTEGER NOT NULL DEFAULT 0,
    total_failed INTEGER NOT NULL DEFAULT 0,

    duration_sum REAL NOT NULL DEFAULT 0,
    duration_squared_sum REAL NOT NULL DEFAULT 0,

    errors TEXT  NOT  NULL  DEFAULT  "[]" CHECK(length(errors) <= 51200 AND json_valid(errors)),
    failed_run_ids_json TEXT NOT NULL DEFAULT "[]" CHECK(length(failed_run_ids_json) <= 1024 AND json_valid(failed_run_ids_json)), -- the day's last failed runs' job_ids

    -- liteq jobs use epoch in seconds
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),

    PRIMARY KEY (day_queue_job)
) STRICT;

-- The oldest version of these migrations whose build can use the DB at each version, see docs/migrations.md
CREATE TABLE schema_compatibility (
    version INTEGER PRIMARY KEY,
    oldest_version INTEGER NOT NULL
) STRICT;
