-------
-- CREATE 
--------
-- name: CreateActivity :exec
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
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllActivities :many
SELECT * FROM activity ORDER BY occurred_at DESC; 

-- name: GetActivityByID :one
SELECT * FROM activity WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
