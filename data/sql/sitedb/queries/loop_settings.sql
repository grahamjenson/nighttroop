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

-- name: GetPreviousLoopSettings :one
-- The set of the same app last seen before this one, the next of the app's down the Loop Settings page, which its
-- changes are shown from
SELECT * FROM loop_settings
WHERE app = sqlc.arg('App') AND (last_seen_at < sqlc.arg('LastSeenAt') OR (last_seen_at = sqlc.arg('LastSeenAt') AND id > sqlc.arg('ID')))
ORDER BY last_seen_at DESC, id
LIMIT 1;

-- name: DeleteUnreferencedLoopSettings :execrows
-- Deletes the sets no profile or device report refers to any more, as when the reports sent with them were thinned,
-- but each app's last seen, its settings now, which AAPS sends only now and then
DELETE FROM loop_settings WHERE id NOT IN (
    SELECT loop_settings_id FROM profile_presets WHERE loop_settings_id IS NOT NULL
    UNION SELECT remote_settings_id FROM profile_presets WHERE remote_settings_id IS NOT NULL
    UNION SELECT configuration_id FROM device_reports WHERE configuration_id IS NOT NULL
) AND last_seen_at < (SELECT MAX(newest.last_seen_at) FROM loop_settings newest WHERE newest.app = loop_settings.app);
