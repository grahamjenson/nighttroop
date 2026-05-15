
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

-- name: CreateSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url, username, dexcom_region) 
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateSyncSource :exec
UPDATE sync_sources SET password = ?, enabled = ?, url = ?, username = ?, dexcom_region = ? WHERE type = ?;

-- name: UpdateSyncSourceWithoutPassword :exec
UPDATE sync_sources SET enabled = ?, url = ?, username = ?, dexcom_region = ? WHERE type = ?;


-- HARD DELETE FOR SAFTEY
-- name: DeleteSyncSource :exec
DELETE FROM sync_sources WHERE type = ?;

