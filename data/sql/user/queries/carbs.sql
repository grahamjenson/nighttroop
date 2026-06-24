-------
-- CREATE 
--------
-- name: CreateCarb :exec
INSERT INTO carbs (
    id,
    occurred_at,
    amount,
    type,
    raw_json_data,
    created_by_identity_name,
    created_by_identity_id)
VALUES ( ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllCarbs :many
SELECT * FROM carbs ORDER BY occurred_at DESC;

-- name: GetCarb :one
SELECT * FROM carbs WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
