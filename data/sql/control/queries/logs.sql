-- name: TruncateEvents :exec
DELETE FROM events WHERE id < (SELECT MAX(e.id) - ? FROM events as e);
