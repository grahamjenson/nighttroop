-------
-- CREATE 
--------

-- name: CreateLoopState :one
INSERT INTO loop_states (
    id,
    type,
    battery_percentage,
    charging_state,
    status,
    insertion_time,
    timestamp,
    iob,
    cob,
    iob_basal,
    iob_treatment,
    pump_reservoir,
    raw_json_data,
    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllLoopStates :many
SELECT * FROM loop_states ORDER BY occurred_at DESC;

-- name: GetLoopState :one
SELECT * FROM loop_states WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------