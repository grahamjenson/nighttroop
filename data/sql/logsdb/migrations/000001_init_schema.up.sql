-- Logs, events and errors from around the server, each optionally about a site. This DB is not backed up, and a
-- cron job trims each table to its newest LOGS_DB_MAX_ROWS_PER_TABLE rows.

CREATE TABLE logs (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    level TEXT NOT NULL CHECK(level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL CHECK(length(message) <= 1024),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX logs_site ON logs (site_id, id);

CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX events_site ON events (site_id, id);

CREATE TABLE errors (
    id INTEGER PRIMARY KEY,
    site_id TEXT CHECK(length(site_id) <= 32),
    message TEXT NOT NULL CHECK(length(message) <= 1024),
    attributes TEXT CHECK(length(attributes) <= 4096 AND json_valid(attributes)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
) STRICT;

CREATE INDEX errors_site ON errors (site_id, id);
