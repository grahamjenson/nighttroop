-------
-- CREATE 
--------
-- name: CreateNotificationEvent :one
INSERT INTO notification_events (
    id,
    title,
    severity,
    content,
    notification_type,
    notes,
    reason,
    
    utc_offset_minutes,
    occurred_at,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllNotificationEvents :many
SELECT * FROM notification_events
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


-- name: GetNotificationEvent :one
SELECT * FROM notification_events WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);

-- name: SearchNotificationEvents :many
SELECT * FROM notification_events
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

    AND (CAST(sqlc.narg('TitleEq') AS TEXT) IS NULL OR title = sqlc.narg('TitleEq'))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: UpdateNotificationEvent :one
UPDATE notification_events SET
    title = COALESCE(sqlc.narg('Title'), title),
    severity = COALESCE(sqlc.narg('Severity'), severity),
    notification_type = COALESCE(sqlc.narg('NotificationType'), notification_type),
    content = COALESCE(sqlc.narg('Content'), content),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),
    
    raw_json_data = json_patch(
        COALESCE(raw_json_data, '{}'),
        COALESCE(CAST(sqlc.narg('RawJsonData') AS TEXT), '{}')
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteNotificationEvent :execresult
UPDATE notification_events SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteNotificationEvent :execresult
DELETE FROM notification_events WHERE id = ?;