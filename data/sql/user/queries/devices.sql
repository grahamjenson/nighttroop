-- name: AllDevices :many
SELECT * FROM devices ORDER BY created_at DESC;

-- name: CreateDevice :exec
INSERT INTO devices (name, password, totp_secret, role) VALUES (?, ?, ?, ?);

-- name: GetDevice :one
SELECT * FROM devices where name = ? LIMIT 1;

-- name: UpdateDeviceLastSeenAt :exec
UPDATE devices SET last_seen_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE name = ?;

------
-- Device Status
-------

-- name: AllDeviceStatus :many
SELECT * FROM device_status ORDER BY created_at DESC;
