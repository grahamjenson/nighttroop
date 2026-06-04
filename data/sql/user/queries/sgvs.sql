-- name: AllEntries :many
SELECT * FROM sgvs ORDER BY created_at DESC;

-- name: CreateEntry :exec
INSERT INTO sgvs (eid, created_at, value, created_by_identity_id) VALUES (?, ?, ?, ?);

-- name: CreateEntryIgnoreEIDConstraint :execresult
INSERT INTO sgvs (
    eid,
    created_at,
    value,
    created_by_identity_id
)
VALUES (?, ?, ?, ?)
ON CONFLICT(eid) DO NOTHING;