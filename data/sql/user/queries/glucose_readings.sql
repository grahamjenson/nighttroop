
-------
-- CREATE 
--------

-- name: CreateGlucoseReadingIgnoreEIDConstraint :execresult
INSERT INTO glucose_readings (
    eid,
    occurred_at,
    value,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(eid) DO NOTHING;

-- name: CreateGlucoseReading :execresult
INSERT INTO glucose_readings (
    eid,
    occurred_at,
    value,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: GetGlucoseReadingByEID :one
SELECT * FROM glucose_readings WHERE eid = ?;

-- name: AllGlucoseReadings :many
SELECT * FROM glucose_readings ORDER BY occurred_at DESC;

--------
-- UPDATE
--------

-- name: UpdateGlucoseReadingByEID :one
UPDATE glucose_readings SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    value = COALESCE(sqlc.narg(value), value),
    raw_json_data = COALESCE(sqlc.narg(raw_json_data), raw_json_data),
    created_by_identity_id = COALESCE(sqlc.narg(created_by_identity_id), created_by_identity_id),
    created_by_identity_name = COALESCE(sqlc.narg(created_by_identity_name), created_by_identity_name),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE eid = ?
RETURNING *; 

--------
-- DELETE
--------

-- name: DeleteGlucoseReadingByEID :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE eid = ?;

-- name: DeletePermanentGlucoseReadingByEID :execresult
DELETE FROM glucose_readings WHERE eid = ?;

-- name: DeleteGlucoseReadingsByEIDs :execresult
UPDATE glucose_readings SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE eid IN (sqlc.slice('eids'));

-- name: DeletePermanentGlucoseReadingsByEIDs :execresult
DELETE FROM glucose_readings
WHERE eid IN (sqlc.slice('eids'));