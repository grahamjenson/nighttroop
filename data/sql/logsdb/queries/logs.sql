-------
-- CREATE
--------

-- name: InsertLog :exec
INSERT INTO logs (site_id, level, message, attributes) VALUES (?, ?, ?, ?);

-- name: InsertEvent :exec
INSERT INTO events (site_id, name, attributes, started_at, finished_at) VALUES (?, ?, ?, ?, ?);

-- name: InsertError :exec
INSERT INTO errors (site_id, message, attributes) VALUES (?, ?, ?);

-------
-- READ
--------

-- name: SearchLogs :many
-- Newest first, only the site's when SiteID is set
SELECT * FROM logs
WHERE CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID')
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchEvents :many
-- Newest first, only the site's when SiteID is set
SELECT * FROM events
WHERE CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID')
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchErrors :many
-- Newest first, only the site's when SiteID is set
SELECT * FROM errors
WHERE CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID')
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-------
-- DELETE
--------

-- name: TrimLogs :execresult
-- Deletes all but the newest Keep rows
DELETE FROM logs WHERE id <= (SELECT id FROM logs ORDER BY id DESC LIMIT 1 OFFSET sqlc.arg('Keep'));

-- name: TrimEvents :execresult
-- Deletes all but the newest Keep rows
DELETE FROM events WHERE id <= (SELECT id FROM events ORDER BY id DESC LIMIT 1 OFFSET sqlc.arg('Keep'));

-- name: TrimErrors :execresult
-- Deletes all but the newest Keep rows
DELETE FROM errors WHERE id <= (SELECT id FROM errors ORDER BY id DESC LIMIT 1 OFFSET sqlc.arg('Keep'));
