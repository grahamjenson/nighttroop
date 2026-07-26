-------
-- CREATE 
--------
-- name: CreateCarb :one
INSERT INTO carbs (
    id,
    occurred_at,
    utc_offset_minutes,
    amount,
    type,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllCarbs :many
SELECT * FROM carbs 
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

-- name: GetCarb :one
SELECT * FROM carbs WHERE id = ?;

-- name: SearchWithDefaultSortCarbs :many
SELECT * FROM carbs
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR occurred_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR occurred_at < sqlc.narg('DateLT'))

    AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('CreatedAtGTE'))
    AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('CreatedAtLTE'))
    AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('CreatedAtGT'))
    AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('CreatedAtLT'))

    AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR updated_at >= sqlc.narg('UpdatedGTE'))
    AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR updated_at <= sqlc.narg('UpdatedLTE'))
    AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR updated_at > sqlc.narg('UpdatedGT'))
    AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR updated_at < sqlc.narg('UpdatedLT'))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))

    AND (CAST(sqlc.narg('CarbsEQ') AS INTEGER) IS NULL OR amount = sqlc.narg('CarbsEQ'))
    AND (CAST(sqlc.narg('CarbsGTE') AS INTEGER) IS NULL OR amount >= sqlc.narg('CarbsGTE'))
    AND (CAST(sqlc.narg('CarbsGT') AS INTEGER) IS NULL OR amount > sqlc.narg('CarbsGT'))
    AND (CAST(sqlc.narg('CarbsLTE') AS INTEGER) IS NULL OR amount <= sqlc.narg('CarbsLTE'))
    AND (CAST(sqlc.narg('CarbsLT') AS INTEGER) IS NULL OR amount < sqlc.narg('CarbsLT'))
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(skip);
--------
-- UPDATE
--------

-- name: UpdateCarb :one
UPDATE carbs SET
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
