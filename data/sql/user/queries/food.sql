-------
-- CREATE 
--------
-- name: CreateFood :one
INSERT INTO food (
    id,
    name,
    category,
    carbs,
    calories,
    raw_json_data,
    created_by_app_name,
    created_by_device_name, 
    created_by_identity_id,
    created_by_identity_name
) VALUES ( ?,  ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllFoods :many
SELECT * FROM food ORDER BY created_at DESC; 

-- name: GetFood :one
SELECT * FROM food WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
