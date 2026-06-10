
-------
-- CREATE 
--------

-- name: CreateSGVIgnoreEIDConstraint :execresult
INSERT INTO sgvs (
    eid,
    occurred_at,
    value,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(eid) DO NOTHING;

-- name: CreateSGV :execresult
INSERT INTO sgvs (
    eid,
    occurred_at,
    value,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: GetSGVByEID :one
SELECT * FROM sgvs WHERE eid = ?;

-- name: AllSGVs :many
SELECT * FROM sgvs ORDER BY occurred_at DESC;

--------
-- UPDATE
--------

-- name: UpdateSGVByEID :one
UPDATE sgvs SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    value = COALESCE(sqlc.narg(value), value),
    raw_json_data = COALESCE(sqlc.narg(raw_json_data), raw_json_data),
    created_by_identity_id = COALESCE(sqlc.narg(created_by_identity_id), created_by_identity_id),
    created_by_identity_name = COALESCE(sqlc.narg(created_by_identity_name), created_by_identity_name),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE eid = ?
RETURNING *; 

--------
-- DELETE
--------

-- name: DeleteSGVByEID :execresult
DELETE FROM sgvs WHERE eid = ?;

-- name: DeleteMatchingSGVs :execresult
DELETE FROM sgvs
WHERE
    (? IS NULL OR occurred_at >= ?) AND
    (? IS NULL OR occurred_at <= ?);