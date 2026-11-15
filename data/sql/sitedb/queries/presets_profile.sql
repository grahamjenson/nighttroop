-------
-- CREATE 
--------
-- name: CreateProfilePreset :one
INSERT INTO profile_presets (
    id,
    name,
    profile_store_id,
    is_default,
    
    dia,
    units,
    timezone,
    carbs_hr,
    delay,

    carb_ratios_json,
    insulin_sensitivities_json,
    basal_rates_json,
    targets_high_json,
    targets_low_json,

    loop_settings_id,
    remote_settings_id,
    entered_by,
    profile_name,
    
    occurred_at,
    utc_offset_minutes,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllProfilePresets :many
SELECT * FROM profile_presets
ORDER BY created_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


-- name: GetProfilePreset :one
SELECT * FROM profile_presets WHERE id = ?;


--------
-- UPDATE
--------

-- name: UpdateProfilePreset :one
UPDATE profile_presets SET
    name = COALESCE(sqlc.narg('Name'), name),

    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID')
RETURNING *; 

--------
-- DELETE
--------

-- name: HardDeleteProfilePreset :execresult
DELETE FROM profile_presets WHERE id = ?;



-----------------------------
--- PROFILE STORE
-----------------------------

--------
-- READ
--------
-- name: GetMostRecentProfileStore :many
-- The store with the newest preset, found through occurred_at's index rather than grouping every preset by store
SELECT *
FROM profile_presets
WHERE profile_store_id = (
    SELECT profile_store_id
    FROM profile_presets
    ORDER BY occurred_at DESC
    LIMIT 1
) ORDER BY is_default DESC;

-- name: GetProfileStore :many
SELECT * FROM profile_presets WHERE profile_store_id = ? ORDER BY is_default DESC;



-- name: ProfileStoreAscendingIDs :many
-- The stores in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date
SELECT profile_store_id FROM profile_presets
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
GROUP BY profile_store_id ORDER BY MIN(occurred_at) LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ProfileStoreHistoryIDs :many
-- The stores changed since Since, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT profile_store_id FROM profile_presets WHERE updated_at >= sqlc.arg('Since')
GROUP BY profile_store_id ORDER BY MAX(updated_at), MIN(occurred_at), profile_store_id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchProfileStores :many
WITH distinct_profile_store_ids AS (
    SELECT DISTINCT profile_store_id FROM profile_presets
    WHERE
        (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
        AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
        AND (CAST(sqlc.narg('DateEQ') AS INTEGER) IS NULL OR occurred_at = sqlc.narg('DateEQ'))
        AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR occurred_at > sqlc.narg('DateGT'))
        AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR occurred_at < sqlc.narg('DateLT'))

        AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('CreatedAtGTE'))
        AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('CreatedAtLTE'))
        AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('CreatedAtGT'))
        AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('CreatedAtLT'))

        AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR updated_at >= sqlc.narg('UpdatedGTE'))
        AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR updated_at <= sqlc.narg('UpdatedLTE'))
        AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR updated_at > sqlc.narg('UpdatedGT'))
        AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR updated_at < sqlc.narg('UpdatedLT'))

        AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR profile_store_id = sqlc.narg('IDEq'))
        AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR profile_store_id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

        AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR name = sqlc.narg('NameEq'))

        AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
        AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    ORDER BY occurred_at DESC
    LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER)
)
-- Newest first by when each store starts, as Nightscout sorts profiles by startDate
SELECT * FROM profile_presets
WHERE profile_store_id IN (SELECT profile_store_id FROM distinct_profile_store_ids)
ORDER BY occurred_at DESC, created_at DESC;


--------
-- DELETE
--------

-- name: HardDeleteProfileStore :execresult
DELETE FROM profile_presets WHERE profile_store_id = ?;
-- name: SearchSummaryDefaultProfiles :many
-- The default profiles in force from Since to Until, newest first: the last uploaded by Since and each one after
SELECT * FROM profile_presets
WHERE profile_presets.is_default = 1 AND profile_presets.occurred_at <= sqlc.arg('Until')
    AND profile_presets.occurred_at >= COALESCE((SELECT MAX(before.occurred_at) FROM profile_presets AS before WHERE before.is_default = 1 AND before.occurred_at <= sqlc.arg('Since')), 0)
ORDER BY profile_presets.occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);
