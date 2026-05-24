------ MassCreationOfSyncReports

-- name: CreateSyncReportsForAllSyncSources :exec
INSERT OR IGNORE INTO sync_source_reports (unique_hash)
SELECT us.unique_hash
FROM user_sync_sources us
WHERE NOT EXISTS (
    SELECT 1
    FROM sync_source_reports sr
    WHERE sr.unique_hash = us.unique_hash
);


-- return all sync reports that have an associated enabled user_sync_souce
-- AND have last_attempted_sync is more than sync_interval_seconds ago or last_attempted_sync is null (never attempted)

-- name: GetSyncReportsToRun :many
SELECT
    ssr.unique_hash,
    CAST(MIN(us.type) AS TEXT) AS type
FROM sync_source_reports ssr
JOIN user_sync_sources us
    ON us.unique_hash = ssr.unique_hash
WHERE us.enabled = 1
  AND (
      ssr.last_attempted_sync IS NULL
      OR ssr.last_attempted_sync <= CAST(unixepoch('subsec') * 1000 AS INTEGER) - (ssr.sync_interval_seconds * 1000)
  )
GROUP BY ssr.id;

------ Updates to the Sync Reporting -----
-- name: UpdateSyncReportAttempt :exec
UPDATE sync_source_reports
SET
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_attempted_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE unique_hash = sqlc.arg(unique_hash);



-- name: UpdateSyncSourceReportsSuccess :exec
UPDATE sync_source_reports
SET
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_successful_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_attempted_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),

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
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_failed_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    previous_sync_errors = sqlc.arg(error_messages),
    last_attempted_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    
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
