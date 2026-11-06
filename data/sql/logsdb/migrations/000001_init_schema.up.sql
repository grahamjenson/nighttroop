-- Logs, events and errors from around the server, each optionally about a site, and the request or background job run
-- it was for. This DB is not backed up, and a cron job trims each table to its newest LOGS_DB_MAX_ROWS_PER_TABLE rows.
-- Every slog log is kept in logs, written in batches. A log's attributes can be up to 256 KB, so a staging server
-- capturing requests (CAPTURE_REQUESTS) keeps each one's headers and bodies in one log; every other log keeps its
-- attributes within 4 KB, as the code limits them.

CREATE TABLE logs (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    level TEXT NOT NULL CHECK(level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL CHECK(length(message) <= 1024),
    attributes TEXT CHECK(length(attributes) <= 262144 AND json_valid(attributes)),
    request_id TEXT CHECK(length(request_id) <= 64),
    job_id TEXT CHECK(length(job_id) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX logs_site ON logs (site_id, id);
CREATE INDEX logs_request ON logs (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX logs_job ON logs (job_id) WHERE job_id IS NOT NULL;

-- An event started at started_at, and took until finished_at when it has one, so how long it took is known. An
-- event for a batch of sites, rather than one site, has the batch, so each site in it can see the event.
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    batch TEXT CHECK(length(batch) <= 32),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    started_at INTEGER NOT NULL,
    finished_at INTEGER CHECK(finished_at >= started_at),
    request_id TEXT CHECK(length(request_id) <= 64),
    job_id TEXT CHECK(length(job_id) <= 32)
) STRICT;

CREATE INDEX events_site ON events (site_id, id);
CREATE INDEX events_batch ON events (batch, id);
CREATE INDEX events_request ON events (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX events_job ON events (job_id) WHERE job_id IS NOT NULL;

CREATE TABLE errors (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    message TEXT NOT NULL CHECK(length(message) <= 1024),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    request_id TEXT CHECK(length(request_id) <= 64),
    job_id TEXT CHECK(length(job_id) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX errors_site ON errors (site_id, id);
CREATE INDEX errors_request ON errors (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX errors_job ON errors (job_id) WHERE job_id IS NOT NULL;

-- The oldest version of these migrations whose build can use the DB at each version, see docs/migrations.md
CREATE TABLE schema_compatibility (
    version INTEGER PRIMARY KEY,
    oldest_version INTEGER NOT NULL
) STRICT;
