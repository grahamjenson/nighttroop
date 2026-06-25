-------
-- CREATE 
--------

-- name: CreateDeviceStatus :one
INSERT INTO device_status (
    id,
    identity_id,
    battery_level,
    charging_state,
    pump_battery,
    pump_status,
    pump_reservoir,
    pump_insertion_time,
    loop_status,
    loop_timestamp,
    cgm_status,
    cgm_insertion_time,
    raw_json_data,
    utc_offset_minutes,
    occurred_at,
    created_by_identity_name,
    created_by_identity_id
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllDeviceStatus :many
SELECT * FROM device_status ORDER BY occurred_at DESC;

-- name: GetDeviceStatus :one
SELECT * FROM device_status WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------