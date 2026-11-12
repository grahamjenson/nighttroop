-------
-- CREATE 
--------
-- name: CreateProfileSwitch :one
INSERT INTO profile_switches (
    id,
    name,
    duration_minutes,
    multiplier,

    dia,
    units,
    timezone,
    carbs_hr,
    delay,
    timeshift,
    original_profile_name,
    original_customized_name,
    original_timeshift_ms,
    original_percentage,
    original_duration_ms,
    original_end,
    icfg_json,

    carb_ratios_json,
    insulin_sensitivities_json,
    basal_rates_json,
    targets_high_json,
    targets_low_json,

    notes,
    reason,

    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (
    sqlc.arg('ID'),
    sqlc.arg('Name'),
    sqlc.arg('DurationMinutes'),
    sqlc.arg('Multiplier'),
    sqlc.arg('Dia'),
    sqlc.arg('Units'),
    sqlc.arg('Timezone'),
    sqlc.arg('CarbsHr'),
    sqlc.arg('Delay'),
    sqlc.arg('Timeshift'),
    NULLIF(CAST(sqlc.narg('OriginalProfileName') AS TEXT), ''),
    NULLIF(CAST(sqlc.narg('OriginalCustomizedName') AS TEXT), ''),
    sqlc.narg('OriginalTimeshiftMs'),
    sqlc.narg('OriginalPercentage'),
    sqlc.narg('OriginalDurationMs'),
    sqlc.narg('OriginalEnd'),
    sqlc.narg('IcfgJson'),
    sqlc.arg('CarbRatiosJson'),
    sqlc.arg('InsulinSensitivitiesJson'),
    sqlc.arg('BasalRatesJson'),
    sqlc.arg('TargetsHighJson'),
    sqlc.arg('TargetsLowJson'),
    sqlc.arg('Notes'),
    sqlc.arg('Reason'),
    sqlc.arg('UtcOffsetMinutes'),
    sqlc.arg('OccurredAt'),
    sqlc.arg('CreatedByAppName'),
    sqlc.arg('CreatedByDeviceName'),
    sqlc.arg('CreatedByIdentityName'),
    sqlc.arg('CreatedByIdentityID')
)
RETURNING *;

--------
-- READ
--------

-- name: AllProfileSwitches :many
SELECT * FROM profile_switches
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ProfileSwitchTotals :one
-- How many profile switches AllProfileSwitches pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM profile_switches WHERE deleted_at IS NULL;


-- name: GetProfileSwitch :one
SELECT * FROM profile_switches WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);

-- name: SearchProfileSwitches :many
SELECT * FROM profile_switches
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

-- name: UpdateProfileSwitch :one
UPDATE profile_switches SET
    name = COALESCE(sqlc.narg('Name'), name),
    duration_minutes = COALESCE(sqlc.narg('DurationMinutes'), duration_minutes),
    multiplier = COALESCE(sqlc.narg('Multiplier'), multiplier),
    timeshift = COALESCE(sqlc.narg('Timeshift'), timeshift),
    original_profile_name = COALESCE(NULLIF(CAST(sqlc.narg('OriginalProfileName') AS TEXT), ''), original_profile_name),
    original_customized_name = COALESCE(NULLIF(CAST(sqlc.narg('OriginalCustomizedName') AS TEXT), ''), original_customized_name),
    original_timeshift_ms = COALESCE(sqlc.narg('OriginalTimeshiftMs'), original_timeshift_ms),
    original_percentage = COALESCE(sqlc.narg('OriginalPercentage'), original_percentage),
    original_duration_ms = COALESCE(sqlc.narg('OriginalDurationMs'), original_duration_ms),
    original_end = COALESCE(sqlc.narg('OriginalEnd'), original_end),
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

-- name: SoftDeleteProfileSwitch :execresult
UPDATE profile_switches SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteProfileSwitch :execresult
DELETE FROM profile_switches WHERE id = ?;
-- name: SearchSummaryProfileSwitches :many
-- The profile switches in force from Since to Until, newest first: those since the last switch by Since that lasts
-- until the next, or since Earliest when there is none
SELECT * FROM profile_switches
WHERE profile_switches.deleted_at IS NULL AND profile_switches.occurred_at <= sqlc.arg('Until')
    AND profile_switches.occurred_at >= COALESCE((SELECT MAX(before.occurred_at) FROM profile_switches AS before WHERE before.deleted_at IS NULL AND before.duration_minutes = 0 AND before.occurred_at <= sqlc.arg('Since')), sqlc.arg('Earliest'))
ORDER BY profile_switches.occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);
