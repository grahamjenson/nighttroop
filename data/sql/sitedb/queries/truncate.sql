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
-- A device report every few minutes is kept only for the last hour, before Before; earlier, each device keeps one in
-- every 10 minutes, the newest, which is enough to chart its IOB, COB, battery and reservoir, and keeps its newest
-- report, and the loop status on it, however old. A deleted report is neither kept nor removed, so clients syncing
-- deletions still see it. Only reports written since Since are looked at, as ClearOldLoopStatuses does.
DELETE FROM device_reports
WHERE device_reports.updated_at >= @since AND device_reports.occurred_at < @before AND device_reports.deleted_at IS NULL
    AND EXISTS (
        SELECT 1 FROM device_reports AS kept
        WHERE kept.created_by_device_name IS device_reports.created_by_device_name AND kept.deleted_at IS NULL
            AND kept.occurred_at >= device_reports.occurred_at
            AND kept.occurred_at < (device_reports.occurred_at / 600000 + 1) * 600000
            AND (kept.occurred_at > device_reports.occurred_at OR kept.id > device_reports.id)
    );
