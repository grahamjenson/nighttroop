-------
-- CREATE 
--------

-- name: CreateEventGroupStatus :one
INSERT INTO event_groups (
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

-- name: AllEventGroupStatus :many
SELECT * FROM event_groups ORDER BY occurred_at DESC;

-- name: GetEventGroupStatus :one
SELECT * FROM event_groups WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------