-------
-- TRUNCATE
-- Clearing the raw JSON kept of what each row was sent, the largest part of a site's DB, keeps all its data
--------

-- name: ClearActivitiesRawJSON :exec
UPDATE activities SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearComponentEventsRawJSON :exec
UPDATE component_events SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearDeviceReportsRawJSON :exec
UPDATE device_reports SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearEventGroupsRawJSON :exec
UPDATE event_groups SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearFoodIntakesRawJSON :exec
UPDATE food_intakes SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearFoodPresetsRawJSON :exec
UPDATE food_presets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearGlucoseReadingsRawJSON :exec
UPDATE glucose_readings SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearInsulinDosesRawJSON :exec
UPDATE insulin_doses SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearNotificationEventsRawJSON :exec
UPDATE notification_events SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearProfilePresetsRawJSON :exec
UPDATE profile_presets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearProfileSwitchesRawJSON :exec
UPDATE profile_switches SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearTempBasalsRawJSON :exec
UPDATE temp_basals SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;

-- name: ClearTempTargetsRawJSON :exec
UPDATE temp_targets SET raw_json_data = NULL WHERE raw_json_data IS NOT NULL;
