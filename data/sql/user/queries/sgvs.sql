
-------
-- CREATE 
--------

-- name: CreateSGVIgnoreEIDConstraint :execresult
INSERT INTO sgvs (
    eid,
    created_at,
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
    created_at,
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
SELECT * FROM sgvs ORDER BY created_at DESC;

--------
-- UPDATE
--------


--------
-- DELETE
--------
