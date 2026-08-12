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
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllNotificationEvents :many
SELECT * FROM notification_events
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetNotificationEvent :one
SELECT * FROM notification_events WHERE id = ?;

-- name: SearchNotificationEvents :many
SELECT * FROM notification_events
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR occurred_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR occurred_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR occurred_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR occurred_at < sqlc.narg('DateLT'))

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
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

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
    
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteNotificationEvent :execresult
UPDATE notification_events SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteNotificationEvent :execresult
DELETE FROM notification_events WHERE id = ?;