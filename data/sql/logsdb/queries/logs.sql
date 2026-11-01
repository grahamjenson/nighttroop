-------
-- CREATE
--------

-- name: InsertLog :exec
INSERT INTO logs (site_id, level, message, attributes, request_id, job_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: InsertEvent :exec
INSERT INTO events (site_id, batch, name, attributes, started_at, finished_at, request_id, job_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertError :exec
INSERT INTO errors (site_id, message, attributes, request_id, job_id) VALUES (?, ?, ?, ?, ?);

-------
-- READ
--------

-- name: SearchLogs :many
-- Newest first, only the site's, request's and job's when SiteID, RequestID and JobID are set, and only those with
-- Search in their text when it is set
SELECT * FROM logs
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
AND (CAST(sqlc.narg('RequestID') AS TEXT) IS NULL OR request_id = sqlc.narg('RequestID'))
AND (CAST(sqlc.narg('JobID') AS TEXT) IS NULL OR job_id = sqlc.narg('JobID'))
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchEvents :many
-- Newest first, filtered as SearchLogs is, and only those named Type when it is set
SELECT * FROM events
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR name = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(name), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
AND (CAST(sqlc.narg('RequestID') AS TEXT) IS NULL OR request_id = sqlc.narg('RequestID'))
AND (CAST(sqlc.narg('JobID') AS TEXT) IS NULL OR job_id = sqlc.narg('JobID'))
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchErrors :many
-- Newest first, filtered as SearchLogs is, and only those of the kind Type when it is set
SELECT * FROM errors
WHERE (CAST(sqlc.narg('SiteID') AS TEXT) IS NULL OR site_id = sqlc.narg('SiteID'))
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR json_extract(attributes, '$.kind') = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
AND (CAST(sqlc.narg('RequestID') AS TEXT) IS NULL OR request_id = sqlc.narg('RequestID'))
AND (CAST(sqlc.narg('JobID') AS TEXT) IS NULL OR job_id = sqlc.narg('JobID'))
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: GetLog :one
SELECT * FROM logs WHERE id = sqlc.arg('ID');

-- name: GetEvent :one
SELECT * FROM events WHERE id = sqlc.arg('ID');

-- name: GetError :one
SELECT * FROM errors WHERE id = sqlc.arg('ID');

-- The site's own rows, for its logs pages. SearchLogs and SearchErrors with SiteID set match the same rows, but
-- their optional SiteID keeps SQLite from using the (site_id, id) index, so they walk every site's rows.

-- name: SearchSiteLogs :many
-- Newest first, only those with Search in their text when it is set
SELECT * FROM logs
WHERE site_id = sqlc.arg('SiteID')
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchSiteErrors :many
-- Newest first, only those of the kind Type when it is set, and only those with Search in their text when it is set
SELECT * FROM errors
WHERE site_id = sqlc.arg('SiteID')
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR json_extract(attributes, '$.kind') = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(message), lower(sqlc.narg('Search'))) > 0 OR instr(lower(attributes), lower(sqlc.narg('Search'))) > 0)
ORDER BY id DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchSiteEvents :many
-- Newest first, the site's own events and its batch's, which is all a site may see, only those named Type when it is set.
-- Each half reads its own index newest first and SQLite merges the two, rather than sorting all of the site's events.
SELECT own.* FROM events own
WHERE own.site_id = sqlc.arg('SiteID')
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR own.name = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(own.name), lower(sqlc.narg('Search'))) > 0
  OR instr(lower(own.attributes), lower(sqlc.narg('Search'))) > 0)
UNION ALL
-- Search looks only in a batch's names, as its attributes are not shown
SELECT batched.* FROM events batched
WHERE batched.site_id IS NULL AND batched.batch = sqlc.arg('Batch')
AND (CAST(sqlc.narg('Type') AS TEXT) IS NULL OR batched.name = sqlc.narg('Type'))
AND (CAST(sqlc.narg('Search') AS TEXT) IS NULL OR instr(lower(batched.name), lower(sqlc.narg('Search'))) > 0)
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
-- Deletes all but the newest Keep rows. A row's id is one more than the newest's and rows are only deleted oldest
-- first, so the newest Keep are those within Keep of the newest id, which is read from the end of the table rather
-- than walking back over Keep rows on every trim.
DELETE FROM logs WHERE id <= (SELECT MAX(newest.id) FROM logs newest) - CAST(sqlc.arg('Keep') AS INTEGER);

-- name: TrimEvents :execresult
-- Deletes all but the newest Keep rows, as TrimLogs does
DELETE FROM events WHERE id <= (SELECT MAX(newest.id) FROM events newest) - CAST(sqlc.arg('Keep') AS INTEGER);

-- name: TrimErrors :execresult
-- Deletes all but the newest Keep rows, as TrimLogs does
DELETE FROM errors WHERE id <= (SELECT MAX(newest.id) FROM errors newest) - CAST(sqlc.arg('Keep') AS INTEGER);

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
