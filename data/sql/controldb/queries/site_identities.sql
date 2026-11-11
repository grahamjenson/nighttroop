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
    client,
    created_by_identity_id,
    created_by_identity_name,
    nightscout_token_digest
) VALUES (
    ?,
    ?,
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


-- The identity whose Nightscout token's SHA1 has this digest, see site_identities.nightscout_token_digest
-- name: GetSiteIdentityByTokenDigest :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND nightscout_token_digest = sqlc.arg('Digest') LIMIT 1;

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



-- A recovery gives the site's admin new credentials and turns it back on
-- name: ResetSiteAdminCredentials :execresult
UPDATE site_identities SET password_hash = sqlc.arg('PasswordHash'), totp_secret_enc = sqlc.arg('TotpSecretEnc'), enabled = 1, updated_by_identity_name = 'backup key',
    credentials_changed_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE site_id = sqlc.arg('SiteID') AND name = sqlc.arg('Name') AND role = 'admin';

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
    nightscout_token_digest = sqlc.arg('NightscoutTokenDigest'),
    credentials_changed_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = NULL,
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE site_id = sqlc.arg('SiteID') AND id = sqlc.arg('ID');

-- name: UpdateSiteIdentityTOTP :execresult
UPDATE site_identities
SET
    totp_secret_enc = sqlc.arg('TotpSecretEnc'),
    credentials_changed_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
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

-- name: DeleteAllSiteIdentities :exec
-- Hard deletes all a site's identities, with their credentials and push registrations, as when it is deleted for good
DELETE FROM site_identities WHERE site_id = sqlc.arg('SiteID');
-- name: CountActiveIdentities :one
-- The identities seen since Since, in milliseconds
SELECT COUNT(*) FROM site_identities WHERE last_seen_at >= sqlc.arg('Since');
