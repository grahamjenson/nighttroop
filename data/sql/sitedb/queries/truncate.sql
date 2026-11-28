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
-- A device report every few minutes is kept only for KeepAllMs before Now; then each device keeps, of its reports,
-- the newest in every SmallMs until WeekMs old, in every MediumMs until FortnightMs old, and in every LargeMs after
-- that. Its IOB, COB, battery and reservoir of each 5 minutes are packaged before (packageSite). The ages and buckets
-- are utils/thin's (db.ThinParams). Each device keeps its newest report, and the
-- loop status on it, however old, until DeleteOldDeviceReports removes it. The buckets are aligned so each hour's
-- newest is also its 20 and 10 minutes' newest. A deleted report is neither kept nor removed, so clients syncing
-- deletions still see it. Only reports written since Since are looked at, as ClearOldLoopStatuses does, and those that
-- passed a week or two weeks old since then, as those are thinned again; 0 looks at them all.
DELETE FROM device_reports WHERE device_reports.id IN (
    SELECT report.id FROM (
        SELECT thinned.id, thinned.created_by_device_name, thinned.occurred_at,
            CASE WHEN thinned.occurred_at < CAST(@now AS INTEGER) - CAST(@fortnight_ms AS INTEGER) THEN CAST(@large_ms AS INTEGER)
                WHEN thinned.occurred_at < CAST(@now AS INTEGER) - CAST(@week_ms AS INTEGER) THEN CAST(@medium_ms AS INTEGER)
                ELSE CAST(@small_ms AS INTEGER) END AS bucket
        FROM device_reports AS thinned
        WHERE thinned.occurred_at < CAST(@now AS INTEGER) - CAST(@keep_all_ms AS INTEGER) AND thinned.deleted_at IS NULL AND (
            thinned.updated_at >= CAST(@since AS INTEGER)
            OR thinned.occurred_at BETWEEN CAST(@since AS INTEGER) - CAST(@week_ms AS INTEGER) AND CAST(@now AS INTEGER) - CAST(@week_ms AS INTEGER)
            OR thinned.occurred_at BETWEEN CAST(@since AS INTEGER) - CAST(@fortnight_ms AS INTEGER) AND CAST(@now AS INTEGER) - CAST(@fortnight_ms AS INTEGER)
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

-- A site far over its DB size limit has the JSON that matters only while recent cleared from its device reports that
-- occurred before Before, a day ago, but each device's newest report, which followers read. Their updated_at stays, so
-- clients do not fetch them again. Its other records are not touched: they are deleted only once their day is packaged.
-- name: ClearOldDeviceReportJSON :execrows
UPDATE device_reports SET loop_status_json = NULL, override_json = NULL, pump_json = NULL, uploader_json = NULL
WHERE device_reports.occurred_at < @before
    AND COALESCE(loop_status_json, override_json, pump_json, uploader_json) IS NOT NULL
    AND device_reports.occurred_at < (SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name AND newest.deleted_at IS NULL);
