-- name: AllEvents :many
SELECT * FROM events ORDER BY start_at DESC;

-- name: CreateEvent :exec
INSERT INTO events (type, details, start_at, end_at) VALUES (?, ?, ?, ?);





