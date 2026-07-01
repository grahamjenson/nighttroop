-------
-- CREATE 
--------
-- name: CreateActivity :one
INSERT INTO activity (
        id,
        occurred_at,
        utc_offset_minutes,
        name,
        type,
        notes,
        duration_minutes,
        calories,
        steps,
        heart_rate,
        raw_json_data,
        created_by_app_name,
        created_by_device_name,
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllActivities :many
SELECT * FROM activity ORDER BY occurred_at DESC; 

-- name: GetActivity :one
SELECT * FROM activity WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
