
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

-- name: CreateSyncSource :exec
INSERT INTO sync_sources (id, type, url, username, password, created_at, last_attempted_sync, last_successful_sync, consecutive_failed_syncs, total_failed_syncs, previous_sync_error) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);