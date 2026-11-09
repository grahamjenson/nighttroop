-- name: UpsertLoopSettings :exec
-- Keeps a set of an app's settings once, however often it is sent, widening when it was seen to take in this time
INSERT INTO loop_settings (id, app, source, settings_json, first_seen_at, last_seen_at, created_by_device_name)
VALUES (sqlc.arg('ID'), sqlc.arg('App'), sqlc.arg('Source'), sqlc.arg('SettingsJson'), sqlc.arg('SeenAt'), sqlc.arg('SeenAt'), sqlc.narg('CreatedByDeviceName'))
ON CONFLICT(id) DO UPDATE SET
    first_seen_at = MIN(first_seen_at, excluded.first_seen_at),
    last_seen_at = MAX(last_seen_at, excluded.last_seen_at);

-- name: GetLoopSettings :one
SELECT * FROM loop_settings WHERE id = ?;

-- name: SearchLoopSettings :many
-- Each set of settings, the last seen first, of one app or all of them
SELECT * FROM loop_settings
WHERE CAST(sqlc.narg('App') AS TEXT) IS NULL OR app = sqlc.narg('App')
ORDER BY last_seen_at DESC, id
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: DeleteLoopSettings :execresult
DELETE FROM loop_settings WHERE id = ?;
