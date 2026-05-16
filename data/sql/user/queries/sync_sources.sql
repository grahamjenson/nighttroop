
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





-- DELETE ALL COLUMNS EXCEPT the 
-- TODO: we need to remember the sync_interval_seconds so that if they have been punished for bad servers they are not reset with a delete
-- name: DeleteSyncSource :exec
DELETE FROM sync_sources WHERE type = ?;



------ Updates to the Sync Reporting -----

-- name: UpdateSyncSuccess :exec
UPDATE sync_sources
SET
    last_attempted_sync = CURRENT_TIMESTAMP,
    last_successful_sync = CURRENT_TIMESTAMP,

    consecutive_failed_syncs = 0,

    success_ema = CASE
        WHEN success_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * success_ema)
    END,

    total_ema = CASE
        WHEN total_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * total_ema)
    END,

    last_sync_duration = sqlc.arg(elapsed_ms),

    previous_sync_error = NULL

WHERE type = sqlc.arg(type);


-- name: UpdateSyncFailure :exec
UPDATE sync_sources
SET
    last_attempted_sync = CURRENT_TIMESTAMP,

    consecutive_failed_syncs = consecutive_failed_syncs + 1,
    total_failed_syncs = total_failed_syncs + 1,

    failure_ema = CASE
        WHEN failure_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * failure_ema)
    END,

    total_ema = CASE
        WHEN total_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * total_ema)
    END,

    previous_sync_error = sqlc.arg(error_message),

    last_sync_duration = sqlc.arg(elapsed_ms)

WHERE type = sqlc.arg(type);
