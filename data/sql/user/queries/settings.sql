-- name: getSettingFromDB :one
SELECT value
FROM settings
WHERE key = ?;

-- name: upsertSettingToDB :exec
INSERT INTO settings(key, value)
VALUES (?, json(?))
ON CONFLICT(key) DO UPDATE SET
    value = excluded.value;