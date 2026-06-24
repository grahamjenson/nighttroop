-------
-- CREATE 
--------
-- name: CreateCarb :exec
INSERT INTO carbs (
    id,
    occurred_at,
    utc_offset_minutes,
    amount,
    type,
    raw_json_data,
    created_by_identity_name,
    created_by_identity_id)
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: UpdateCarb :one
UPDATE carbs SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    amount = COALESCE(sqlc.narg(amount), amount),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------
