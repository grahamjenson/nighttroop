
-- name: GetSyncSource :one
SELECT * FROM site_sync_sources where type = ? AND site_id = ? LIMIT 1;

-- name: GetAllSyncSourcesForSite :many
SELECT * FROM site_sync_sources where site_id = ?;


-- name: CreateSyncSource :exec
INSERT INTO site_sync_sources (
    site_id,
    type,
    url_or_username,
    password,
    region,
    enabled
)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: UpdateSyncSource :exec
UPDATE site_sync_sources SET 
    password = COALESCE(sqlc.narg('Password'), password),
    enabled =  COALESCE(sqlc.narg('Enabled'), enabled),
    url_or_username = COALESCE(sqlc.narg('UrlOrUsername'), url_or_username),
    region = COALESCE(sqlc.narg('Region'), region),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) 
WHERE type = sqlc.arg('Type') AND site_id = sqlc.arg('SiteID');

------
-- DELETE just removes some values, but we remember things that would slow down, to punish malicious activity
------

-- name: DeleteSyncSource :exec
UPDATE site_sync_sources SET 
    enabled = 0,
    url_or_username = NULL,
    password = NULL,
    previous_sync_errors = NULL,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE type = ? AND site_id = ?;

------
-- REPORTS
-------

-- Get All sync reports that:
-- 1. The site_sync_source is enabled
-- 2. The associated site has not been deleted
-- 3. The associated site has been seen in the last day
-- 4. The last attempted sync is either null or older than the sync interval

-- name: GetSyncReportsToRun :many
SELECT us.* FROM site_sync_sources us
JOIN sites u ON u.id = us.site_id
WHERE us.enabled = 1
  AND u.deleted_at IS NULL
  AND u.last_seen_at >= CAST(unixepoch('subsec') * 1000 AS INTEGER) - (24 * 60 * 60 * 1000) -- site has been seen in the last day
  AND (
      us.last_attempted_sync IS NULL
      OR us.last_attempted_sync <= CAST(unixepoch('subsec') * 1000 AS INTEGER) - (us.sync_interval_seconds * 1000)
  );

------ Updates to the Sync Reporting -----

-- name: UpdateSyncReportAttempt :exec
UPDATE site_sync_sources
SET
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_attempted_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE type = ? AND site_id = ?;



-- name: UpdateSyncSourceReportsSuccess :exec
UPDATE site_sync_sources
SET
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_successful_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_attempted_sync = CAST(unixepoch('subsec') * 1000 AS INTEGER),

    -- Counters
    total_syncs = total_syncs + 1,
    total_successful_syncs = total_successful_syncs + 1,
    
    consecutive_successful_syncs = consecutive_successful_syncs + 1,
    consecutive_failed_syncs = 0,

    -- Back to the default interval once a sync works again
    sync_interval_seconds = 60,
    sync_interval_seconds_reason = 'default',

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

WHERE type = sqlc.arg('Type') AND site_id = sqlc.arg('SiteID');


-- name: UpdateSyncSourceReportsFailure :exec
UPDATE site_sync_sources
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

    -- Back off: double the wait after every failure (bad status, timeout,
    -- bad credentials), capped at an hour, so a broken source is not hammered
    sync_interval_seconds = MIN(sync_interval_seconds * 2, 3600),
    sync_interval_seconds_reason = 'error_response',

    

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

WHERE type = sqlc.arg('Type') AND site_id = sqlc.arg('SiteID');
