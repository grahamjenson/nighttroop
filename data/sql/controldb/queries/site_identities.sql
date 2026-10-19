-------
-- CREATE 
--------

-- name: CreateSiteIdentity :exec
INSERT INTO site_identities (
    site_id,
    name,
    password_hash,
    totp_secret_enc,
    role,
    created_by_identity_id,
    created_by_identity_name
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
 );




--------
-- READ
--------


-- name: AllSiteIdentities :many
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND deleted_at IS NULL ORDER BY created_at DESC;


-- name: GetSiteIdentity :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND name = sqlc.arg('Name') AND deleted_at IS NULL LIMIT 1;


-- name: GetSiteIdentityByID :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID') AND deleted_at IS NULL LIMIT 1;


--------
-- UPDATE
--------

-- name: UpdateSiteIdentityDateFlags :exec
UPDATE site_identities
SET
    last_seen_at = COALESCE(?,
     last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE id = ?;



-- name: UpdateSiteIdentityEnabled :exec
UPDATE site_identities
SET
    enabled = sqlc.arg('Enabled'),
    updated_by_identity_id = sqlc.arg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID') AND deleted_at IS NULL;


--------
-- DELETE
--------

-- Soft delete keeps the row for the audit trail, but wipes the credentials so it can never log in again
-- name: DeleteSiteIdentity :exec
UPDATE site_identities
SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    enabled = 0,
    password_hash = '',
    totp_secret_enc = '',
    updated_by_identity_id = sqlc.arg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID') AND deleted_at IS NULL;