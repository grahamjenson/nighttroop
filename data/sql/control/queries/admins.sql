-- name: CreateAdmin :one
INSERT INTO admins (name, password, role, totp_secret) VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetAdmin :one
SELECT * FROM admins WHERE name = ?;



