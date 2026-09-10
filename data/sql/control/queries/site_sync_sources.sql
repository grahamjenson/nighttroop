
-- name: GetAllSiteSyncSourcesForSite :many
SELECT * FROM site_sync_sources where site_id = ?;

-- name: GetAllSiteSyncSourcesForUniqueHash :many
SELECT uss.* FROM site_sync_sources uss JOIN sites u ON u.id = uss.site_id WHERE uss.unique_hash = ? AND u.deleted_at IS NULL;

-----
-- Nighscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO site_sync_sources (site_id, type, password, enabled, url_or_username, unique_hash) 
VALUES (?, "NIGHTSCOUT", ?, ?, ?, ?);

-- name: UpdateNightscoutSyncSource :exec
UPDATE site_sync_sources SET password = ?, enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND site_id = ?;

-- name: UpdateNightscoutSyncSourceWithoutPassword :exec
UPDATE site_sync_sources SET enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND site_id = ?;

-- name: DisableNightscoutSyncSource :exec
UPDATE site_sync_sources SET enabled = 0, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND site_id = ?;

-----
-- DEXCOM
------
-- name: CreateDexcomSyncSource :exec
INSERT INTO site_sync_sources (site_id, type, password, enabled, url_or_username, dexcom_region, unique_hash) 
VALUES (?, "DEXCOM", ?, ?, ?, ?, ?);


-- name: UpdateDexcomSyncSource :exec
UPDATE site_sync_sources SET password = ?, enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND site_id = ?;

-- name: UpdateDexcomSyncSourceWithoutPassword :exec
UPDATE site_sync_sources SET enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND site_id = ?;


-- DELETE is real for security
-- name: DeleteSyncSource :exec
DELETE FROM site_sync_sources WHERE type = ? and site_id = ?;




