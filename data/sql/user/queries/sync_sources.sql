
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

-- name: CreateSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url, username, dexcom_region) 
VALUES (?, ?, ?, ?, ?, ?);

-- HARD DELETE FOR SAFTEY
-- name: DeleteSyncSource :exec
DELETE FROM sync_sources WHERE id = ?;


-- name: UpsertSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url, username, dexcom_region)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(type) DO UPDATE SET
    password = excluded.password,
    enabled = excluded.enabled,
    url = excluded.url,
    username = excluded.username,
    dexcom_region = excluded.dexcom_region;