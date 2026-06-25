-- name: getSettingFromDB :one
SELECT value
FROM settings
WHERE id = ?;

-- name: upsertSettingToDB :exec
INSERT INTO settings(id, value)
VALUES (?, json(?))
ON CONFLICT(id) DO UPDATE SET
    value = excluded.value;