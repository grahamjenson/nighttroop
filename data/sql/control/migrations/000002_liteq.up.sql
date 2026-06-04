CREATE TABLE jobs (
    id INTEGER NOT NULL,
    queue TEXT NOT NULL CHECK(length(queue) <= 64),
    job TEXT NOT NULL CHECK(length(job) <= 256),
    job_status TEXT NOT NULL DEFAULT 'queued' CHECK(length(job_status) <= 32),
    execute_after INTEGER NOT NULL DEFAULT 0,
    remaining_attempts INTEGER NOT NULL DEFAULT 1,
    consumer_fetched_at INTEGER NOT NULL DEFAULT 0,
    finished_at INTEGER NOT NULL DEFAULT 0,
    deduping_key TEXT NOT NULL DEFAULT '' CHECK(length(deduping_key) <= 256),
    errors TEXT NOT NULL DEFAULT "[]" CHECK(length(errors) <= 1024),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (id)
);

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

    errors TEXT NOT NULL DEFAULT '[]',

    -- liteq jobs use epoch in seconds
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch() AS INTEGER)),

    -- Virtual columns for metrics
    mean_duration REAL GENERATED ALWAYS AS (
        CASE WHEN total_jobs > 0 
             THEN duration_sum / total_jobs 
             ELSE 0 
        END
    ) VIRTUAL,

    PRIMARY KEY (day_queue_job)
);