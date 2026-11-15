-------
-- CREATE 
--------
-- name: CreateTempBasal :one
INSERT INTO temp_basals (
    id,
    duration_minutes,
    rate,
    multiplier,
    type,
    insulin_type,
    automatic,
    notes,
    reason,

    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllTempBasals :many
SELECT * FROM temp_basals
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
AND occurred_at <= COALESCE(CAST(sqlc.narg('Before') AS INTEGER), 253402300799999)
AND (sqlc.narg('Before') IS NULL OR occurred_at < sqlc.narg('Before') OR id < CAST(sqlc.narg('BeforeID') AS TEXT))
ORDER BY occurred_at DESC, id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: TempBasalTotals :one
-- How many temp basals AllTempBasals pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM temp_basals WHERE deleted_at IS NULL;


-- name: GetTempBasal :one
SELECT * FROM temp_basals WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);

-- name: SearchTempBasals :many
SELECT * FROM temp_basals
WHERE
    -- The date filters as one range, occurred_at's CHECK where unset, so SQLite reads only that range of its index
    occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
        AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))

    AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('CreatedAtGTE'))
    AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('CreatedAtLTE'))
    AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('CreatedAtGT'))
    AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('CreatedAtLT'))

    AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR updated_at >= sqlc.narg('UpdatedGTE'))
    AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR updated_at <= sqlc.narg('UpdatedLTE'))
    AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR updated_at > sqlc.narg('UpdatedGT'))
    AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR updated_at < sqlc.narg('UpdatedLT'))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))

    AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR duration_minutes = sqlc.narg('DurationEQ'))
    AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR duration_minutes >= sqlc.narg('DurationGTE'))
    AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR duration_minutes > sqlc.narg('DurationGT'))
    AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR duration_minutes <= sqlc.narg('DurationLTE'))
    AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR duration_minutes < sqlc.narg('DurationLT'))

    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: RewriteTempBasal :one
-- RewriteTempBasal writes every column CreateTempBasal does over the temp basal with the ID, as a Nightscout update
-- writes its merged document. It keeps when it was created and who by: the identity given is who updated it.
UPDATE temp_basals SET
    id = sqlc.arg('ID'),
    duration_minutes = sqlc.arg('DurationMinutes'),
    rate = sqlc.arg('Rate'),
    multiplier = sqlc.arg('Multiplier'),
    type = sqlc.arg('Type'),
    insulin_type = sqlc.arg('InsulinType'),
    automatic = sqlc.arg('Automatic'),
    notes = sqlc.arg('Notes'),
    reason = sqlc.arg('Reason'),
    utc_offset_minutes = sqlc.arg('UtcOffsetMinutes'),
    occurred_at = sqlc.arg('OccurredAt'),
    created_by_app_name = sqlc.arg('CreatedByAppName'),
    created_by_device_name = sqlc.arg('CreatedByDeviceName'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = CAST(sqlc.arg('CreatedByIdentityName') AS TEXT),
    updated_by_identity_id = sqlc.arg('CreatedByIdentityID')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTempBasal :one
UPDATE temp_basals SET
    duration_minutes = COALESCE(sqlc.narg('DurationMinutes'), duration_minutes),
    rate = COALESCE(sqlc.narg('Rate'), rate),
    multiplier = COALESCE(sqlc.narg('Multiplier'), multiplier),
    type = COALESCE(sqlc.narg('Type'), type),
    insulin_type = COALESCE(NULLIF(CAST(sqlc.narg('InsulinType') AS TEXT), ''), insulin_type),
    automatic = COALESCE(sqlc.narg('Automatic'), automatic),

    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),

    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteTempBasal :execresult
UPDATE temp_basals SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteTempBasal :execresult
DELETE FROM temp_basals WHERE id = ?;