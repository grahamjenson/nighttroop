-- name: AllIdentities :many
SELECT * FROM identities ORDER BY created_at DESC;

-- name: CreateIdentity :exec
INSERT INTO identities (name, password, totp_secret, role) VALUES (?, ?, ?, ?);

-- name: GetIdentity :one
SELECT * FROM identities where name = ? LIMIT 1;

-- name: UpdateIdentityLastSeenAt :exec
UPDATE identities SET last_seen_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE name = ?;

------
-- Identity Status
-------

-- name: AllIdentityStatus :many
SELECT * FROM identity_status ORDER BY created_at DESC;
