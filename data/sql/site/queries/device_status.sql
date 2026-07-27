-------
-- CREATE 
--------

-- name: CreateDeviceStatus :one
INSERT INTO device_status (
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