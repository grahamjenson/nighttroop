-- name: AllEntries :many
SELECT * FROM entries ORDER BY created_at DESC;

-- name: CreateEntry :exec
INSERT INTO entries (eid, created_by_type, created_by_id, value) VALUES ( ?, ?, ?, ?);
