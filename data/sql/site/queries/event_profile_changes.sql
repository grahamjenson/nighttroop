-------
-- CREATE 
--------

-- name: CreateProfileChange :one
INSERT INTO profile_changes (
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

-- name: AllProfileChanges :many
SELECT * FROM profile_changes ORDER BY start DESC;


-- name: GetProfileChange :one
SELECT * FROM profile_changes WHERE id = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
