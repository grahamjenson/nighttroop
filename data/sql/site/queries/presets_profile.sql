-------
-- CREATE 
--------

-- name: CreateProfilePreset :one
INSERT INTO profile_presets (
    id,
    name,
    start,
    duration_minutes,
    source,
    percent,
    dia,
    units,
    timezone,
    carb_ratios_json,
    insulin_sensitivities_json,
    basal_rates_json,
    targets_json,
    raw_json_data,
    utc_offset_minutes,
    created_by_app_name,
    created_by_device_name, 
    created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllProfiles :many
SELECT * FROM profile_presets ORDER BY start DESC;


-- name: GetProfile :one
SELECT * FROM profile_presets WHERE id = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
