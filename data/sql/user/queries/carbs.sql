-------
-- CREATE 
--------
-- name: CreateCarb :exec
INSERT INTO carbs (
    eid,
    occurred_at,
    amount,
    raw_json_data,
    created_by_identity_name,
    created_by_identity_id)
VALUES ( ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllCarbs :many
SELECT * FROM carbs ORDER BY occurred_at DESC;

-- name: GetCarbByEID :one
SELECT * FROM carbs WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
