-- name: getSiteSettingFromDB :one
SELECT * FROM site_settings WHERE subdomain = ? AND key = ?;

-- name: upsertSiteSettingToDB :one
INSERT INTO site_settings(subdomain, key, value, created_at)
VALUES (?, ?, json(?), CAST(unixepoch('subsec') * 1000 AS INTEGER))
ON CONFLICT(subdomain, key) DO UPDATE SET
    value = excluded.value,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;