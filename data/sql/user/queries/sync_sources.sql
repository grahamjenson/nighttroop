
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

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


