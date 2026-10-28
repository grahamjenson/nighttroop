-------
-- CREATE
--------

-- name: UpsertAlarm :exec
INSERT INTO alarms (site_id, identity_id, type, active_since, notified_at, snoozed_until, alerting)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (site_id, identity_id, type) DO UPDATE SET
    active_since = excluded.active_since,
    notified_at = excluded.notified_at,
    snoozed_until = excluded.snoozed_until,
    alerting = excluded.alerting;

-- name: SnoozeAlarm :exec
-- Snoozes an identity's alarm until a time, whether it is kept or not, so the snooze also covers it starting again
INSERT INTO alarms (site_id, identity_id, type, active_since, notified_at, snoozed_until, alerting)
VALUES (sqlc.arg('SiteID'), sqlc.arg('IdentityID'), sqlc.arg('Type'), 0, 0, sqlc.arg('Until'), 0)
ON CONFLICT (site_id, identity_id, type) DO UPDATE SET snoozed_until = excluded.snoozed_until;

-------
-- READ
--------

-- name: SiteAlarms :many
SELECT * FROM alarms WHERE site_id = sqlc.arg('SiteID');

-- name: AlarmSiteIDs :many
SELECT DISTINCT site_id FROM alarms;

-------
-- DELETE
--------

-- name: DeleteAlarm :exec
DELETE FROM alarms WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID') AND type = sqlc.arg('Type');
