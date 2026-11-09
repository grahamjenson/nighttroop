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
