
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;


-- ToBeSyncedSources gets enabled and ready for syncing souces using sync_interval_seconds and last_attempted_sync to determine when the next sync should be attempted
-- name: ToBeSyncedSources :many
SELECT * FROM sync_sources WHERE enabled = 1 AND (last_attempted_sync IS NULL OR last_attempted_sync <= datetime('now', '-' || sync_interval_seconds || ' seconds'));

-----
-- Nighscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url_or_username) 
VALUES ("NIGHTSCOUT", ?, ?, ?);

-- name: UpdateNightscoutSyncSource :exec
UPDATE sync_sources SET password = ?, enabled = ?, url_or_username = ?,  updated_at = CURRENT_TIMESTAMP WHERE type = "NIGHTSCOUT";

-- name: UpdateNightscoutSyncSourceWithoutPassword :exec
UPDATE sync_sources SET enabled = ?, url_or_username = ?,  updated_at = CURRENT_TIMESTAMP WHERE type = "NIGHTSCOUT";

-- name: DisableNightscoutSyncSource :exec
UPDATE sync_sources SET enabled = 0, updated_at = CURRENT_TIMESTAMP WHERE type = "NIGHTSCOUT";

-----
-- DEXCOM
------
-- name: CreateDexcomSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url_or_username, dexcom_region) 
VALUES ("DEXCOM", ?, ?, ?, ?);


-- name: UpdateDexcomSyncSource :exec
UPDATE sync_sources SET password = ?, enabled = ?, url_or_username = ?, dexcom_region = ?,  updated_at = CURRENT_TIMESTAMP WHERE type = "DEXCOM";

-- name: UpdateDexcomSyncSourceWithoutPassword :exec
UPDATE sync_sources SET enabled = ?, url_or_username = ?, dexcom_region = ?,  updated_at = CURRENT_TIMESTAMP WHERE type = "DEXCOM";





-- DELETE is real for security
-- name: DeleteSyncSource :exec
DELETE FROM sync_sources WHERE type = ?;


