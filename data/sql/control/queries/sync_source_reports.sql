------ Updates to the Sync Reporting -----
-- name: CreateOrUpdateSyncSourceReport :one
INSERT INTO sync_source_reports (
    uniq_hash,
    last_attempted_sync,
    updated_at
)
VALUES (
    sqlc.arg(uniq_hash),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT(uniq_hash) DO UPDATE SET
    updated_at = CURRENT_TIMESTAMP,
    last_attempted_sync = CURRENT_TIMESTAMP
RETURNING *;

-- name: UpdateSyncSourceReportsSuccess :exec
UPDATE sync_source_reports
SET
    updated_at = CURRENT_TIMESTAMP,

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

WHERE uniq_hash = sqlc.arg(uniq_hash);


-- name: UpdateSyncSourceReportsFailure :exec
UPDATE sync_source_reports
SET
    updated_at = CURRENT_TIMESTAMP,
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

WHERE uniq_hash = sqlc.arg(uniq_hash);
