
-- name: GetAllUserSyncSourcesForUser :many
SELECT * FROM user_sync_sources where subdomain = ?;

-- name: GetAllUserSyncSourcesForUniqueHash :many
SELECT * FROM user_sync_sources where unique_hash = ?;

-----
-- Nighscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO user_sync_sources (subdomain, type, password, enabled, url_or_username, unique_hash) 
VALUES (?, "NIGHTSCOUT", ?, ?, ?, ?);

-- name: UpdateNightscoutSyncSource :exec
UPDATE user_sync_sources SET password = ?, enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-- name: UpdateNightscoutSyncSourceWithoutPassword :exec
UPDATE user_sync_sources SET enabled = ?, url_or_username = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-- name: DisableNightscoutSyncSource :exec
UPDATE user_sync_sources SET enabled = 0, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "NIGHTSCOUT" AND subdomain = ?;

-----
-- DEXCOM
------
-- name: CreateDexcomSyncSource :exec
INSERT INTO user_sync_sources (subdomain, type, password, enabled, url_or_username, dexcom_region, unique_hash) 
VALUES (?, "DEXCOM", ?, ?, ?, ?, ?);


-- name: UpdateDexcomSyncSource :exec
UPDATE user_sync_sources SET password = ?, enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND subdomain = ?;

-- name: UpdateDexcomSyncSourceWithoutPassword :exec
UPDATE user_sync_sources SET enabled = ?, url_or_username = ?, dexcom_region = ?, unique_hash = ?, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE type = "DEXCOM" AND subdomain = ?;


-- DELETE is real for security
-- name: DeleteSyncSource :exec
DELETE FROM user_sync_sources WHERE type = ? and subdomain = ?;




