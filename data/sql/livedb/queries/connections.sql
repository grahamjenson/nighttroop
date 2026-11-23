-------
-- CREATE
--------

-- name: OpenConnection :one
-- Opens a connection, unless its site already has Max of its kind open, or its identity IdentityMax, when there is no row
INSERT INTO connections (site_id, identity_id, kind, collection)
SELECT sqlc.arg('SiteID'), sqlc.arg('IdentityID'), sqlc.arg('Kind'), sqlc.arg('Collection')
WHERE (SELECT COUNT(*) FROM connections c WHERE c.site_id = sqlc.arg('SiteID') AND c.kind = sqlc.arg('Kind')) < CAST(sqlc.arg('Max') AS INTEGER)
AND (SELECT COUNT(*) FROM connections c WHERE c.site_id = sqlc.arg('SiteID') AND c.kind = sqlc.arg('Kind')
    AND c.identity_id = sqlc.arg('IdentityID')) < CAST(sqlc.arg('IdentityMax') AS INTEGER)
RETURNING id;

-------
-- DELETE
--------

-- name: CloseConnection :exec
DELETE FROM connections WHERE id = sqlc.arg('ID');
