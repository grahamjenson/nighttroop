-------
-- CREATE 
--------
-- name: CreateFood :exec
INSERT INTO food (
        eid,
        occurred_at,
        name,
        carbs,
        calories,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllFoods :many
SELECT * FROM food ORDER BY occurred_at DESC; 

-- name: GetFoodByEID :one
SELECT * FROM food WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
