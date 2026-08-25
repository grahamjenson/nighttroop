-------
-- CREATE 
--------
-- name: CreateProfileSwitch :one
INSERT INTO profile_switches (
    id,
    name,
    duration_minutes,
    percent,

    dia,
    units,
    timezone,
    timeshift,
    carb_ratios_json,
    insulin_sensitivities_json,
    basal_rates_json,
    targets_high_json,
    targets_low_json,

    notes,
    reason,

    utc_offset_minutes,
    occurred_at,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllProfileSwitches :many
SELECT * FROM profile_switches
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetProfileSwitch :one
SELECT * FROM profile_switches WHERE id = ?;

-- name: SearchProfileSwitches :many
SELECT * FROM profile_switches
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

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

    AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR name = sqlc.narg('NameEq'))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))

    AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR duration_minutes = sqlc.narg('DurationEQ'))
    AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR duration_minutes >= sqlc.narg('DurationGTE'))
    AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR duration_minutes > sqlc.narg('DurationGT'))
    AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR duration_minutes <= sqlc.narg('DurationLTE'))
    AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR duration_minutes < sqlc.narg('DurationLT'))

    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateProfileSwitch :one
UPDATE profile_switches SET
    name = COALESCE(sqlc.narg('Name'), name),
    duration_minutes = COALESCE(sqlc.narg('DurationMinutes'), duration_minutes),
    percent = COALESCE(sqlc.narg('Percent'), percent),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),

    raw_json_data = json_patch(
        COALESCE(raw_json_data, '{}'),
        COALESCE(CAST(sqlc.narg('RawJsonData') AS TEXT), '{}')
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteProfileSwitch :execresult
UPDATE profile_switches SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteProfileSwitch :execresult
DELETE FROM profile_switches WHERE id = ?;