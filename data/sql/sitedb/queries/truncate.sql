-- name: ThinOldLoopStatuses :exec
-- A loop's status, most of a device report and of a site's DB, is kept whole only on its device's newest report with
-- one, as followers (LoopFollow, the Nightscout web app) read its predictions and reason from that one. The others
-- keep what is read of an old status, its IOB and COB (also in their own columns) and the rest, and lose:
-- - its debug output (OpenAPS's consoleLog and consoleError) and reason, in suggested and enacted, the longest text;
-- - its predicted glucoses (OpenAPS's predBGs, Loop's predicted);
-- - Trio's IOB with a zero temp and last temp, and an oref0 rig's projected IOBs, of which the first, now, is kept.
-- Only reports written since Since are looked at, by their updated_at index, as those before were thinned by an earlier
-- run; 0 looks at them all.
UPDATE device_reports SET loop_status_json = json_remove(
    CASE WHEN json_type(device_reports.loop_status_json, '$.iob') = 'array'
        THEN json_set(device_reports.loop_status_json, '$.iob', json_extract(device_reports.loop_status_json, '$.iob[0]'))
        ELSE device_reports.loop_status_json END,
    '$.suggested.consoleLog', '$.suggested.consoleError', '$.suggested.predBGs', '$.suggested.reason',
    '$.enacted.consoleLog', '$.enacted.consoleError', '$.enacted.predBGs', '$.enacted.reason',
    '$.iob.iobWithZeroTemp', '$.iob.lastTemp', '$.predicted')
WHERE device_reports.updated_at >= @since AND device_reports.loop_status_json IS NOT NULL
    AND device_reports.occurred_at < (
        SELECT MAX(newest.occurred_at) FROM device_reports AS newest
        WHERE newest.created_by_device_name IS device_reports.created_by_device_name
            AND newest.loop_status_json IS NOT NULL AND newest.deleted_at IS NULL
    )
    AND (
        json_type(device_reports.loop_status_json, '$.suggested.consoleLog') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.suggested.consoleError') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.suggested.predBGs') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.suggested.reason') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.enacted.consoleLog') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.enacted.consoleError') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.enacted.predBGs') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.enacted.reason') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.iob') = 'array'
        OR json_type(device_reports.loop_status_json, '$.iob.iobWithZeroTemp') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.iob.lastTemp') IS NOT NULL
        OR json_type(device_reports.loop_status_json, '$.predicted') IS NOT NULL
    );
