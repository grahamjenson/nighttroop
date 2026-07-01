-------
-- CREATE 
--------

-- name: CreateIdentity :exec
INSERT INTO identities (
    name,
    password,
    totp_secret,
    role,
    created_by_identity_id,
    created_by_identity_name
) VALUES (?, ?, ?, ?, ?, ?);


--------
-- READ
--------


-- name: AllIdentities :many
SELECT * FROM identities ORDER BY created_at DESC;


-- name: GetIdentity :one
SELECT * FROM identities where name = ? LIMIT 1;


--------
-- UPDATE
--------

-- name: UpdateIdentityDateFlags :exec
UPDATE identities
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE id = ?;

--------
-- DELETE
--------