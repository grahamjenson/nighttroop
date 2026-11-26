
-------
-- CREATE 
--------

-- name: CreateSite :one
INSERT INTO sites (id, backup_key_hash)
VALUES (?, ?) RETURNING *;

-------
-- READ 
--------

-- name: SiteExists :one
SELECT CAST(COUNT(*) > 0 AS BOOLEAN) FROM sites WHERE id = ?;

-- name: GetAllSites :many
SELECT * FROM sites;

-- name: GetSite :one
SELECT * FROM sites WHERE id = ? AND state = 'live';

-- name: GetSiteDBSizeLimits :many
-- Each site's DB size limit, for the truncate job to find the sites far over theirs without a query for each
SELECT id, max_db_size_kb_limit FROM sites WHERE max_db_size_kb_limit IS NOT NULL;


-- name: SearchSites :many
SELECT * FROM sites
ORDER BY created_at DESC, id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: SearchSitesByLastWrite :many
-- The longest since written to first, the first to archive; a site never written to counts from when it was made
SELECT * FROM sites
ORDER BY COALESCE(last_write_at, created_at) ASC, id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: SearchSitesByDBSize :many
-- The largest site DB first, the first to truncate; a site not yet measured is last
SELECT * FROM sites
ORDER BY estimated_db_size_kb IS NULL, estimated_db_size_kb DESC, created_at DESC, id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: SearchArchivedSites :many
-- The archived sites, which can be hydrated, the most recently archived first
SELECT * FROM sites
WHERE state = 'archived'
ORDER BY archived_at DESC, id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: CountActiveSites :one
-- The live sites seen since Since, in milliseconds
SELECT COUNT(*) FROM sites WHERE state = 'live' AND last_seen_at >= sqlc.arg('Since');

-- name: CountSitesCreated :one
-- The sites created since Since, in milliseconds, whatever their state now
SELECT COUNT(*) FROM sites WHERE created_at >= sqlc.arg('Since');

-- name: CountSitesByState :many
SELECT state, COUNT(*) AS count FROM sites GROUP BY state ORDER BY state;

-- name: GetSiteInAnyState :one
SELECT * FROM sites WHERE id = ?;

-- name: GetSiteIDsWithAlarms :many
-- The live sites with an alarm turned on, for the site or for one of its enabled identities, which are checked periodically
SELECT id FROM sites WHERE state = 'live' AND (
    json_extract(settings_json, '$.alarms.urgent_low.enabled') OR json_extract(settings_json, '$.alarms.low.enabled') OR json_extract(settings_json, '$.alarms.urgent_high.enabled') OR json_extract(settings_json, '$.alarms.high.enabled') OR json_extract(settings_json, '$.alarms.no_data.enabled') OR json_extract(settings_json, '$.alarms.device_battery_low.enabled')
    OR id IN (
        SELECT site_id FROM site_identities WHERE enabled AND (
            json_extract(settings_json, '$.alarms.urgent_low.enabled') OR json_extract(settings_json, '$.alarms.low.enabled') OR json_extract(settings_json, '$.alarms.urgent_high.enabled') OR json_extract(settings_json, '$.alarms.high.enabled') OR json_extract(settings_json, '$.alarms.no_data.enabled') OR json_extract(settings_json, '$.alarms.device_battery_low.enabled')
        )
    )
);

-- name: GetLiveAndDisabledSites :many
SELECT * FROM sites WHERE state IN ('live', 'disabled');


-------
-- UPDATE
--------

-- A backup key works once: using it replaces its hash, only while the hash is still the one checked, so two uses of
-- the same key cannot both succeed
-- name: RotateSiteBackupKey :execresult
UPDATE sites SET backup_key_hash = sqlc.arg('NewHash')
WHERE id = sqlc.arg('ID') AND backup_key_hash = sqlc.arg('OldHash');

-- name: UpdateSiteDateFlags :exec
UPDATE sites
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at),
    estimated_db_size_kb = COALESCE(?, estimated_db_size_kb)
WHERE id = ?;

-- name: UpdateSiteAbuseScore :exec
-- Only a score that changed is written
UPDATE sites SET abuse_score = sqlc.arg('AbuseScore')
WHERE id = sqlc.arg('ID') AND abuse_score IS NOT sqlc.arg('AbuseScore');

-- name: UpdateSiteDBSize :exec
-- Only a size that changed is written, so measuring every site rewrites none of the ones that did not
UPDATE sites SET estimated_db_size_kb = sqlc.arg('EstimatedDbSizeKb')
WHERE id = sqlc.arg('ID') AND estimated_db_size_kb IS NOT sqlc.arg('EstimatedDbSizeKb');

-- name: UpdateSiteBackedUp :exec
-- The site's backup copy was made now, at BackupSizeKb on disk
UPDATE sites SET backed_up_at = CAST(unixepoch('subsec') * 1000 AS INTEGER), backup_size_kb = CAST(sqlc.arg('BackupSizeKb') AS INTEGER)
WHERE id = sqlc.arg('ID');

-- name: UpdateSitesBackupUploaded :exec
-- The sites were uploaded to S3 now
UPDATE sites SET backup_uploaded_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id IN (sqlc.slice('IDs'));


-------
-- DELETE
--------

-- name: DisableSites :many
UPDATE sites
SET 
    disabled_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'disabled'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;

-- name: DisableSite :execresult
UPDATE sites
SET
    disabled_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'disabled'
WHERE id = ? AND state = 'live';

-- name: EnableSite :execresult
-- Brings a disabled site back live; its DB is still here
UPDATE sites
SET
    disabled_at = NULL,
    state = 'live'
WHERE id = ? AND state = 'disabled';

-- name: ArchiveSites :many
UPDATE sites
SET 
    archived_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'archived'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;


-- name: HydrateSite :execresult
-- Brings an archived site back live, as if just seen and written to, so it is not disabled again straight away
UPDATE sites
SET
    state = 'live',
    disabled_at = NULL,
    archived_at = NULL,
    last_seen_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_write_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    -- Its backup is uploaded again soon
    backup_uploaded_at = NULL
WHERE id = ? AND state = 'archived';

-- Pass a list of sites as a string to delete
-- name: DeleteSites :many
UPDATE sites
SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    state = 'deleted'
WHERE id IN (SELECT value FROM json_each(?))
RETURNING id;


--------------------------------
-- BACKUP
--------------------------------


-- name: GetSitesToBackUp :many
-- The live and disabled sites to back up, the longest waiting first: those written since their last backup, copied
-- before BackedUpBefore, and, while Uploading, those whose last copy is not uploaded: with S3 off no copy uploads, so
-- such a site would be copied again each tick. A site's last write is recorded at most every 10 minutes, so one written
-- in the 10 minutes after its last copy, its last write before it, is written since too.
SELECT * FROM sites
WHERE state IN ('live', 'disabled') AND (
    (last_write_at IS NOT NULL AND (backed_up_at IS NULL OR (last_write_at + 600000 > backed_up_at
        AND backed_up_at < CAST(sqlc.arg('BackedUpBefore') AS INTEGER))))
    OR (CAST(sqlc.arg('Uploading') AS BOOLEAN) AND backup_uploaded_at < backed_up_at))
ORDER BY COALESCE(backed_up_at, 0), id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);


-- name: GetSitesNotUploaded :many
-- The live and disabled sites not uploaded yet, e.g. a new or restored site
SELECT * FROM sites WHERE state IN ('live', 'disabled') AND backup_uploaded_at IS NULL ORDER BY id;


-- name: GetSitesToDelete :many
-- finds deleted_at is null sites where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago
--  1. Delete if never even logged in was X mins ago
--  2. Delete if never written and last_seen_at was X mins ago
SELECT * FROM sites WHERE 
    state IN ('live', 'disabled') AND -- must be live or disabled
    last_write_at IS NULL AND -- must not have been written anything (this is important)
    (
        (last_seen_at IS NULL     AND created_at < (unixepoch('subsec') * 1000 - CAST(@last_seen_at AS REAL) * 60 * 1000) )
        OR 
        (last_seen_at IS NOT NULL AND last_seen_at < (unixepoch('subsec') * 1000 - CAST(@last_seen_at AS REAL) * 60 * 1000))
    
    );









-- name: UpdateSiteTableSizes :exec
UPDATE sites SET table_sizes_json = sqlc.arg('TableSizesJson') WHERE id = sqlc.arg('ID');

-- name: UpdateSiteSummariesCheckedAt :exec
UPDATE sites SET summaries_checked_at = sqlc.arg('SummariesCheckedAt') WHERE id = sqlc.arg('ID');

-- name: GetSitesWithVerifiedEmail :many
-- The live sites whose admin verified an email, the ones a control admin can send an email to
SELECT id, email_verified_at FROM sites
WHERE state = 'live' AND email_verified_at IS NOT NULL
ORDER BY email_verified_at DESC, id LIMIT 1000;

-- name: UpdateSiteEmail :exec
-- A new email is not verified, and none takes the email away
UPDATE sites SET email_enc = sqlc.narg('EmailEnc'), email_verified_at = NULL WHERE id = sqlc.arg('ID');

-- name: VerifySiteEmail :execrows
-- Only the email the link was sent to is verified, not one saved since
UPDATE sites SET email_verified_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = sqlc.arg('ID') AND email_enc = sqlc.arg('EmailEnc');
