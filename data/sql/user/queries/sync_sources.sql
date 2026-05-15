
-- name: AllSyncSources :many
SELECT * FROM sync_sources ORDER BY created_at DESC;

-- name: AllEnabledSyncSources :many
SELECT * FROM sync_sources WHERE enabled = 1;

-----
-- Nighscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, url, hostname) 
VALUES ("NIGHTSCOUT", ?, ?, ?, ?);

-- name: UpdateNightscoutSyncSource :exec
UPDATE sync_sources SET password = ?, enabled = ?, url = ?, hostname = ? WHERE type = "NIGHTSCOUT";

-- name: UpdateNightscoutSyncSourceWithoutPassword :exec
UPDATE sync_sources SET enabled = ?, url = ?, hostname = ? WHERE type = "NIGHTSCOUT";

-----
-- DEXCOM
------ 
-- name: CreateDexcomSyncSource :exec
INSERT INTO sync_sources (type, password, enabled, username, dexcom_region) 
VALUES ("DEXCOM", ?, ?, ?, ?);


-- name: UpdateDexcomSyncSource :exec
UPDATE sync_sources SET password = ?, enabled = ?, username = ?, dexcom_region = ? WHERE type = "DEXCOM";

-- name: UpdateDexcomSyncSourceWithoutPassword :exec
UPDATE sync_sources SET enabled = ?, username = ?, dexcom_region = ? WHERE type = "DEXCOM";

-- HARD DELETE FOR SAFTEY
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
