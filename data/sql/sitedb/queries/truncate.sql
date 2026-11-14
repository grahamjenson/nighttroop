-- name: ClearOldLoopStatuses :exec
-- A loop's status as sent is kept whole for a day, as followers read a day of them and cache each as never changing:
-- LoopFollow its predictions, timestamp and recommended bolus, the Nightscout web app's Loop pill its last enacted and
-- OK times. Then only its bulky predictions are cleared, Loop's predicted.values and openaps' predBGs in suggested and
-- enacted; the rest stays as sent, with Loop's predicted.startDate. A status is cleared once it is a day older than Now
-- and than its device's newest status with one, so a device that stops keeps its last day whole. Its updated_at stays,
-- so clients syncing by srvModified do not fetch it again. Only reports written since Since are looked at, by their
-- updated_at index, and those that turned a day old since then, by their occurred_at index, as those before were
-- cleared by an earlier run; 0 looks at them all. A cleared status has no predictions, so it is not written again.
UPDATE device_reports
SET loop_status_json = json_remove(device_reports.loop_status_json, '$.predicted.values', '$.suggested.predBGs', '$.enacted.predBGs')
WHERE device_reports.occurred_at < CAST(@now AS INTEGER) - 86400000 AND (
        device_reports.updated_at >= CAST(@since AS INTEGER)
        OR device_reports.occurred_at BETWEEN CAST(@since AS INTEGER) - 86400000 AND CAST(@now AS INTEGER) - 86400000
    )
    AND (json_type(device_reports.loop_status_json, '$.predicted.values') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.suggested.predBGs') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.enacted.predBGs') IS NOT NULL)
    AND device_reports.occurred_at < (
        SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name
            AND newest.loop_status_json IS NOT NULL AND newest.deleted_at IS NULL
    ) - 86400000;

-- name: ClearOldPumps :exec
-- A pump as sent, the most of a report after its loop's status, is kept only on its device's newest report
-- with one, as followers read it from that one. What is worth keeping of the others, their battery, reservoir, clock,
-- manufacturer, model and ID, and whether it was suspended or bolusing, was extracted into their own columns when they
-- were written.
UPDATE device_reports SET pump_json = NULL
WHERE device_reports.updated_at >= @since AND device_reports.pump_json IS NOT NULL
    AND device_reports.occurred_at < (
        SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name
            AND newest.pump_json IS NOT NULL AND newest.deleted_at IS NULL
    );

-- name: ThinOldDeviceReports :exec
-- A device report every few minutes is kept only for the last hour before Now; then each device keeps, of its
-- reports, the newest in every 10 minutes until a week old, in every 20 minutes until two weeks old, and in every hour
-- after that, which is enough to chart its IOB, COB, battery and reservoir. Each device keeps its newest report, and the
-- loop status on it, however old, until DeleteOldDeviceReports removes it. The buckets are aligned so each hour's
-- newest is also its 20 and 10 minutes' newest. A deleted report is neither kept nor removed, so clients syncing
-- deletions still see it. Only reports written since Since are looked at, as ClearOldLoopStatuses does, and those that
-- passed a week or two weeks old since then, as those are thinned again; 0 looks at them all.
DELETE FROM device_reports WHERE device_reports.id IN (
    SELECT report.id FROM (
        SELECT thinned.id, thinned.created_by_device_name, thinned.occurred_at,
            CASE WHEN thinned.occurred_at < CAST(@now AS INTEGER) - 14 * 86400000 THEN 3600000
                WHEN thinned.occurred_at < CAST(@now AS INTEGER) - 7 * 86400000 THEN 1200000
                ELSE 600000 END AS bucket
        FROM device_reports AS thinned
        WHERE thinned.occurred_at < CAST(@now AS INTEGER) - 3600000 AND thinned.deleted_at IS NULL AND (
            thinned.updated_at >= CAST(@since AS INTEGER)
            OR thinned.occurred_at BETWEEN CAST(@since AS INTEGER) - 7 * 86400000 AND CAST(@now AS INTEGER) - 7 * 86400000
            OR thinned.occurred_at BETWEEN CAST(@since AS INTEGER) - 14 * 86400000 AND CAST(@now AS INTEGER) - 14 * 86400000
        )
    ) AS report
    WHERE EXISTS (
        SELECT 1 FROM device_reports AS kept
        WHERE kept.created_by_device_name IS report.created_by_device_name AND kept.deleted_at IS NULL
            AND kept.occurred_at >= report.occurred_at
            AND kept.occurred_at < (report.occurred_at / report.bucket + 1) * report.bucket
            AND (kept.occurred_at > report.occurred_at OR kept.id > report.id)
    )
);

-- name: DeleteOldDeviceReports :execrows
-- Every device report that occurred before Before, 30 days ago, or 7 for a site far over its size limit, is deleted, a
-- deleted one too, by the occurred_at index
DELETE FROM device_reports WHERE occurred_at < @before;

-- A site far over its DB size limit first has the JSON that matters only while recent cleared from its records that
-- occurred before Before, a day ago, but each device's newest report and the newest profile switch, which followers and
-- the current profile read. Their updated_at stays, so clients do not fetch them again.
-- name: ClearOldDeviceReportJSON :execrows
UPDATE device_reports SET loop_status_json = NULL, override_json = NULL, pump_json = NULL, uploader_json = NULL
WHERE device_reports.occurred_at < @before
    AND COALESCE(loop_status_json, override_json, pump_json, uploader_json) IS NOT NULL
    AND device_reports.occurred_at < (SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name AND newest.deleted_at IS NULL);

-- name: ClearOldBolusCalculatorJSON :execrows
UPDATE event_groups SET bolus_calculator_result_json = NULL
WHERE event_groups.occurred_at < @before AND bolus_calculator_result_json IS NOT NULL;

-- name: ClearOldInsulinDoseJSON :execrows
UPDATE insulin_doses SET icfg_json = NULL WHERE insulin_doses.occurred_at < @before AND icfg_json IS NOT NULL;

-- name: ClearOldProfileSwitchJSON :execrows
UPDATE profile_switches SET icfg_json = NULL WHERE profile_switches.occurred_at < @before AND icfg_json IS NOT NULL
    AND profile_switches.occurred_at < (SELECT MAX(newest.occurred_at) FROM profile_switches AS newest
        WHERE newest.deleted_at IS NULL);

-- If such a site is still over, its treatments and readings that occurred before Before, 90 days ago, are deleted, the
-- deleted ones too, but the newest not deleted of each group type, reading type and table, as a sensor's, site's or
-- profile's age and the current profile read the newest however old. The groups go first, so the events they refer to
-- are kept. The summaries are kept and are not marked stale.
-- name: DeleteOldEventGroups :execrows
DELETE FROM event_groups WHERE event_groups.occurred_at < @before
    AND event_groups.occurred_at < (SELECT MAX(newest.occurred_at)
    FROM event_groups AS newest WHERE newest.group_type = event_groups.group_type AND newest.deleted_at IS NULL);

-- name: DeleteOldGlucoseReadings :execrows
DELETE FROM glucose_readings WHERE glucose_readings.occurred_at < @before
    AND glucose_readings.id NOT IN (SELECT glucose_reading_id FROM event_groups WHERE glucose_reading_id IS NOT NULL)
    AND glucose_readings.occurred_at < (SELECT MAX(newest.occurred_at) FROM glucose_readings AS newest
        WHERE newest.deleted_at IS NULL
        AND newest.reading_type = glucose_readings.reading_type);

-- name: DeleteOldInsulinDoses :execrows
DELETE FROM insulin_doses WHERE insulin_doses.occurred_at < @before
    AND insulin_doses.id NOT IN (SELECT insulin_dose_id FROM event_groups WHERE insulin_dose_id IS NOT NULL)
    AND insulin_doses.occurred_at < (SELECT MAX(newest.occurred_at) FROM insulin_doses AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldFoodIntakes :execrows
DELETE FROM food_intakes WHERE food_intakes.occurred_at < @before
    AND food_intakes.id NOT IN (SELECT food_intake_id FROM event_groups WHERE food_intake_id IS NOT NULL)
    AND food_intakes.occurred_at < (SELECT MAX(newest.occurred_at) FROM food_intakes AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldTempBasals :execrows
DELETE FROM temp_basals WHERE temp_basals.occurred_at < @before
    AND temp_basals.id NOT IN (SELECT temp_basal_id FROM event_groups WHERE temp_basal_id IS NOT NULL)
    AND temp_basals.occurred_at < (SELECT MAX(newest.occurred_at) FROM temp_basals AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldTempTargets :execrows
DELETE FROM temp_targets WHERE temp_targets.occurred_at < @before
    AND temp_targets.id NOT IN (SELECT temp_target_id FROM event_groups WHERE temp_target_id IS NOT NULL)
    AND temp_targets.occurred_at < (SELECT MAX(newest.occurred_at) FROM temp_targets AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldProfileSwitches :execrows
DELETE FROM profile_switches WHERE profile_switches.occurred_at < @before
    AND profile_switches.id NOT IN (SELECT profile_switch_id FROM event_groups WHERE profile_switch_id IS NOT NULL)
    AND profile_switches.occurred_at < (SELECT MAX(newest.occurred_at) FROM profile_switches AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldActivities :execrows
DELETE FROM activities WHERE activities.occurred_at < @before
    AND activities.id NOT IN (SELECT activity_id FROM event_groups WHERE activity_id IS NOT NULL)
    AND activities.occurred_at < (SELECT MAX(newest.occurred_at) FROM activities AS newest
        WHERE newest.deleted_at IS NULL);

-- name: DeleteOldNotificationEvents :execrows
DELETE FROM notification_events WHERE notification_events.occurred_at < @before
    AND notification_events.id NOT IN (SELECT notification_event_id FROM event_groups WHERE notification_event_id IS NOT NULL)
    AND notification_events.occurred_at < (SELECT MAX(newest.occurred_at) FROM notification_events AS newest
        WHERE newest.deleted_at IS NULL);
