-- A level's settings are saved whole, from a settings.SettingMap, which holds only known settings with valid values

-- name: UpdateSiteSettings :exec
UPDATE sites SET settings_json = sqlc.arg('SettingsJson') WHERE id = sqlc.arg('ID');

-- name: UpdateSiteIdentitySettings :exec
UPDATE site_identities SET settings_json = sqlc.arg('SettingsJson') WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');

-- A save from a page or the API changes only the settings it sets: Patch is a JSON merge patch, each changed setting
-- by its identifier and null for one cleared, so two saves at once keep each other's settings. A patch that is not
-- JSON changes nothing.

-- name: PatchSiteSettings :exec
UPDATE sites SET settings_json = json_patch(settings_json, CASE WHEN json_valid(sqlc.arg('Patch')) THEN CAST(sqlc.arg('Patch') AS TEXT) ELSE '{}' END) WHERE id = sqlc.arg('ID');

-- name: PatchSiteIdentitySettings :exec
UPDATE site_identities SET settings_json = json_patch(settings_json, CASE WHEN json_valid(sqlc.arg('Patch')) THEN CAST(sqlc.arg('Patch') AS TEXT) ELSE '{}' END) WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');
