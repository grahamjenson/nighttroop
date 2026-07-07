
-- name: GetAllSiteSyncSourcesForSite :many
SELECT * FROM site_sync_sources where subdomain = ?;

-- name: GetAllSiteSyncSourcesForUniqueHash :many
SELECT uss.* FROM site_sync_sources uss JOIN sites u ON u.subdomain = uss.subdomain WHERE uss.unique_hash = ? AND u.deleted_at IS NULL;

-----
-- Nighscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO site_sync_sources (subdomain, type, password, enabled, url_or_username, unique_hash) 
VALUES (?, "NIGHTSCOUT", ?, ?, ?, ?);

-- name: UpdateNightscoutSyncSource :exec
UPDATE site_sync_sources SET password = ?, enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-- name: UpdateNightscoutSyncSourceWithoutPassword :exec
UPDATE site_sync_sources SET enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-- name: DisableNightscoutSyncSource :exec
UPDATE site_sync_sources SET enabled = 0, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-----
-- DEXCOM
------
-- name: CreateDexcomSyncSource :exec
INSERT INTO site_sync_sources (subdomain, type, password, enabled, url_or_username, dexcom_region, unique_hash) 
VALUES (?, "DEXCOM", ?, ?, ?, ?, ?);


-- name: UpdateDexcomSyncSource :exec
UPDATE site_sync_sources SET password = ?, enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND subdomain = ?;

-- name: UpdateDexcomSyncSourceWithoutPassword :exec
UPDATE site_sync_sources SET enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND subdomain = ?;


-- DELETE is real for security
-- name: DeleteSyncSource :exec
DELETE FROM site_sync_sources WHERE type = ? and subdomain = ?;




