
-- name: SubdomainExists :one
SELECT CAST(COUNT(*) > 0 AS BOOLEAN) FROM sites WHERE subdomain = ?;

-- name: GetAllSites :many
SELECT * FROM sites;

-- name: CreateSite :one
INSERT INTO sites (subdomain, backup_key_hash)
VALUES (?, ?) RETURNING *;

-- name: GetSite :one
SELECT * FROM sites WHERE subdomain = ? AND deleted_at IS NULL;

-- Pass a list of sites as a string to delete
-- name: DeleteSites :many
UPDATE sites
SET deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE subdomain IN (SELECT value FROM json_each(?))
RETURNING subdomain;

-- name: GetBatchSites :many
SELECT * FROM sites 
    WHERE batch = ? AND deleted_at IS NULL ORDER BY subdomain DESC;


-- name: GetInactiveSites :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- Inactive Sites are sites that were once active but have stopped, so the assumption is they have data 
SELECT * FROM sites 
WHERE deleted_at IS NULL AND (last_seen_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes') OR last_write_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'));

-- name: GetNeverUsedSites :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- finds deleted_at is null sites where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago
SELECT * FROM sites WHERE deleted_at IS NULL 
    AND (
            (last_seen_at IS NULL AND created_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes')) 
            OR 
            (last_seen_at IS NOT NULL AND last_write_at IS NULL AND created_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'))
        
        );

-- name: GetNotDeletedSites :many
SELECT * FROM sites WHERE deleted_at IS NULL;

-- name: UpdateSiteDateFlags :exec
UPDATE sites
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE subdomain = ?;

