-- name: CreateLoopSettings :one
INSERT INTO loop_settings (app, source, settings_json, occurred_at, created_by_device_name)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetLoopSettingsAt :one
-- An app's settings in force at a time: its last row at or before it
SELECT * FROM loop_settings
WHERE app = sqlc.arg('App') AND occurred_at <= sqlc.arg('At')
ORDER BY occurred_at DESC, id DESC
LIMIT 1;
