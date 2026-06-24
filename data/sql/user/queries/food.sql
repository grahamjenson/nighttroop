-------
-- CREATE 
--------
-- name: CreateFood :exec
INSERT INTO food (
        id,
        name,
        carbs,
        calories,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?,  ?, ?, ?, ?, ?, ?);

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
