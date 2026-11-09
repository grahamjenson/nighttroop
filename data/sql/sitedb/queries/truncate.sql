-- name: ClearOldLoopStatuses :exec
-- A loop's status, most of a device report and of a site's DB, is kept only on its device's newest report with one,
-- as followers (LoopFollow, the Nightscout web app) read its predictions and reason from that one. Once a newer one
-- comes in, the others are cleared: what is worth keeping of them, the loop's IOB and COB, was extracted into their own
-- columns when they were written. Only reports written since Since are looked at, by their updated_at index, as those
-- before were cleared by an earlier run; 0 looks at them all.
UPDATE device_reports SET loop_status_json = NULL
WHERE device_reports.updated_at >= @since AND device_reports.loop_status_json IS NOT NULL
    AND device_reports.occurred_at < (
        SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name
            AND newest.loop_status_json IS NOT NULL AND newest.deleted_at IS NULL
    );

-- name: ClearOldPumps :exec
-- A pump as sent, the most of a report after its loop's status, is kept as that is: only on its device's newest report
-- with one, as followers read it from that one. What is worth keeping of the others, their battery, reservoir, clock,
-- manufacturer and model, was extracted into their own columns when they were written.
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

-- name: DeleteOldDeviceReports :exec
-- Every device report that occurred before Before, 30 days ago, is deleted, a deleted one too, by the occurred_at index
DELETE FROM device_reports WHERE occurred_at < @before;
