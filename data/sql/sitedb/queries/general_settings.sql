-- name: UpsertGeneralSetting :one
INSERT INTO general_settings (key, type, value_json, created_by_identity_name) VALUES (?, ?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
    type = excluded.type,
    value_json = excluded.value_json,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: GetGeneralSetting :one
SELECT * FROM general_settings WHERE key = ? AND type != 'settings';

-- name: SearchGeneralSettings :many
SELECT * FROM general_settings WHERE type != 'settings' AND updated_at >= sqlc.arg('UpdatedAtGTE') ORDER BY updated_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: CountGeneralSettings :one
SELECT COUNT(*) FROM general_settings WHERE type != 'settings';

-- name: DeleteGeneralSetting :one
DELETE FROM general_settings WHERE key = ? AND type != 'settings' RETURNING *;

-- name: DeleteGeneralSettingsBefore :exec
DELETE FROM general_settings WHERE type = ? AND updated_at < ?;

-- name: GeneralSettingsBytes :one
-- The size of every document's JSON but the one at the key, which a save replaces
SELECT CAST(COALESCE(SUM(LENGTH(value_json)), 0) AS INTEGER) FROM general_settings WHERE key != ? AND type != 'settings';

-- Nighttroop's own settings, kept beside the documents above and never among them: every query above leaves out the
-- type 'settings'.

-- name: GetSettings :many
-- The site's settings and one identity's: two rows at most, each found by its primary key
SELECT key, value_json, updated_at FROM general_settings
WHERE key IN (sqlc.arg('SiteKey'), sqlc.arg('IdentityKey')) AND type = 'settings';

-- name: UpsertSettings :exec
INSERT INTO general_settings (key, type, value_json, created_by_identity_name) VALUES (sqlc.arg('Key'), 'settings', sqlc.arg('ValueJson'), sqlc.arg('By'))
ON CONFLICT(key) DO UPDATE SET
    value_json = excluded.value_json,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE type = 'settings';

-- name: DeleteSettings :exec
-- An identity's settings go with the identity
DELETE FROM general_settings WHERE key = ? AND type = 'settings';

-- name: CountSettings :one
SELECT COUNT(*) FROM general_settings WHERE type = 'settings';

-- name: AllSettings :many
-- The site's settings and every identity's
SELECT key, value_json, updated_at FROM general_settings WHERE type = 'settings';
