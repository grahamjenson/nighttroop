-------
-- CREATE 
--------
-- name: CreateInsulinDose :one
INSERT INTO insulin_doses (
    id,
    dose,
    dose_type,
    split_now,
    split_ext,
    automatic,
    is_basal_insulin,
    icfg_json,

    notes,
    reason,

    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllInsulinDoses :many
SELECT * FROM insulin_doses
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: InsulinDoseTotals :one
-- How many insulin doses AllInsulinDoses pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM insulin_doses WHERE deleted_at IS NULL;


-- name: GetInsulinDose :one
SELECT * FROM insulin_doses WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);

-- name: SearchInsulinDoses :many
SELECT * FROM insulin_doses
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
    
    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR dose_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR dose_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR dose_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR dose_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('InsulinEQ') AS INTEGER) IS NULL OR dose = sqlc.narg('InsulinEQ'))
    AND (CAST(sqlc.narg('InsulinGTE') AS INTEGER) IS NULL OR dose >= sqlc.narg('InsulinGTE'))
    AND (CAST(sqlc.narg('InsulinGT') AS INTEGER) IS NULL OR dose > sqlc.narg('InsulinGT'))
    AND (CAST(sqlc.narg('InsulinLTE') AS INTEGER) IS NULL OR dose <= sqlc.narg('InsulinLTE'))
    AND (CAST(sqlc.narg('InsulinLT') AS INTEGER) IS NULL OR dose < sqlc.narg('InsulinLT'))

    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: UpdateInsulinDose :one
UPDATE insulin_doses SET
    dose_type = COALESCE(sqlc.narg('DoseType'), dose_type),
    dose = COALESCE(sqlc.narg('Dose'), dose),
    split_now = COALESCE(sqlc.narg('SplitNow'), split_now),
    split_ext = COALESCE(sqlc.narg('SplitExt'), split_ext),
    automatic = COALESCE(sqlc.narg('Automatic'), automatic),
    is_basal_insulin = COALESCE(sqlc.narg('IsBasalInsulin'), is_basal_insulin),
    icfg_json = COALESCE(sqlc.narg('IcfgJson'), icfg_json),
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

-- name: SoftDeleteInsulinDose :execresult
UPDATE insulin_doses SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteInsulinDose :execresult
DELETE FROM insulin_doses WHERE id = ?;