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

    loop_iob,
    loop_iob_basal,
    loop_iob_treatment,
    loop_cob,
    configuration_id,
    loop_status_json,
    loop_name,
    pump_clock,
    pump_manufacturer,
    pump_model,
    pump_id,
    pump_suspended,
    pump_bolusing,
    override_json,
    pump_json,
    uploader_json,


    utc_offset_minutes,
    occurred_at,
    created_by_app_name,
    created_by_device_name,
    created_by_identity_name,
    created_by_identity_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
AND occurred_at <= COALESCE(CAST(sqlc.narg('Before') AS INTEGER), 253402300799999)
AND (sqlc.narg('Before') IS NULL OR occurred_at < sqlc.narg('Before') OR id < CAST(sqlc.narg('BeforeID') AS TEXT))
ORDER BY occurred_at DESC, id DESC
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: DeviceReportTotals :one
-- How many a device's statuses DeviceReportsForDevice pages through, and when the earliest and latest occurred, 0 when there are none
SELECT COUNT(*) AS total, CAST(COALESCE(MIN(occurred_at), 0) AS INTEGER) AS earliest_at,
    CAST(COALESCE(MAX(occurred_at), 0) AS INTEGER) AS latest_at
FROM device_reports WHERE deleted_at IS NULL AND created_by_device_name = sqlc.arg('Device');

-- name: GetDeviceReport :one
SELECT * FROM device_reports WHERE id = ? AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL);


-- name: DeviceReportsLoopBetween :many
-- The loop's IOB and COB of the reports from Start until End that have any, newest first, from occurred_at's index, to
-- summarise a day
SELECT occurred_at, loop_iob, loop_iob_basal, loop_iob_treatment, loop_cob FROM device_reports
WHERE occurred_at >= sqlc.arg('Start') AND occurred_at < sqlc.arg('End') AND deleted_at IS NULL
    AND COALESCE(loop_iob, loop_iob_basal, loop_iob_treatment, loop_cob) IS NOT NULL
ORDER BY occurred_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: DeviceReportAscendingIDs :many
-- The reports in the dates searched, the oldest first, from occurred_at's index, for a search sorted by date
SELECT id FROM device_reports
WHERE occurred_at BETWEEN MAX(COALESCE(CAST(sqlc.narg('DateGTE') AS INTEGER), 0), COALESCE(CAST(sqlc.narg('DateGT') AS INTEGER) + 1, 0), COALESCE(CAST(sqlc.narg('DateEQ') AS INTEGER), 0))
    AND MIN(COALESCE(CAST(sqlc.narg('DateLTE') AS INTEGER), 253402300799999), COALESCE(CAST(sqlc.narg('DateLT') AS INTEGER) - 1, 253402300799999), COALESCE(sqlc.narg('DateEQ'), 253402300799999))
    AND (CAST(sqlc.arg('IncludeDeleted') AS BOOL) = TRUE OR deleted_at IS NULL)
ORDER BY occurred_at LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: DeviceReportHistoryIDs :many
-- The reports changed since Since, deleted ones too, the oldest change first, then by date and id as History sorts them, from updated_at's index, for API v3
-- history
SELECT id FROM device_reports WHERE updated_at >= sqlc.arg('Since')
ORDER BY updated_at, occurred_at, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

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

-- name: RewriteDeviceReport :one
-- RewriteDeviceReport writes every column CreateDeviceReport does over the device report with the ID, as a Nightscout
-- update writes its merged document. It keeps when it was created and who by: the identity given is who updated it.
UPDATE device_reports SET
    id = sqlc.arg('ID'),
    device_battery_percent = sqlc.arg('DeviceBatteryPercent'),
    is_device_charging = sqlc.arg('IsDeviceCharging'),
    pump_state = sqlc.arg('PumpState'),
    pump_reservoir = sqlc.arg('PumpReservoir'),
    pump_battery_percent = sqlc.arg('PumpBatteryPercent'),
    pump_charging_state = sqlc.arg('PumpChargingState'),
    loop_iob = sqlc.arg('LoopIob'),
    loop_iob_basal = sqlc.arg('LoopIobBasal'),
    loop_iob_treatment = sqlc.arg('LoopIobTreatment'),
    loop_cob = sqlc.arg('LoopCob'),
    configuration_id = sqlc.arg('ConfigurationID'),
    loop_status_json = sqlc.arg('LoopStatusJson'),
    loop_name = sqlc.arg('LoopName'),
    pump_clock = sqlc.arg('PumpClock'),
    pump_manufacturer = sqlc.arg('PumpManufacturer'),
    pump_model = sqlc.arg('PumpModel'),
    pump_id = sqlc.arg('PumpID'),
    pump_suspended = sqlc.arg('PumpSuspended'),
    pump_bolusing = sqlc.arg('PumpBolusing'),
    override_json = sqlc.arg('OverrideJson'),
    pump_json = sqlc.arg('PumpJson'),
    uploader_json = sqlc.arg('UploaderJson'),
    utc_offset_minutes = sqlc.arg('UtcOffsetMinutes'),
    occurred_at = sqlc.arg('OccurredAt'),
    created_by_app_name = sqlc.arg('CreatedByAppName'),
    created_by_device_name = sqlc.arg('CreatedByDeviceName'),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = CAST(sqlc.arg('CreatedByIdentityName') AS TEXT),
    updated_by_identity_id = sqlc.arg('CreatedByIdentityID')
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

-- name: ClearDeviceReportStatuses :exec
-- A device report keeps none of its loop's status or pump as sent, as one a backfill copies, an old one's being of no
-- use; its IOB, COB and pump columns are kept
UPDATE device_reports SET loop_status_json = NULL, pump_json = NULL WHERE id = sqlc.arg('ID');

-- name: HardDeleteDeviceReport :execresult
DELETE FROM device_reports WHERE id = ?;