-------
-- CREATE 
--------
-- name: CreateActivity :exec
INSERT INTO activity (
        eid,
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

-- name: GetActivityByEID :one
SELECT * FROM activity WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
