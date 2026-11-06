-- Back to attributes within 4 KB: a captured request's longer attributes are replaced with a note that they were cut
-- changes rows of logs
CREATE TEMP TABLE logs_copy AS SELECT * FROM logs;
DROP TABLE logs;
CREATE TABLE logs (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    level TEXT NOT NULL CHECK(level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL CHECK(length(message) <= 1024),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    request_id TEXT CHECK(length(request_id) <= 64),
    job_id TEXT CHECK(length(job_id) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;
INSERT INTO logs (id, site_id, level, message, attributes, request_id, job_id, created_at)
SELECT id, site_id, level, message, CASE WHEN length(attributes) > 4096 THEN '{"truncated":true}' ELSE attributes END, request_id, job_id, created_at FROM logs_copy;
DROP TABLE logs_copy;
CREATE INDEX logs_site ON logs (site_id, id);
CREATE INDEX logs_request ON logs (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX logs_job ON logs (job_id) WHERE job_id IS NOT NULL;
