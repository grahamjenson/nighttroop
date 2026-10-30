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
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') ORDER BY created_at DESC, id DESC; -- id breaks ties, as identities made together can share a millisecond

-- name: CountSiteIdentities :one
SELECT COUNT(*) FROM site_identities WHERE site_id = sqlc.arg('SiteID');


-- name: GetSiteIdentity :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND name = sqlc.arg('Name') LIMIT 1;


-- name: GetSiteIdentityByID :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID') LIMIT 1;


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
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');


-- name: UpdateSiteIdentityPassword :execresult
UPDATE site_identities
SET
    password_hash = sqlc.arg('PasswordHash'),
    updated_by_identity_id = NULL,
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');

-- name: UpdateSiteIdentityTOTP :execresult
UPDATE site_identities
SET
    totp_secret_enc = sqlc.arg('TotpSecretEnc'),
    updated_by_identity_id = NULL,
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');


--------
-- DELETE
--------

-- Hard delete: the row and its credentials are gone, so the name can be used again
-- name: DeleteSiteIdentity :exec
DELETE FROM site_identities
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');