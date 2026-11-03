--------------------
-- PUSH CREDENTIALS
--------------------

-- name: CreatePushCredential :one
INSERT INTO push_credentials (
    site_id, name, provider, team_or_project_id, key_id, secret_enc, apps_json, bundle_ids_json, created_by_identity_name
) VALUES (
    sqlc.narg('SiteID'), sqlc.arg('Name'), sqlc.arg('Provider'), sqlc.arg('TeamOrProjectID'), sqlc.narg('KeyID'),
    sqlc.arg('SecretEnc'), sqlc.arg('AppsJson'), sqlc.narg('BundleIdsJson'), sqlc.arg('CreatedByIdentityName')
)
RETURNING *;

-- name: GetPushCredential :one
-- A site's credential, or with no SiteID one for every site
SELECT * FROM push_credentials
WHERE id = sqlc.arg('ID') AND site_id IS sqlc.narg('SiteID') AND deleted_at IS NULL;

-- name: SearchPushCredentials :many
-- A site's credentials, or with no SiteID those for every site, newest first
SELECT * FROM push_credentials
WHERE site_id IS sqlc.narg('SiteID') AND deleted_at IS NULL
ORDER BY id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: CountPushCredentials :one
SELECT COUNT(*) FROM push_credentials WHERE site_id IS sqlc.narg('SiteID') AND deleted_at IS NULL;

-- name: DeletePushCredential :execrows
-- Soft deletes it, removing its secret
UPDATE push_credentials SET
    secret_enc = NULL,
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = sqlc.arg('ID') AND site_id IS sqlc.narg('SiteID') AND deleted_at IS NULL;

-- name: DeleteSitePushCredentials :exec
-- Soft deletes all a site's credentials, removing their secrets, as when it is deleted
UPDATE push_credentials SET
    secret_enc = NULL,
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE site_id = sqlc.arg('SiteID') AND deleted_at IS NULL;

-- name: PushCredentialFor :one
-- The credential to push to an app on a site's device with: the site's own, else the newest for every site, that
-- pushes to the app and allows its bundle ID. Apps and bundle IDs have no quotes, so each is found in its JSON list
-- as itself in quotes.
SELECT * FROM push_credentials
WHERE (site_id = sqlc.arg('SiteID') OR site_id IS NULL) AND provider = sqlc.arg('Provider') AND deleted_at IS NULL
  AND instr(apps_json, '"' || CAST(sqlc.arg('App') AS TEXT) || '"') > 0
  AND (bundle_ids_json IS NULL OR instr(bundle_ids_json, '"' || CAST(sqlc.arg('BundleID') AS TEXT) || '"') > 0)
ORDER BY site_id IS NULL, id DESC
LIMIT 1;

----------------------
-- PUSH REGISTRATIONS
----------------------

-- name: UpsertPushRegistration :one
-- Registers a token, moving it to this identity if another had it, as the app was signed in again
INSERT INTO push_registrations (site_id, identity_id, app, provider, bundle_id, production, token_enc, token_hash)
VALUES (
    sqlc.arg('SiteID'), sqlc.arg('IdentityID'), sqlc.arg('App'), sqlc.arg('Provider'), sqlc.arg('BundleID'),
    sqlc.arg('Production'), sqlc.arg('TokenEnc'), sqlc.arg('TokenHash')
)
ON CONFLICT (token_hash) DO UPDATE SET
    site_id = excluded.site_id,
    identity_id = excluded.identity_id,
    app = excluded.app,
    provider = excluded.provider,
    bundle_id = excluded.bundle_id,
    production = excluded.production,
    token_enc = excluded.token_enc,
    last_error = NULL,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: SearchPushRegistrations :many
-- A site's registrations, or with IdentityID only that identity's, newest first
SELECT * FROM push_registrations
WHERE site_id = sqlc.arg('SiteID') AND identity_id = COALESCE(sqlc.narg('IdentityID'), identity_id)
ORDER BY id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: CountPushRegistrations :one
SELECT COUNT(*) FROM push_registrations WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID');

-- name: DeletePushRegistration :execrows
-- Deletes a site's registration, or with IdentityID only if it is that identity's
DELETE FROM push_registrations
WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID')
  AND identity_id = COALESCE(sqlc.narg('IdentityID'), identity_id);

-- name: DeletePushRegistrationByTokenHash :exec
-- Deletes a token the push service says is no longer registered
DELETE FROM push_registrations WHERE token_hash = sqlc.arg('TokenHash');

-- name: DeleteSitePushRegistrations :exec
DELETE FROM push_registrations WHERE site_id = sqlc.arg('SiteID');

-- name: RecordPush :exec
-- Records a push sent to a registration, and its error, if it had one
UPDATE push_registrations SET
    last_pushed_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    last_error = sqlc.narg('LastError')
WHERE id = sqlc.arg('ID');
