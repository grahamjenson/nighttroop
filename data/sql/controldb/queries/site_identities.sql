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
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND deleted_at IS NULL ORDER BY created_at DESC, id DESC; -- id breaks ties, as identities made together can share a millisecond

-- name: CountSiteIdentities :one
SELECT COUNT(*) FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND deleted_at IS NULL;


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

-- Hard delete: the row and its credentials are gone, so the name can be used again
-- name: DeleteSiteIdentity :exec
DELETE FROM site_identities
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');