-------
-- CREATE 
--------
-- name: CreateProfilePreset :one
INSERT INTO profile_presets (
    id,
    name,
    carb_ratios_json,
    insulin_sensitivities_json,
    basal_rates_json,
    targets_json,

    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllProfilePresets :many
SELECT * FROM profile_presets
ORDER BY created_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetProfilePreset :one
SELECT * FROM profile_presets WHERE id = ?;

-- name: SearchProfilePresets :many
SELECT * FROM profile_presets
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('DateLT'))

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
ORDER BY created_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateProfilePreset :one
UPDATE profile_presets SET
    name = COALESCE(sqlc.narg('Name'), name),
    carb_ratios_json = COALESCE(sqlc.narg('CarbRatiosJson'), carb_ratios_json),
    insulin_sensitivities_json = COALESCE(sqlc.narg('InsulinSensitivitiesJson'), insulin_sensitivities_json),
    basal_rates_json = COALESCE(sqlc.narg('BasalRatesJson'), basal_rates_json),
    targets_json = COALESCE(sqlc.narg('TargetsJson'), targets_json),

    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------


-- name: HardDeleteProfilePreset :execresult
DELETE FROM profile_presets WHERE id = ? RETURNING *;