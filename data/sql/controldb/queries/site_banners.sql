-- name: CreateSiteBanner :one
INSERT INTO site_banners (message, level, site_ids_json, expires_at, created_by_identity_name)
VALUES (sqlc.arg('Message'), sqlc.arg('Level'), sqlc.arg('SiteIdsJson'), sqlc.arg('ExpiresAt'), sqlc.arg('CreatedByIdentityName'))
RETURNING *;

-- name: AllSiteBanners :many
-- Every banner, newest first. There are few: a new one is refused once there are the most a server may have.
SELECT * FROM site_banners ORDER BY id DESC LIMIT 100;

-- name: CountSiteBanners :one
SELECT COUNT(*) FROM site_banners;

-- name: DeleteSiteBanner :execrows
DELETE FROM site_banners WHERE id = sqlc.arg('ID');

-- name: DeleteExpiredSiteBanners :execrows
DELETE FROM site_banners WHERE expires_at <= sqlc.arg('Now');

-- name: ActiveSiteBanners :many
-- The banners not yet expired, oldest first, for a site's pages to show those that are its own
SELECT * FROM site_banners WHERE expires_at > sqlc.arg('Now') ORDER BY id LIMIT 100;
