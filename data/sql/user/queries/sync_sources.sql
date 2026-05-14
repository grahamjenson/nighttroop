
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

-- name: CreateSyncSource :exec

INSERT INTO sync_sources (type, password, enabled, url, username, dexcom_region) 
VALUES (?, ?, ?, ?, ?, ?);