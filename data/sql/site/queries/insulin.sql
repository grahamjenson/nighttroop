-------
-- CREATE 
--------
-- name: CreateInsulin :one
INSERT INTO insulin (
        id,
        occurred_at,
        utc_offset_minutes,
        amount,
        type,
        raw_json_data, 
        created_by_app_name,
        created_by_device_name,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllInsulin :many
SELECT * FROM insulin ORDER BY occurred_at DESC; 

-- name: GetInsulin :one
SELECT * FROM insulin WHERE id = ?;

--------
-- UPDATE
--------

-- name: UpdateInsulin :one
UPDATE insulin SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    amount = COALESCE(sqlc.narg(amount), amount),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------
