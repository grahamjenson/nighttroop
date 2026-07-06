-------
-- CREATE 
--------

-- name: CreateSiteIdentity :exec
INSERT INTO site_identities (
    subdomain,
    name,
    password,
    totp_secret,
    role,
    created_by_identity_id,
    created_by_identity_name
) VALUES (?, ?, ?, ?, ?, ?, ?);


--------
-- READ
--------


-- name: AllSiteIdentities :many
SELECT * FROM site_identities WHERE subdomain = ? ORDER BY created_at DESC;


-- name: GetSiteIdentity :one
SELECT * FROM site_identities WHERE subdomain = ? AND name = ? LIMIT 1;


--------
-- UPDATE
--------

-- name: UpdateIdentityDateFlags :exec
UPDATE site_identities
SET
    last_seen_at = COALESCE(?, last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE id = ?;

--------
-- DELETE
--------