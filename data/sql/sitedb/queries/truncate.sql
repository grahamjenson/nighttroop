-------
-- TRUNCATE
-- Clearing the raw JSON kept of what each row was sent, the largest part of a site's DB, keeps all its data
-- Only rows from before @before are cleared, by when they occurred, or for food presets when they were last updated
--------

-- name: ClearActivitiesRawJSON :exec
UPDATE activities SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearComponentEventsRawJSON :exec
UPDATE component_events SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearDeviceReportsRawJSON :exec
UPDATE device_reports SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ThinDeviceReportsLoopStatus :exec
-- A loop's status, most of a device report and of a site's DB, keeps its debug output (OpenAPS's consoleLog and
-- consoleError) and predicted glucoses (OpenAPS's predBGs, Loop's predicted) only while the Nightscout web app's retro
-- view can show them; its IOB, COB, reason and the rest are kept
UPDATE device_reports SET loop_status_json = json_remove(loop_status_json,
    '$.suggested.consoleLog', '$.suggested.consoleError', '$.suggested.predBGs',
    '$.enacted.consoleLog', '$.enacted.consoleError', '$.enacted.predBGs', '$.predicted')
WHERE occurred_at < @before AND (
    json_type(loop_status_json, '$.suggested.consoleLog') IS NOT NULL
    OR json_type(loop_status_json, '$.suggested.consoleError') IS NOT NULL
    OR json_type(loop_status_json, '$.suggested.predBGs') IS NOT NULL
    OR json_type(loop_status_json, '$.enacted.consoleLog') IS NOT NULL
    OR json_type(loop_status_json, '$.enacted.consoleError') IS NOT NULL
    OR json_type(loop_status_json, '$.enacted.predBGs') IS NOT NULL
    OR json_type(loop_status_json, '$.predicted') IS NOT NULL);

-- name: ClearEventGroupsRawJSON :exec
UPDATE event_groups SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearFoodIntakesRawJSON :exec
UPDATE food_intakes SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearFoodPresetsRawJSON :exec
UPDATE food_presets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND updated_at < @before;

-- name: ClearGlucoseReadingsRawJSON :exec
UPDATE glucose_readings SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearInsulinDosesRawJSON :exec
UPDATE insulin_doses SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearNotificationEventsRawJSON :exec
UPDATE notification_events SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearProfilePresetsRawJSON :exec
UPDATE profile_presets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearProfileSwitchesRawJSON :exec
UPDATE profile_switches SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearTempBasalsRawJSON :exec
UPDATE temp_basals SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;

-- name: ClearTempTargetsRawJSON :exec
UPDATE temp_targets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL AND occurred_at < @before;
