-------
-- CREATE
--------

-- name: SetNightscoutCursor :exec
INSERT INTO nightscout_cursors (site_id, collection, source, last_modified)
VALUES (sqlc.arg('SiteID'), sqlc.arg('Collection'), sqlc.arg('Source'), sqlc.arg('LastModified'))
ON CONFLICT (site_id, collection) DO UPDATE SET source = excluded.source, last_modified = excluded.last_modified;

-------
-- READ
--------

-- name: NightscoutCursors :many
-- The collections a site's sync last read from the same Nightscout
SELECT * FROM nightscout_cursors WHERE site_id = sqlc.arg('SiteID') AND source = sqlc.arg('Source');
