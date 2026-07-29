-------
-- CREATE 
--------
-- name: CreateEventGroup :one
INSERT INTO event_groups (
    id,
    group_type,
    glucose_reading_id,
    insulin_dose_id,
    food_intake_id,
    temp_basal_id,
    temp_target_id,
    profile_switch_id,
    device_state_id,
    loop_state_id,
    pump_state_id,
    cgm_state_id,
    activity_id,
    note_id,

    occurred_at,
    utc_offset_minutes,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllEventGroups :many
SELECT 
    eg.id as id,
    eg.group_type as group_type,

    -- Glucose Reading
    gr.reading as glucose_reading,
    gr.reading_type as glucose_reading_type,

    -- Insulin Dose
    idose.dose as insulin_dose,
    idose.dose_type as insulin_dose_type,

    -- Food Intake
    fi.name as food_name,
    fi.food_type as food_type,
    fi.carbs as food_carbs,
    fi.fat as food_fat,
    fi.protein as food_protein,
    fi.fibre as food_fibre,
    fi.calories as food_calories,

    -- Activities
    a.name as activity_name,
    a.activity_type as activity_type,
    a.duration_minutes as activity_duration_minutes,
    a.calories as activity_calories,
    a.steps as activity_steps,
    a.heart_rate as activity_heart_rate,

    -- Notes
    n.content as note_content,

    -- Temp Basals
    tb.duration_minutes as temp_basal_duration_minutes,
    tb.absolute as temp_basal_absolute,
    tb.percent as temp_basal_percent,

    -- Temp Targets
    tt.duration_minutes as temp_target_duration_minutes,
    tt.high_target as temp_target_high_target,
    tt.low_target as temp_target_low_target,

    -- Profile Switches
    ps.name as profile_switch_name,
    ps.duration_minutes as profile_switch_duration_minutes,
    ps.percent as profile_switch_percent,
    ps.carb_ratios_json as profile_switch_carb_ratios_json,
    ps.insulin_sensitivities_json as profile_switch_insulin_sensitivities_json,
    ps.basal_rates_json as profile_switch_basal_rates_json,
    ps.targets_json as profile_switch_targets_json,

    -- Pump States
    ps2.pump_state as pump_state,
    ps2.pump_timestamp as pump_timestamp,
    ps2.pump_insertion_time as pump_insertion_time,
    ps2.pump_reservoir as pump_reservoir,
    ps2.pump_battery_percent as pump_battery_percent,
    ps2.pump_charging_state as pump_charging_state,

    -- CGM States
    cs.cgm_state as cgm_state,
    cs.cgm_timestamp as cgm_timestamp,
    cs.cgm_insertion_time as cgm_insertion_time,

    -- Loop States
    ls.loop_state as loop_state,
    ls.loop_timestamp as loop_timestamp,
    ls.iob as loop_iob,
    ls.iob_basal as loop_iob_basal,
    ls.iob_treatment as loop_iob_treatment,
    ls.cob as loop_cob,

    -- Device States
    ds.device_state as device_state,
    ds.device_timestamp as device_timestamp,
    ds.device_battery_percent as device_battery_percent,
    ds.device_charging_state as device_charging_state,

    -- Default stuff
    eg.occurred_at as occurred_at,
    eg.utc_offset_minutes as utc_offset_minutes,
    eg.created_at as created_at,
    eg.updated_at as updated_at,
    eg.deleted_at as deleted_at,
    eg.raw_json_data as raw_json_data,
    eg.created_by_app_name as created_by_app_name,
    eg.created_by_device_name as created_by_device_name,
    eg.created_by_identity_name as created_by_identity_name


FROM event_groups eg
LEFT JOIN glucose_readings gr ON eg.glucose_reading_id = gr.id
LEFT JOIN insulin_doses idose ON eg.insulin_dose_id = idose.id
LEFT JOIN food_intakes fi ON eg.food_intake_id = fi.id
LEFT JOIN temp_basals tb ON eg.temp_basal_id = tb.id
LEFT JOIN temp_targets tt ON eg.temp_target_id = tt.id
LEFT JOIN profile_switches ps ON eg.profile_switch_id = ps.id
LEFT JOIN device_states ds ON eg.device_state_id = ds.id
LEFT JOIN loop_states ls ON eg.loop_state_id = ls.id
LEFT JOIN pump_states ps2 ON eg.pump_state_id = ps2.id
LEFT JOIN cgm_states cs ON eg.cgm_state_id = cs.id
LEFT JOIN activities a ON eg.activity_id = a.id
LEFT JOIN notes n ON eg.note_id = n.id
WHERE (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR eg.deleted_at IS NULL)
ORDER BY eg.occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');


-- name: GetEventGroup :one
SELECT * FROM event_groups WHERE id = ?;

-- name: SearchEventGroups :many
SELECT 
    eg.id as id,
    eg.group_type as group_type,

    -- Glucose Reading
    gr.reading as glucose_reading,
    gr.reading_type as glucose_reading_type,

    -- Insulin Dose
    idose.dose as insulin_dose,
    idose.dose_type as insulin_dose_type,

    -- Food Intake
    fi.name as food_name,
    fi.food_type as food_type,
    fi.carbs as food_carbs,
    fi.fat as food_fat,
    fi.protein as food_protein,
    fi.fibre as food_fibre,
    fi.calories as food_calories,

    -- Activities
    a.name as activity_name,
    a.activity_type as activity_type,
    a.duration_minutes as activity_duration_minutes,
    a.calories as activity_calories,
    a.steps as activity_steps,
    a.heart_rate as activity_heart_rate,

    -- Notes
    n.content as note_content,

    -- Temp Basals
    tb.duration_minutes as temp_basal_duration_minutes,
    tb.absolute as temp_basal_absolute,
    tb.percent as temp_basal_percent,

    -- Temp Targets
    tt.duration_minutes as temp_target_duration_minutes,
    tt.high_target as temp_target_high_target,
    tt.low_target as temp_target_low_target,

    -- Profile Switches
    ps.name as profile_switch_name,
    ps.duration_minutes as profile_switch_duration_minutes,
    ps.percent as profile_switch_percent,
    ps.carb_ratios_json as profile_switch_carb_ratios_json,
    ps.insulin_sensitivities_json as profile_switch_insulin_sensitivities_json,
    ps.basal_rates_json as profile_switch_basal_rates_json,
    ps.targets_json as profile_switch_targets_json,

    -- Pump States
    ps2.pump_state as pump_state,
    ps2.pump_timestamp as pump_timestamp,
    ps2.pump_insertion_time as pump_insertion_time,
    ps2.pump_reservoir as pump_reservoir,
    ps2.pump_battery_percent as pump_battery_percent,
    ps2.pump_charging_state as pump_charging_state,

    -- CGM States
    cs.cgm_state as cgm_state,
    cs.cgm_timestamp as cgm_timestamp,
    cs.cgm_insertion_time as cgm_insertion_time,

    -- Loop States
    ls.loop_state as loop_state,
    ls.loop_timestamp as loop_timestamp,
    ls.iob as loop_iob,
    ls.iob_basal as loop_iob_basal,
    ls.iob_treatment as loop_iob_treatment,
    ls.cob as loop_cob,

    -- Device States
    ds.device_state as device_state,
    ds.device_timestamp as device_timestamp,
    ds.device_battery_percent as device_battery_percent,
    ds.device_charging_state as device_charging_state,

    -- Default stuff
    eg.occurred_at as occurred_at,
    eg.utc_offset_minutes as utc_offset_minutes,
    eg.created_at as created_at,
    eg.updated_at as updated_at,
    eg.deleted_at as deleted_at,
    eg.raw_json_data as raw_json_data,
    eg.created_by_app_name as created_by_app_name,
    eg.created_by_device_name as created_by_device_name,
    eg.created_by_identity_name as created_by_identity_name


FROM event_groups eg
LEFT JOIN glucose_readings gr ON eg.glucose_reading_id = gr.id
LEFT JOIN insulin_doses idose ON eg.insulin_dose_id = idose.id
LEFT JOIN food_intakes fi ON eg.food_intake_id = fi.id
LEFT JOIN temp_basals tb ON eg.temp_basal_id = tb.id
LEFT JOIN temp_targets tt ON eg.temp_target_id = tt.id
LEFT JOIN profile_switches ps ON eg.profile_switch_id = ps.id
LEFT JOIN device_states ds ON eg.device_state_id = ds.id
LEFT JOIN loop_states ls ON eg.loop_state_id = ls.id
LEFT JOIN pump_states ps2 ON eg.pump_state_id = ps2.id
LEFT JOIN cgm_states cs ON eg.cgm_state_id = cs.id
LEFT JOIN activities a ON eg.activity_id = a.id
LEFT JOIN notes n ON eg.note_id = n.id


WHERE
    (CAST(sqlc.narg('DateGTE') AS INTEGER) IS NULL OR eg.occurred_at >= sqlc.narg('DateGTE'))
    AND (CAST(sqlc.narg('DateLTE') AS INTEGER) IS NULL OR eg.occurred_at <= sqlc.narg('DateLTE'))
    AND (CAST(sqlc.narg('DateGT') AS INTEGER) IS NULL OR eg.occurred_at > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS INTEGER) IS NULL OR eg.occurred_at < sqlc.narg('DateLT'))

    AND (CAST(sqlc.narg('CreatedAtGTE') AS INTEGER) IS NULL OR eg.created_at >= sqlc.narg('CreatedAtGTE'))
    AND (CAST(sqlc.narg('CreatedAtLTE') AS INTEGER) IS NULL OR eg.created_at <= sqlc.narg('CreatedAtLTE'))
    AND (CAST(sqlc.narg('CreatedAtGT') AS INTEGER) IS NULL OR eg.created_at > sqlc.narg('CreatedAtGT'))
    AND (CAST(sqlc.narg('CreatedAtLT') AS INTEGER) IS NULL OR eg.created_at < sqlc.narg('CreatedAtLT'))

    AND (CAST(sqlc.narg('UpdatedGTE') AS INTEGER) IS NULL OR eg.updated_at >= sqlc.narg('UpdatedGTE'))
    AND (CAST(sqlc.narg('UpdatedLTE') AS INTEGER) IS NULL OR eg.updated_at <= sqlc.narg('UpdatedLTE'))
    AND (CAST(sqlc.narg('UpdatedGT') AS INTEGER) IS NULL OR eg.updated_at > sqlc.narg('UpdatedGT'))
    AND (CAST(sqlc.narg('UpdatedLT') AS INTEGER) IS NULL OR eg.updated_at < sqlc.narg('UpdatedLT'))

    AND (CAST(sqlc.narg('TypeEq') AS TEXT) IS NULL OR eg.group_type = sqlc.narg('TypeEq'))
    AND (CAST(sqlc.narg('TypeNeq') AS TEXT) IS NULL OR eg.group_type != sqlc.narg('TypeNeq'))
    AND (CAST(sqlc.narg('TypeIn') AS TEXT) IS NULL OR eg.group_type IN (SELECT value FROM json_each(sqlc.narg('TypeIn'))))
    AND (CAST(sqlc.narg('TypeNin') AS TEXT) IS NULL OR eg.group_type NOT IN (SELECT value FROM json_each(sqlc.narg('TypeNin'))))

    AND (CAST(sqlc.narg('IDEq') AS TEXT) IS NULL OR eg.id = sqlc.narg('IDEq'))
    AND (CAST(sqlc.narg('IDIn') AS TEXT) IS NULL OR eg.id IN (SELECT value FROM json_each(sqlc.narg('IDIn'))))
    
    AND (CAST(sqlc.narg('DeviceEq') AS TEXT) IS NULL OR eg.created_by_device_name = sqlc.narg('DeviceEq'))
    AND (CAST(sqlc.narg('AppEq') AS TEXT) IS NULL OR eg.created_by_app_name = sqlc.narg('AppEq'))

    -- model specific stuff

    AND (CAST(sqlc.narg('NameEq') AS TEXT) IS NULL OR food_name = sqlc.narg('NameEq') OR activity_name = sqlc.narg('NameEq') OR profile_switch_name = sqlc.narg('NameEq'))


    -- glucose readings
    AND (CAST(sqlc.narg('SgvEQ') AS INTEGER) IS NULL OR glucose_reading = sqlc.narg('SgvEQ'))
    AND (CAST(sqlc.narg('SgvGTE') AS INTEGER) IS NULL OR glucose_reading >= sqlc.narg('SgvGTE'))
    AND (CAST(sqlc.narg('SgvGT') AS INTEGER) IS NULL OR glucose_reading > sqlc.narg('SgvGT'))
    AND (CAST(sqlc.narg('SgvLTE') AS INTEGER) IS NULL OR glucose_reading <= sqlc.narg('SgvLTE'))
    AND (CAST(sqlc.narg('SgvLT') AS INTEGER) IS NULL OR glucose_reading < sqlc.narg('SgvLT'))
    
    -- food intake

    AND (CAST(sqlc.narg('CarbsEQ') AS INTEGER) IS NULL OR food_carbs = sqlc.narg('CarbsEQ'))
    AND (CAST(sqlc.narg('CarbsGTE') AS INTEGER) IS NULL OR food_carbs >= sqlc.narg('CarbsGTE'))
    AND (CAST(sqlc.narg('CarbsGT') AS INTEGER) IS NULL OR food_carbs > sqlc.narg('CarbsGT'))
    AND (CAST(sqlc.narg('CarbsLTE') AS INTEGER) IS NULL OR food_carbs <= sqlc.narg('CarbsLTE'))
    AND (CAST(sqlc.narg('CarbsLT') AS INTEGER) IS NULL OR food_carbs < sqlc.narg('CarbsLT'))
    

    -- insulin doses
    AND (CAST(sqlc.narg('InsulinEQ') AS INTEGER) IS NULL OR insulin_dose = sqlc.narg('InsulinEQ'))
    AND (CAST(sqlc.narg('InsulinGTE') AS INTEGER) IS NULL OR insulin_dose >= sqlc.narg('InsulinGTE'))
    AND (CAST(sqlc.narg('InsulinGT') AS INTEGER) IS NULL OR insulin_dose > sqlc.narg('InsulinGT'))
    AND (CAST(sqlc.narg('InsulinLTE') AS INTEGER) IS NULL OR insulin_dose <= sqlc.narg('InsulinLTE'))
    AND (CAST(sqlc.narg('InsulinLT') AS INTEGER) IS NULL OR insulin_dose < sqlc.narg('InsulinLT'))



    -- range values
    AND (CAST(sqlc.narg('DurationEQ') AS INTEGER) IS NULL OR activity_duration_minutes = sqlc.narg('DurationEQ')) OR temp_basal_duration_minutes = sqlc.narg('DurationEQ') OR temp_target_duration_minutes = sqlc.narg('DurationEQ') OR profile_switch_duration_minutes = sqlc.narg('DurationEQ')
    AND (CAST(sqlc.narg('DurationGTE') AS INTEGER) IS NULL OR activity_duration_minutes >= sqlc.narg('DurationGTE')) OR temp_basal_duration_minutes >= sqlc.narg('DurationGTE') OR temp_target_duration_minutes >= sqlc.narg('DurationGTE') OR profile_switch_duration_minutes >= sqlc.narg('DurationGTE')
    AND (CAST(sqlc.narg('DurationGT') AS INTEGER) IS NULL OR activity_duration_minutes > sqlc.narg('DurationGT')) OR temp_basal_duration_minutes > sqlc.narg('DurationGT') OR temp_target_duration_minutes > sqlc.narg('DurationGT') OR profile_switch_duration_minutes > sqlc.narg('DurationGT')
    AND (CAST(sqlc.narg('DurationLTE') AS INTEGER) IS NULL OR activity_duration_minutes <= sqlc.narg('DurationLTE')) OR temp_basal_duration_minutes <= sqlc.narg('DurationLTE') OR temp_target_duration_minutes <= sqlc.narg('DurationLTE') OR profile_switch_duration_minutes <= sqlc.narg('DurationLTE')
    AND (CAST(sqlc.narg('DurationLT') AS INTEGER) IS NULL OR activity_duration_minutes < sqlc.narg('DurationLT')) OR temp_basal_duration_minutes < sqlc.narg('DurationLT') OR temp_target_duration_minutes < sqlc.narg('DurationLT') OR profile_switch_duration_minutes < sqlc.narg('DurationLT')


    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS NULL OR eg.deleted_at IS NULL)
ORDER BY eg.occurred_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Skip');

--------
-- UPDATE
--------

-- name: UpdateEventGroup :one
UPDATE event_groups SET
    group_type = COALESCE(sqlc.narg('GroupType'), group_type),
    glucose_reading_id = COALESCE(sqlc.narg('GlucoseReadingID'), glucose_reading_id),
    insulin_dose_id = COALESCE(sqlc.narg('InsulinDoseID'), insulin_dose_id),
    food_intake_id = COALESCE(sqlc.narg('FoodIntakeID'), food_intake_id),
    temp_basal_id = COALESCE(sqlc.narg('TempBasalID'), temp_basal_id),
    temp_target_id = COALESCE(sqlc.narg('TempTargetID'), temp_target_id),
    profile_switch_id = COALESCE(sqlc.narg('ProfileSwitchID'), profile_switch_id),
    device_state_id = COALESCE(sqlc.narg('DeviceStateID'), device_state_id),
    loop_state_id = COALESCE(sqlc.narg('LoopStateID'), loop_state_id),
    pump_state_id = COALESCE(sqlc.narg('PumpStateID'), pump_state_id),
    cgm_state_id = COALESCE(sqlc.narg('CgmStateID'), cgm_state_id),
    activity_id = COALESCE(sqlc.narg('ActivityID'), activity_id),
    note_id = COALESCE(sqlc.narg('NoteID'), note_id),

    occurred_at = COALESCE(sqlc.narg('OccurredAt'), occurred_at),
    utc_offset_minutes = COALESCE(sqlc.narg('UtcOffsetMinutes'), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ? AND deleted_at IS NULL
RETURNING *; 

--------
-- DELETE
--------

-- name: SoftDeleteEventGroup :one
UPDATE event_groups SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = ?
WHERE id = ? AND deleted_at IS NULL
RETURNING *;

-- name: HardDeleteEventGroup :one
DELETE FROM event_groups WHERE id = ? RETURNING *;