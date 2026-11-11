-- name: GetServerSettings :one
SELECT * FROM server_settings WHERE id = 1;

-- name: UpdatePublicSiteCreation :exec
UPDATE server_settings SET
    public_site_creation = sqlc.arg('PublicSiteCreation'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = 1;

-- name: UpdateRateLimits :exec
UPDATE server_settings SET
    rate_limits_json = sqlc.arg('RateLimitsJson'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = 1;
