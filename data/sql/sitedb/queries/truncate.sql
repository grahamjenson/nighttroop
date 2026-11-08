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

-- name: ClearDeviceReportsSettings :exec
-- The loop's configuration and settings as sent, up to 8 KB each, are dropped from old device reports' other attributes;
-- the loop's name, IOB, COB and predictions are kept, and other attributes left with none is NULL, as when none came
UPDATE device_reports SET other_attributes = NULLIF(json_remove(other_attributes, '$.configuration_settings', '$.loop_settings'), '{}')
WHERE occurred_at < @before
    AND (json_type(other_attributes, '$.configuration_settings') IS NOT NULL OR json_type(other_attributes, '$.loop_settings') IS NOT NULL);

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
