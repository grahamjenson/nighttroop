
-------
-- CREATE 
--------

-- name: CreateSite :one
INSERT INTO sites (id, backup_key_hash)
VALUES (?, ?) RETURNING *;

-------
-- READ 
--------

-- name: SiteExists :one
SELECT CAST(COUNT(*) > 0 AS BOOLEAN) FROM sites WHERE id = ?;

-- name: GetAllSites :many
SELECT * FROM sites;

-- name: GetSite :one
SELECT * FROM sites WHERE id = ? AND state = 'live';


-- name: SearchSites :many
SELECT * FROM sites
ORDER BY created_at DESC, id
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchSitesByLastWrite :many
-- The longest since written to first, the first to archive; a site never written to counts from when it was made
SELECT * FROM sites
ORDER BY COALESCE(last_write_at, created_at) ASC, id
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: SearchSitesByDBSize :many
-- The largest site DB first, the first to truncate; a site not yet measured is last
SELECT * FROM sites
ORDER BY estimated_db_size_kb IS NULL, estimated_db_size_kb DESC, created_at DESC, id
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: GetSiteInAnyState :one
SELECT * FROM sites WHERE id = ?;

-- name: GetLiveAndDisabledSites :many
SELECT * FROM sites WHERE state IN ('live', 'disabled');


-- name: GetSitesToDisable :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- Inactive Sites are sites that were once active but have stopped, so the assumption is they have data 
SELECT * FROM sites 
WHERE state = 'live' AND (last_seen_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes') OR last_write_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'));

-- name: GetSitesToArchive :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- Inactive Sites are sites that were once active but have stopped, so the assumption is they have data 
SELECT * FROM sites 
WHERE state = 'disabled' AND (last_seen_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes') OR last_write_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'));

-------
-- UPDATE
--------

-- name: UpdateSiteDateFlags :exec
UPDATE sites
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at),
    estimated_db_size_kb = COALESCE(?, estimated_db_size_kb)
WHERE id = ?;


-------
-- DELETE
--------

-- name: DisableSites :many
UPDATE sites
SET 
    disabled_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'disabled'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;

-- name: DisableSite :execresult
UPDATE sites
SET
    disabled_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'disabled'
WHERE id = ? AND state = 'live';

-- name: ArchiveSites :many
UPDATE sites
SET 
    archived_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'archived'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;


-- Pass a list of sites as a string to delete
-- name: DeleteSites :many
UPDATE sites
SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'deleted'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;


--------------------------------
-- BATCH (by first letter of id)
--------------------------------


-- name: GetBatchSitesToBackUp :many
-- Get sites that are live  or diabled
-- We must backup even empty sites because a rehydrated control will expect them
SELECT * FROM sites WHERE batch = ? AND state IN ('live', 'disabled') ORDER BY id DESC;


-- name: GetSitesToDeleteInBatch :many
-- finds deleted_at is null sites where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago
--  1. Delete if never even logged in was X mins ago
--  2. Delete if never written and last_seen_at was X mins ago
SELECT * FROM sites WHERE 
    state IN ('live', 'disabled') AND -- must be live or disabled
    batch = @batch AND -- within this batch
    last_write_at IS NULL AND -- must not have been written anything (this is important)
    (
        (last_seen_at IS NULL     AND created_at < (unixepoch('subsec') * 1000 - CAST(@last_seen_at AS REAL) * 60 * 1000) )
        OR 
        (last_seen_at IS NOT NULL AND last_seen_at < (unixepoch('subsec') * 1000 - CAST(@last_seen_at AS REAL) * 60 * 1000))
    
    );








