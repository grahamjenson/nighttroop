-------
-- CREATE 
--------
-- name: CreateActivity :exec
INSERT INTO activity (
        eid,
        occurred_at,
        name,
        duration,
        calories,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?);

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
