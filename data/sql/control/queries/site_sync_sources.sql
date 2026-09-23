
-- name: GetAllSiteSyncSourcesForSite :many
SELECT * FROM site_sync_sources where site_id = ?;

-----
-- Nightscout
------ 

-- name: CreateNightscoutSyncSource :exec
INSERT INTO site_sync_sources (
    site_id,
    type,
    password,
    enabled,
    properties
)
VALUES (
    ?,
    'NIGHTSCOUT',
    ?,
    ?,
    json_object('url', ?)
);

-- name: UpdateNightscoutSyncSource :exec
UPDATE site_sync_sources SET 
    password = CASE
        WHEN sqlc.narg(Password) IS NULL OR sqlc.narg(Password) IN ('', '********') THEN password -- dont update if empty or masked
        ELSE sqlc.narg(Password)
    END,
    enabled =  sqlc.narg(Enabled),
    properties = json_set(
        properties,
        '$.url', COALESCE(NULLIF(sqlc.narg(Url), ''), json_extract(properties, '$.url'))
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) 
WHERE type = 'NIGHTSCOUT' AND site_id = sqlc.arg(SiteID);



-----
-- DEXCOM
------


-- name: CreateDexcomSyncSource :exec
INSERT INTO site_sync_sources (
    site_id, 
    type, 
    password, 
    enabled, 
    properties
) 
VALUES (
    ?, 
    'DEXCOM', 
    ?, 
    ?, 
    json_object('username', sqlc.arg(Username), 'region', sqlc.arg(Region))
);

-- name: UpdateDexcomSyncSource :exec
WITH params AS (SELECT COALESCE(sqlc.narg(Password), '') AS password)
UPDATE site_sync_sources SET 
    password = CASE
        WHEN params.password IN ('', '********') THEN password -- dont update if empty or masked
        ELSE params.password
    END,
    enabled = sqlc.narg(Enabled),
    properties = json_set(
        properties,
        '$.username', COALESCE(NULLIF(sqlc.narg(Username), ''), json_extract(properties, '$.username')),
        '$.region',   COALESCE(NULLIF(sqlc.narg(Region), ''), json_extract(properties, '$.region'))
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE type = 'DEXCOM' AND site_id = sqlc.arg(SiteID);


------
-- VIRTUAL
------


-- name: CreateVirtualSyncSource :exec
INSERT INTO site_sync_sources (
    site_id, 
    type,  
    enabled
) 
VALUES (
    ?, 
    'VIRTUAL', 
    ?
);

-- name: UpdateVirtualSyncSource :exec
UPDATE site_sync_sources SET 
    enabled = sqlc.narg(Enabled),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) 
WHERE type = 'VIRTUAL' AND site_id =  sqlc.arg(SiteID);

------
-- DELETE is real for security
------

-- name: DeleteSyncSource :exec
DELETE FROM site_sync_sources WHERE type = ? and site_id = ?;

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
WHERE site_id = sqlc.arg(SiteID) AND type = sqlc.arg(Type);



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

WHERE site_id = sqlc.arg(SiteID) AND type = sqlc.arg(Type);


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

WHERE site_id = sqlc.arg(SiteID) AND type = sqlc.arg(Type);
