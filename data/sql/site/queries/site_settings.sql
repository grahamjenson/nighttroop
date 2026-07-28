
--------
-- READ
--------

-- name: GetSiteSettings :one
SELECT * FROM site_settings WHERE id = 1;

--------
-- UPDATE
--------

-- name: UpdateTempBasal :one
UPDATE temp_basals SET
    start = COALESCE(sqlc.narg(start), start),
    duration_minutes = COALESCE(sqlc.narg(duration_minutes), duration_minutes),
    value = COALESCE(sqlc.narg(value), value),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ?
RETURNING *; 

--------
-- DELETE
--------
