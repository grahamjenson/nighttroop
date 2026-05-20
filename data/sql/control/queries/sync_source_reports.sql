------ Updates to the Sync Reporting -----
-- name: CreateOrUpdateSyncSourceReport :one
INSERT INTO sync_source_reports (
    unique_hash,
    last_attempted_sync,
    updated_at
)
VALUES (
    sqlc.arg(unique_hash),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT(unique_hash) DO UPDATE SET
    updated_at = CURRENT_TIMESTAMP,
    last_attempted_sync = CURRENT_TIMESTAMP
RETURNING *;



-- name: UpdateSyncSourceReportsSuccess :exec
UPDATE sync_source_reports
SET
    updated_at = CURRENT_TIMESTAMP,
    last_successful_sync = CURRENT_TIMESTAMP,

    -- Counters
    total_syncs = total_syncs + 1,
    total_successful_syncs = total_successful_syncs + 1,
    
    consecutive_successful_syncs = consecutive_successful_syncs + 1,
    consecutive_failed_syncs = 0,

    -- Timing
    last_sync_duration = sqlc.arg(elapsed_ms),

    success_ema = CASE
        WHEN success_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * success_ema)
    END,

    total_ema = CASE
        WHEN total_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * total_ema)
    END

WHERE unique_hash = sqlc.arg(unique_hash);


-- name: UpdateSyncSourceReportsFailure :exec
UPDATE sync_source_reports
SET
    updated_at = CURRENT_TIMESTAMP,
    last_failed_sync = CURRENT_TIMESTAMP,
    previous_sync_error = sqlc.arg(error_message),
    
    -- Counters 
    total_syncs = total_syncs + 1,
    total_failed_syncs = total_failed_syncs + 1,
    
    consecutive_failed_syncs = consecutive_failed_syncs + 1,
    consecutive_successful_syncs = 0,

    

    -- Timing
    last_sync_duration = sqlc.arg(elapsed_ms),

    failure_ema = CASE
        WHEN failure_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * failure_ema)
    END,

    total_ema = CASE
        WHEN total_ema IS NULL THEN CAST(sqlc.arg(elapsed_ms) AS REAL)
        ELSE (0.2 * CAST(sqlc.arg(elapsed_ms) AS REAL))
            + (0.8 * total_ema)
    END

WHERE unique_hash = sqlc.arg(unique_hash);
