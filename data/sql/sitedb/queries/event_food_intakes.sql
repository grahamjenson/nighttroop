-------
-- CREATE 
--------
-- name: CreateFoodIntake :one
INSERT INTO food_intakes (
    id,
    name,
    food_type,

    absorption_time_minutes,
    carbs,
    other_attributes,

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
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllFoodIntakes :many
SELECT * FROM food_intakes
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


-- name: GetFoodIntake :one
SELECT * FROM food_intakes WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);

-- name: SearchFoodIntakes :many
SELECT * FROM food_intakes
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

    AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR name = sqlc.narg('NameEq'))

    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR created_by_app_name = sqlc.narg('AppEq'))
    
    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR food_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR food_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR food_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR food_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('CarbsEQ') AS INTEGER) IS NULL OR carbs = sqlc.narg('CarbsEQ'))
    AND (CAST(sqlc.narg('CarbsGTE') AS INTEGER) IS NULL OR carbs >= sqlc.narg('CarbsGTE'))
    AND (CAST(sqlc.narg('CarbsGT') AS INTEGER) IS NULL OR carbs > sqlc.narg('CarbsGT'))
    AND (CAST(sqlc.narg('CarbsLTE') AS INTEGER) IS NULL OR carbs <= sqlc.narg('CarbsLTE'))
    AND (CAST(sqlc.narg('CarbsLT') AS INTEGER) IS NULL OR carbs < sqlc.narg('CarbsLT'))

    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: UpdateFoodIntake :one
UPDATE food_intakes SET
    name = COALESCE(sqlc.narg('Name'), name),
    food_type = COALESCE(sqlc.narg('FoodType'), food_type),

    carbs = COALESCE(sqlc.narg('Carbs'), carbs),
    absorption_time_minutes = COALESCE(sqlc.narg('AbsorptionTimeMinutes'), absorption_time_minutes),
    -- The fields sent, as JSON (types.OtherJSON), are merged into those kept, so one not sent stays as it is, as a
    -- column of its own would
    other_attributes = NULLIF(json_patch(
        COALESCE(other_attributes, '{}'),
        COALESCE(CAST(sqlc.narg('OtherAttributes') AS TEXT), '{}')
    ), '{}'),

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

-- name: SoftDeleteFoodIntake :execresult
UPDATE food_intakes SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteFoodIntake :execresult
DELETE FROM food_intakes WHERE id = ?;


-- name: SoftDeleteFoodIntakesByType :execresult
UPDATE food_intakes SET 
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = ?
WHERE food_type = ? AND deleted_at IS NULL;

-- name: HardDeleteFoodIntakesByType :execresult
DELETE FROM food_intakes WHERE food_type = ?;