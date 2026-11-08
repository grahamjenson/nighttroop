-- name: CreateSiteIdentityPasskey :exec
INSERT INTO site_identity_passkeys (site_id, identity_id, credential_id, credential_json)
VALUES (sqlc.arg('SiteID'), sqlc.arg('IdentityID'), sqlc.arg('CredentialID'), sqlc.arg('CredentialJson'));

-- name: ListSiteIdentityPasskeys :many
-- An identity's passkeys, at most as many as it may add
SELECT * FROM site_identity_passkeys WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID')
ORDER BY id LIMIT 10;

-- name: GetSiteIdentityPasskey :one
SELECT * FROM site_identity_passkeys WHERE site_id = sqlc.arg('SiteID') AND credential_id = sqlc.arg('CredentialID');

-- name: UpdateSiteIdentityPasskeyUsed :exec
-- A sign in with the passkey, keeping its new sign count
UPDATE site_identity_passkeys
SET credential_json = sqlc.arg('CredentialJson'), last_used_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE site_id = sqlc.arg('SiteID') AND credential_id = sqlc.arg('CredentialID');

-- name: DeleteSiteAdminPasskeys :exec
-- A recovery's admin signs in only with its new credentials, so a lost device's passkey works no more
DELETE FROM site_identity_passkeys
WHERE site_identity_passkeys.site_id = sqlc.arg('SiteID') AND site_identity_passkeys.identity_id IN (
    SELECT site_identities.id FROM site_identities
    WHERE site_identities.site_id = sqlc.arg('SiteID') AND site_identities.name = sqlc.arg('Name') AND site_identities.role = 'admin'
);

-- name: DeleteSiteIdentityPasskey :exec
-- An identity removes one of its own passkeys
DELETE FROM site_identity_passkeys WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID') AND id = sqlc.arg('ID');
