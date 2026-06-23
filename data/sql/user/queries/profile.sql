-------
-- CREATE 
--------

-- name: CreateProfile :exec
INSERT INTO profile (
    eid,
    name,
    START,
    duration,
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
    created_by_identity_name,
    created_by_identity_id
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllProfiles :many
SELECT * FROM profile ORDER BY occurred_at DESC;


-- name: GetProfileByEID :one
SELECT * FROM profile WHERE eid = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
