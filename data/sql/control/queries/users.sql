
-- name: SubdomainExists :one
SELECT CAST(COUNT(*) > 0 AS BOOLEAN) FROM users WHERE subdomain = ?;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :one
INSERT INTO users (subdomain, backup_key_hash)
VALUES (?, ?) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE subdomain = ? AND deleted_at IS NULL;

-- Pass a list of users as a string to delete
-- name: DeleteUsers :many
UPDATE users
SET deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE subdomain IN (SELECT value FROM json_each(?))
RETURNING subdomain;

-- name: GetBatchUsers :many
SELECT * FROM users 
    WHERE batch = ? AND deleted_at IS NULL ORDER BY subdomain DESC;


-- name: GetInactiveUsers :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- Inactive Users are users that were once active but have stopped, so the assumption is they have data 
SELECT * FROM users 
WHERE deleted_at IS NULL AND (last_seen_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes') OR last_write_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'));

-- name: GetNeverSeenUsers :many
-- @param last_seen_at INT
-- @param last_write_at INT
-- finds deleted_at is null users where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago
SELECT * FROM users WHERE deleted_at IS NULL 
    AND (
            (last_seen_at IS NULL AND created_at < datetime('now', '-' || CAST(@last_seen_at AS REAL) || ' minutes')) 
            OR 
            (last_seen_at IS NOT NULL AND last_write_at IS NULL AND created_at < datetime('now', '-' || CAST(@last_write_at AS REAL) || ' minutes'))
        
        );

-- name: GetNotDeletedUsers :many
SELECT * FROM users WHERE deleted_at IS NULL;

-- name: UpdateUserDateFlags :exec
UPDATE users
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE subdomain = ?;


------------------
--- TESTING
------------------

-- name: TESTTruncateUsers :exec
DELETE FROM users;
