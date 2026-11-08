-- A site identity's passkeys: its site and identity are always set, so they are cast to be given as values

-- name: CreateSiteIdentityPasskey :exec
INSERT INTO passkeys (site_id, identity_id, credential_id, credential_json)
VALUES (CAST(sqlc.arg('SiteID') AS TEXT), CAST(sqlc.arg('IdentityID') AS INTEGER), sqlc.arg('CredentialID'), sqlc.arg('CredentialJson'));

-- name: ListSiteIdentityPasskeys :many
-- An identity's passkeys, at most as many as it may add
SELECT * FROM passkeys WHERE site_id = CAST(sqlc.arg('SiteID') AS TEXT) AND identity_id = CAST(sqlc.arg('IdentityID') AS INTEGER)
ORDER BY id LIMIT 10;

-- name: GetSiteIdentityPasskey :one
SELECT * FROM passkeys WHERE site_id = CAST(sqlc.arg('SiteID') AS TEXT) AND credential_id = sqlc.arg('CredentialID');

-- name: UpdateSiteIdentityPasskeyUsed :exec
-- A sign in with the passkey, keeping its new sign count
UPDATE passkeys
SET credential_json = sqlc.arg('CredentialJson'), last_used_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE site_id = CAST(sqlc.arg('SiteID') AS TEXT) AND credential_id = sqlc.arg('CredentialID');

-- name: DeleteSiteAdminPasskeys :exec
-- A recovery's admin signs in only with its new credentials, so a lost device's passkey works no more
DELETE FROM passkeys
WHERE passkeys.site_id = CAST(sqlc.arg('SiteID') AS TEXT) AND passkeys.identity_id IN (
    SELECT site_identities.id FROM site_identities
    WHERE site_identities.site_id = sqlc.arg('SiteID') AND site_identities.name = sqlc.arg('Name') AND site_identities.role = 'admin'
);

-- name: DeleteSiteIdentityPasskey :exec
-- An identity removes one of its own passkeys
DELETE FROM passkeys
WHERE site_id = CAST(sqlc.arg('SiteID') AS TEXT) AND identity_id = CAST(sqlc.arg('IdentityID') AS INTEGER) AND id = sqlc.arg('ID');

-- A control admin's passkeys, with no site

-- name: CreateControlPasskey :exec
INSERT INTO passkeys (control_identity_id, credential_id, credential_json)
VALUES (CAST(sqlc.arg('ControlIdentityID') AS INTEGER), sqlc.arg('CredentialID'), sqlc.arg('CredentialJson'));

-- name: ListControlPasskeys :many
-- A control admin's passkeys, at most as many as it may add
SELECT * FROM passkeys WHERE control_identity_id = CAST(sqlc.arg('ControlIdentityID') AS INTEGER) ORDER BY id LIMIT 10;

-- name: GetControlPasskey :one
SELECT * FROM passkeys WHERE control_identity_id IS NOT NULL AND credential_id = sqlc.arg('CredentialID');

-- name: UpdateControlPasskeyUsed :exec
-- A sign in to control with the passkey, keeping its new sign count
UPDATE passkeys
SET credential_json = sqlc.arg('CredentialJson'), last_used_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE control_identity_id IS NOT NULL AND credential_id = sqlc.arg('CredentialID');

-- name: DeleteControlPasskey :exec
-- A control admin removes one of its own passkeys
DELETE FROM passkeys WHERE control_identity_id = CAST(sqlc.arg('ControlIdentityID') AS INTEGER) AND id = sqlc.arg('ID');
