-------
-- CREATE
--------

-- name: UpsertAlarm :exec
-- Keeps a check's state of an alarm, from the snooze it read. Only SnoozeAlarm changes a snooze, so one made during
-- the check is kept, and the alarm stays alerting, as the check sent it, so the next check finds it snoozed while
-- alerting and stops it.
INSERT INTO alarms (site_id, identity_id, type, active_since, notified_at, snoozed_until, alerting)
VALUES (sqlc.arg('SiteID'), sqlc.arg('IdentityID'), sqlc.arg('Type'), sqlc.arg('ActiveSince'), sqlc.arg('NotifiedAt'),
    sqlc.arg('SnoozedUntil'), sqlc.arg('Alerting'))
ON CONFLICT (site_id, identity_id, type) DO UPDATE SET
    active_since = excluded.active_since,
    notified_at = excluded.notified_at,
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

-- name: SnoozedAlertingSiteIDs :many
-- The sites with an alarm snoozed since it was checked, which a check stops alerting
SELECT DISTINCT site_id FROM alarms WHERE alerting = 1 AND snoozed_until > sqlc.arg('Now');

-- name: IdentityAlarms :many
-- An identity's active alarms, the most urgent first, as alarms.Types lists them
SELECT * FROM alarms
WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID') AND active_since != 0
ORDER BY CASE type WHEN 'urgent_low' THEN 0 WHEN 'low' THEN 1 WHEN 'urgent_high' THEN 2 WHEN 'high' THEN 3 ELSE 4 END;

-- name: GetIdentityAlarm :one
-- One of an identity's active alarms
SELECT * FROM alarms
WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID') AND type = sqlc.arg('Type') AND active_since != 0;

-------
-- DELETE
--------

-- name: DeleteAlarm :exec
-- Forgets an alarm, unless it was snoozed since its check read the snooze it has
DELETE FROM alarms WHERE site_id = sqlc.arg('SiteID') AND identity_id = sqlc.arg('IdentityID') AND type = sqlc.arg('Type')
AND snoozed_until = sqlc.arg('SnoozedUntil');
