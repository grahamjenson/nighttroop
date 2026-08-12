-------
-- CREATE 
--------
-- name: CreateInsulinDose :one
INSERT INTO insulin_doses (
    id,
    dose,
    dose_type,

    notes,
    reason,

    utc_offset_minutes,
    occurred_at,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllInsulinDoses :many
SELECT * FROM insulin_doses
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetInsulinDose :one
SELECT * FROM insulin_doses WHERE id = ?;

-- name: SearchInsulinDoses :many
SELECT * FROM insulin_doses
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


    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    
    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR dose_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR dose_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR dose_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR dose_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('InsulinEQ') AS INTEGER) IS NULL OR dose = sqlc.narg('InsulinEQ'))
    AND (CAST(sqlc.narg('InsulinGTE') AS INTEGER) IS NULL OR dose >= sqlc.narg('InsulinGTE'))
    AND (CAST(sqlc.narg('InsulinGT') AS INTEGER) IS NULL OR dose > sqlc.narg('InsulinGT'))
    AND (CAST(sqlc.narg('InsulinLTE') AS INTEGER) IS NULL OR dose <= sqlc.narg('InsulinLTE'))
    AND (CAST(sqlc.narg('InsulinLT') AS INTEGER) IS NULL OR dose < sqlc.narg('InsulinLT'))

    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateInsulinDose :one
UPDATE insulin_doses SET
    dose_type = COALESCE(sqlc.narg('DoseType'), dose_type),
    dose = COALESCE(sqlc.narg('Dose'), dose),

    notes = COALESCE(sqlc.narg('Notes'), notes),
    reason = COALESCE(sqlc.narg('Reason'), reason),

    occurred_at = COALESCE(sqlc.narg('OccurredAt'), occurred_at),
    utc_offset_minutes = COALESCE(sqlc.narg('UtcOffsetMinutes'), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteInsulinDose :execresult
UPDATE insulin_doses SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteInsulinDose :execresult
DELETE FROM insulin_doses WHERE id = ?;