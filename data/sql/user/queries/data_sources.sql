
-- name: AllDataSources :many
SELECT * FROM data_sources ORDER BY created_at DESC;

-- name: CreateDataSource :exec
INSERT INTO data_sources (id, type, url, username, password, created_at, last_attempted_sync, last_successful_sync, consecutive_failed_syncs, total_failed_syncs, previous_sync_error) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);