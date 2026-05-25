-- name: AllEntries :many
SELECT * FROM sgvs ORDER BY created_at DESC;

-- name: CreateEntry :exec
INSERT INTO sgvs (eid, created_by_type, created_by_id, value) VALUES ( ?, ?, ?, ?);
