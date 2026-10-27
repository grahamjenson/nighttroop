-------
-- CREATE
--------

-- name: InsertLog :exec
INSERT INTO logs (site_id, level, message, attributes) VALUES (?, ?, ?, ?);

-- name: InsertEvent :exec
INSERT INTO events (site_id, batch, name, attributes, started_at, finished_at) VALUES (?, ?, ?, ?, ?, ?);

-- name: InsertError :exec
INSERT INTO errors (site_id, message, attributes) VALUES (?, ?, ?);

-------
-- READ
--------

-- name: SearchLogs :many
-- Newest first, only the site's when SiteID is set, and only those with Search in their text when it is set
SELECT * FROM logs
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchEvents :many
-- Newest first, only the site's when SiteID is set, only those named Type when it is set, and only those with
-- Search in their text when it is set
SELECT * FROM events
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR name = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(name), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchErrors :many
-- Newest first, only the site's when SiteID is set, only those of the kind Type when it is set, and only those
-- with Search in their text when it is set
SELECT * FROM errors
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR json_extract(attributes, '$.kind') = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: GetLog :one
SELECT * FROM logs WHERE id = sqlc.arg('ID');

-- name: GetEvent :one
SELECT * FROM events WHERE id = sqlc.arg('ID');

-- name: GetError :one
SELECT * FROM errors WHERE id = sqlc.arg('ID');

-- name: SearchSiteEvents :many
-- Newest first, the site's own events and its batch's, which is all a site may see, only those named Type when it is set
SELECT * FROM events
WHERE (site_id = sqlc.arg('SiteID') OR (site_id IS NULL AND batch = sqlc.arg('Batch')))
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR name = sqlc.narg('Type'))
-- Search looks in the names of all of them, but only in the attributes of the site's own, as a batch's are not shown
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(name), lower(sqlc.narg('Search'))) > 0
  OR (site_id = sqlc.arg('SiteID') AND instr(lower(attributes), lower(sqlc.narg('Search'))) > 0))
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: GetSiteEvent :one
SELECT * FROM events
WHERE id = sqlc.arg('ID') AND (site_id = sqlc.arg('SiteID') OR (site_id IS NULL AND batch = sqlc.arg('Batch')));

-- name: GetSiteLog :one
SELECT * FROM logs WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');

-- name: GetSiteError :one
SELECT * FROM errors WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');

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

-- name: RecentSyncEvents :many
-- The newest events of one of the site's sync sources, by the sync_type the sync events record
SELECT * FROM events
WHERE site_id = sqlc.arg('SiteID') AND json_extract(attributes, '$.sync_type') = sqlc.arg('SyncType')
ORDER BY id DESC
LIMIT sqlc.arg('Limit');

-- name: RecentSyncErrors :many
-- The newest errors of one of the site's sync sources
SELECT * FROM errors
WHERE site_id = sqlc.arg('SiteID') AND json_extract(attributes, '$.kind') = 'sync' AND json_extract(attributes, '$.sync_type') = sqlc.arg('SyncType')
ORDER BY id DESC
LIMIT sqlc.arg('Limit');
