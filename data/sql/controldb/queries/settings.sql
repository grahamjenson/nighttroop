-- A level's settings are saved whole, from a settings.SettingMap, which holds only known settings with valid values

-- name: UpdateSiteSettings :exec
UPDATE sites SET settings_json = sqlc.arg('SettingsJson') WHERE id = sqlc.arg('ID');

-- name: UpdateSiteIdentitySettings :exec
UPDATE site_identities SET settings_json = sqlc.arg('SettingsJson') WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');
