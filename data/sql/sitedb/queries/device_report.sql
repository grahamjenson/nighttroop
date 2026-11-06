-------
-- CREATE 
--------
-- name: CreateDeviceReport :one
INSERT INTO device_reports (
    id,
    device_battery_percent,
    is_device_charging,
    pump_state,
    pump_reservoir,
    pump_battery_percent,
    pump_charging_state,
    pump_clock,
    pump_manufacturer,
    pump_model,
    override_json,

    loop_name,
    configuration_settings_json,
    loop_settings_json,
    
    loop_iob,
    loop_iob_basal,
    loop_iob_treatment,
    loop_cob,

    loop_predictions_start_time,
    loop_predictions_default,
    loop_predictions_zt,

    utc_offset_minutes,
    occurred_at,
    raw_json_data,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

--------
-- READ
--------

-- name: AllDeviceReports :many
SELECT * FROM device_reports
WHERE (CAST(sqlc.arg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


-- The devices that reported since Since, for showing each device's latest reports
-- name: RecentDeviceNames :many
SELECT DISTINCT created_by_device_name FROM device_reports
WHERE deleted_at IS NULL AND occurred_at >= sqlc.arg('Since') AND created_by_device_name IS NOT NULL
ORDER BY created_by_device_name
LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: DeviceReportsForDevice :many
SELECT * FROM device_reports
WHERE deleted_at IS NULL AND created_by_device_name = sqlc.arg('Device')
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: GetDeviceReport :one
SELECT * FROM device_reports WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);


-- name: DeviceReportHistoryIDs :many
-- The reports changed since Since, deleted ones too, the oldest change first, from updated_at's index, for API v3 history
SELECT id FROM device_reports WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: SearchDeviceReports :many
SELECT * FROM device_reports
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
    
    AND (CAST(sqlc.narg('IncludeDeleted') AS BOOL) IS TRUE OR deleted_at IS NULL)
ORDER BY occurred_at DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);


--
-- This will select the latest information (assuming if you report loop information with PWD device)
-- Limit to 20 minutes search window
-- name: GetLatestDeviceReportWithPWDInformation :one
SELECT * from device_reports
WHERE 
    deleted_at IS NULL AND 
    occurred_at >= CAST(unixepoch('subsec') * 1000 AS INTEGER) - 1200000 AND
    loop_name IS NOT NULL
ORDER BY occurred_at DESC
LIMIT 1;



--------
-- UPDATE
--------

-- name: UpdateDeviceReport :one
UPDATE device_reports SET
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

-- name: SoftDeleteDeviceReport :execresult
UPDATE device_reports SET
    deleted_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_at = MAX(updated_at + 1, CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID') AND deleted_at IS NULL;

-- name: HardDeleteDeviceReport :execresult
DELETE FROM device_reports WHERE id = ?;