-- name: CreateAdmin :one
INSERT INTO admins (name, password, role, totp_secret) VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetAdmin :one
SELECT * FROM admins WHERE name = ?;



-- name: TruncateEvents :exec
DELETE FROM events WHERE id < (SELECT MAX(e.id) - ? FROM events as e);


-- name: TESTTruncateUsers :exec
DELETE FROM users;



-----------------------------
-- USERS
-----------------------------

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :one
INSERT INTO users (subdomain, backup_key_hash)
VALUES (?, ?) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE subdomain = ? AND deleted_at IS NULL;

-- Pass a list of users as a string to delete
-- name: DeleteUsers :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE subdomain IN (SELECT value FROM json_each(?));

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

-- name: UpdateUserLastSeenAt :exec
UPDATE users
SET last_seen_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;

-- name: UpdateUserLastWriteAt :exec
UPDATE users
SET last_write_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;