-------
-- CREATE 
--------
-- name: CreateFoodPreset :one
INSERT INTO food_presets (
    id,
    name,
    food_type,
    category,
    subcategory,

    
    absorption_time_minutes,
    pre_bolus,
    portions,
    carbs,
    fat,
    protein,
    fibre,
    energy,

    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllFoodPresets :many
SELECT * FROM food_presets
ORDER BY created_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


-- name: GetFoodPreset :one
SELECT * FROM food_presets WHERE id = ?;



-- name: FoodPresetAscendingIDs :many
-- The foods in the dates searched, the oldest first, from created_at's index, for a search sorted by date
SELECT id FROM food_presets
WHERE created_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
ORDER BY created_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: FoodPresetHistoryIDs :many
-- The foods changed since Since, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT id FROM food_presets WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at, created_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchFoodPresets :many
SELECT * FROM food_presets
WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR created_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR created_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateEQ') AS INTEGER) IS NULL OR created_at = sqlc.narg('DateEQ'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR created_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR created_at < sqlc.narg('DateLT'))

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
ORDER BY created_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

--------
-- UPDATE
--------

-- name: UpdateFoodPreset :one
UPDATE food_presets SET
    name = COALESCE(sqlc.narg('Name'), name),
    food_type = COALESCE(sqlc.narg('FoodType'), food_type),
    category = COALESCE(sqlc.narg('Category'), category),
    subcategory = COALESCE(sqlc.narg('Subcategory'), subcategory),

    absorption_time_minutes = COALESCE(sqlc.narg('AbsorptionTimeMinutes'), absorption_time_minutes),
    pre_bolus = COALESCE(sqlc.narg('PreBolus'), pre_bolus),
    portions = COALESCE(sqlc.narg('Portions'), portions),
    carbs = COALESCE(sqlc.narg('Carbs'), carbs),
    fat = COALESCE(sqlc.narg('Fat'), fat),
    protein = COALESCE(sqlc.narg('Protein'), protein),
    fibre = COALESCE(sqlc.narg('Fibre'), fibre),
    energy = COALESCE(sqlc.narg('Energy'), energy),

    raw_json_data = json_patch(
        COALESCE(raw_json_data, '{}'),
        COALESCE(CAST(sqlc.narg('RawJsonData') AS TEXT), '{}')
    ),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_id = sqlc.narg('UpdatedByIdentityID'),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID')
RETURNING *; 

--------
-- DELETE
--------


-- name: HardDeleteFoodPreset :execresult
DELETE FROM food_presets WHERE id = ?;

-- name: HardDeleteFoodPresetsByType :execresult
DELETE FROM food_presets WHERE food_type = ?;