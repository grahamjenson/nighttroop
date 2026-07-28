-------
-- CREATE 
--------

-- name: CreatePumpState :one
INSERT INTO pump_states (
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

-- name: AllPumpStates :many
SELECT * FROM pump_states ORDER BY occurred_at DESC;

-- name: GetPumpState :one
SELECT * FROM pump_states WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------