-- name: UpsertColdDay :exec
-- Packaging a day again writes it again, and says when. A day that would be written unchanged is not (see upsertColdDay).
INSERT INTO cold_days (kind, date, body) VALUES (sqlc.arg('Kind'), sqlc.arg('Date'), sqlc.arg('Body'))
ON CONFLICT(kind, date) DO UPDATE SET body = excluded.body, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER);

-- name: LastWrittenOfDay :one
-- When a row that happened from From to before To, of the tables whose rows are deleted once their cold days have them,
-- was last written, deleted ones too, or 0 for none
SELECT CAST(COALESCE(MAX(written.updated_at), 0) AS INTEGER) AS written_at FROM (
    SELECT glucose_readings.updated_at FROM glucose_readings WHERE glucose_readings.occurred_at >= sqlc.arg('From') AND glucose_readings.occurred_at < sqlc.arg('To')
    UNION ALL SELECT insulin_doses.updated_at FROM insulin_doses WHERE insulin_doses.occurred_at >= sqlc.arg('From') AND insulin_doses.occurred_at < sqlc.arg('To')
    UNION ALL SELECT food_intakes.updated_at FROM food_intakes WHERE food_intakes.occurred_at >= sqlc.arg('From') AND food_intakes.occurred_at < sqlc.arg('To')
    UNION ALL SELECT temp_basals.updated_at FROM temp_basals WHERE temp_basals.occurred_at >= sqlc.arg('From') AND temp_basals.occurred_at < sqlc.arg('To')
    UNION ALL SELECT temp_targets.updated_at FROM temp_targets WHERE temp_targets.occurred_at >= sqlc.arg('From') AND temp_targets.occurred_at < sqlc.arg('To')
    UNION ALL SELECT activities.updated_at FROM activities WHERE activities.occurred_at >= sqlc.arg('From') AND activities.occurred_at < sqlc.arg('To')
    UNION ALL SELECT notification_events.updated_at FROM notification_events WHERE notification_events.occurred_at >= sqlc.arg('From') AND notification_events.occurred_at < sqlc.arg('To')
    UNION ALL SELECT profile_switches.updated_at FROM profile_switches WHERE profile_switches.occurred_at >= sqlc.arg('From') AND profile_switches.occurred_at < sqlc.arg('To')
    UNION ALL SELECT event_groups.updated_at FROM event_groups WHERE event_groups.occurred_at >= sqlc.arg('From') AND event_groups.occurred_at < sqlc.arg('To')
) AS written;

-- name: GetColdDay :one
SELECT * FROM cold_days WHERE kind = sqlc.arg('Kind') AND date = sqlc.arg('Date');

-- name: ColdDaysNewestFirst :many
-- The days of one kind from From to To, as dates, the newest first, to read a time range a few days at a time
SELECT * FROM cold_days WHERE kind = sqlc.arg('Kind') AND date >= sqlc.arg('FromDate') AND date <= sqlc.arg('ToDate')
ORDER BY date DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: ColdDaysOldestFirst :many
-- The same days, the oldest first, for a search the oldest first
SELECT * FROM cold_days WHERE kind = sqlc.arg('Kind') AND date >= sqlc.arg('FromDate') AND date <= sqlc.arg('ToDate')
ORDER BY date LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: ColdDaysPage :many
-- The cold days of Kind, after DateGT and before DateLT (each when given), the newest first, for the Cold Storage page
SELECT * FROM cold_days
WHERE (CAST(sqlc.narg('Kind') AS TEXT) IS NULL OR kind = sqlc.narg('Kind'))
    AND (CAST(sqlc.narg('DateGT') AS TEXT) IS NULL OR date > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS TEXT) IS NULL OR date < sqlc.narg('DateLT'))
ORDER BY date DESC, kind
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ColdDaysPageOldest :many
-- ColdDaysPage, the oldest first
SELECT * FROM cold_days
WHERE (CAST(sqlc.narg('Kind') AS TEXT) IS NULL OR kind = sqlc.narg('Kind'))
    AND (CAST(sqlc.narg('DateGT') AS TEXT) IS NULL OR date > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS TEXT) IS NULL OR date < sqlc.narg('DateLT'))
ORDER BY date, kind
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ColdDayTotals :one
-- How many cold days of Kind, after DateGT and before DateLT (each when given), there are, and the bytes of their bodies
SELECT COUNT(*) AS total, CAST(COALESCE(SUM(length(body)), 0) AS INTEGER) AS bytes FROM cold_days
WHERE (CAST(sqlc.narg('Kind') AS TEXT) IS NULL OR kind = sqlc.narg('Kind'))
    AND (CAST(sqlc.narg('DateGT') AS TEXT) IS NULL OR date > sqlc.narg('DateGT'))
    AND (CAST(sqlc.narg('DateLT') AS TEXT) IS NULL OR date < sqlc.narg('DateLT'));

-- name: ColdDayKinds :many
-- The kinds of data the site has cold days of, to filter by
SELECT DISTINCT kind FROM cold_days ORDER BY kind;

-- name: HasColdDays :one
-- Whether any day was ever packaged: a site with none has all its days packaged, not only those written lately
SELECT EXISTS (SELECT 1 FROM cold_days);

-- name: CGMDaysWrittenSince :many
-- The UTC days, as days since 1970, that ended before Before and had a CGM reading written, changed or deleted since
-- Since: the days to package again
SELECT DISTINCT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM glucose_readings
WHERE updated_at >= sqlc.arg('Since') AND reading_type = 'sgv' AND occurred_at < sqlc.arg('Before')
ORDER BY 1 LIMIT 5000;

-- name: CGMReadingsOfDay :many
-- A day's CGM readings, the deleted too, as a deleted one clears its step of a day packaged before
SELECT * FROM glucose_readings
WHERE occurred_at >= sqlc.arg('From') AND occurred_at < sqlc.arg('To') AND reading_type = 'sgv'
ORDER BY occurred_at LIMIT 5000;

-- name: DeletePackagedCGMReadings :execrows
-- The CGM readings from before Before whose day's glucose was packaged after they were last written, so none is deleted
-- that its cold day does not have. One with a note, or that a treatment refers to, is kept: it is an event.
DELETE FROM glucose_readings
WHERE reading_type = 'sgv' AND glucose_readings.occurred_at < sqlc.arg('Before') AND notes IS NULL
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'glucose' AND date = date(glucose_readings.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= glucose_readings.updated_at)
AND id NOT IN (SELECT glucose_reading_id FROM event_groups WHERE glucose_reading_id IS NOT NULL);

-- name: DeviceReportDaysWrittenSince :many
-- The UTC days, as days since 1970, up to Before that had a device report written, changed or deleted since Since:
-- the days whose device series are packaged again
SELECT DISTINCT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM device_reports
WHERE updated_at >= sqlc.arg('Since') AND occurred_at < sqlc.arg('Before')
ORDER BY 1 LIMIT 5000;

-- name: DeviceReportNumbersOfDay :many
-- The numbers of a day's device reports that are packaged as series, the oldest first, without the JSON, which is
-- most of a report; the deleted too, as a deleted one clears its step
SELECT occurred_at, deleted_at, loop_name, loop_iob, loop_iob_basal, loop_cob, pump_reservoir, pump_battery_percent, device_battery_percent, pump_state
FROM device_reports
WHERE occurred_at >= sqlc.arg('From') AND occurred_at < sqlc.arg('To')
ORDER BY occurred_at LIMIT 5000;

-- name: TreatmentDaysWrittenSince :many
-- The UTC days, as days since 1970, that ended before Before and had an insulin dose, a food intake or a temp basal
-- written, changed or deleted since Since: the days whose treatment series are packaged again
SELECT DISTINCT day FROM (
    SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM insulin_doses WHERE insulin_doses.updated_at >= sqlc.arg('Since') AND insulin_doses.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) FROM food_intakes WHERE food_intakes.updated_at >= sqlc.arg('Since') AND food_intakes.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) FROM temp_basals WHERE temp_basals.updated_at >= sqlc.arg('Since') AND temp_basals.occurred_at < sqlc.arg('Before')
) ORDER BY 1 LIMIT 5000;

-- name: EventDaysWrittenSince :many
-- The UTC days, as days since 1970, that ended before Before and had an event written, changed or deleted since
-- Since: a meal, a dose, a meter's reading or one with a note, an activity, a temp target, a profile switch, a
-- notification, or a treatment of them. The days whose events are packaged again.
SELECT DISTINCT day FROM (
    SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM food_intakes WHERE food_intakes.updated_at >= sqlc.arg('Since') AND food_intakes.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM insulin_doses WHERE insulin_doses.updated_at >= sqlc.arg('Since') AND insulin_doses.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM glucose_readings WHERE glucose_readings.updated_at >= sqlc.arg('Since') AND glucose_readings.occurred_at < sqlc.arg('Before') AND (glucose_readings.reading_type = 'mbg' OR glucose_readings.notes IS NOT NULL)
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM activities WHERE activities.updated_at >= sqlc.arg('Since') AND activities.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM temp_targets WHERE temp_targets.updated_at >= sqlc.arg('Since') AND temp_targets.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM profile_switches WHERE profile_switches.updated_at >= sqlc.arg('Since') AND profile_switches.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM notification_events WHERE notification_events.updated_at >= sqlc.arg('Since') AND notification_events.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM event_groups WHERE event_groups.updated_at >= sqlc.arg('Since') AND event_groups.occurred_at < sqlc.arg('Before') AND event_groups.temp_basal_id IS NULL
) ORDER BY 1 LIMIT 5000;

-- name: EventGroupsOfDay :many
-- A day's treatments, without their events, but those of temp basals, which are not events: a loop writes one every
-- 5 minutes, and their rate is packaged as a series
SELECT * FROM event_groups
WHERE occurred_at >= sqlc.arg('From') AND occurred_at < sqlc.arg('To') AND temp_basal_id IS NULL AND deleted_at IS NULL
ORDER BY occurred_at LIMIT 5000;

-- name: DeletePackagedEventGroups :execrows
-- The treatments from before Before whose day's events were packaged after they were last written, and not since
-- SettledBefore, so the day is as it will stay. A treatment goes before its events, which it refers to.
DELETE FROM event_groups
WHERE event_groups.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(event_groups.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= event_groups.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'));

-- name: DeletePackagedInsulinDoses :execrows
-- The insulin doses from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM insulin_doses
WHERE insulin_doses.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(insulin_doses.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= insulin_doses.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT insulin_dose_id FROM event_groups WHERE insulin_dose_id IS NOT NULL);

-- name: DeletePackagedFoodIntakes :execrows
-- The food intakes from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM food_intakes
WHERE food_intakes.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(food_intakes.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= food_intakes.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT food_intake_id FROM event_groups WHERE food_intake_id IS NOT NULL);

-- name: DeletePackagedTempBasals :execrows
-- The temp basals from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM temp_basals
WHERE temp_basals.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(temp_basals.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= temp_basals.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT temp_basal_id FROM event_groups WHERE temp_basal_id IS NOT NULL);

-- name: DeletePackagedTempTargets :execrows
-- The temp targets from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM temp_targets
WHERE temp_targets.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(temp_targets.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= temp_targets.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT temp_target_id FROM event_groups WHERE temp_target_id IS NOT NULL);

-- name: DeletePackagedActivities :execrows
-- The activities from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM activities
WHERE activities.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(activities.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= activities.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT activity_id FROM event_groups WHERE activity_id IS NOT NULL);

-- name: DeletePackagedNotificationEvents :execrows
-- The notifications from before Before whose day was packaged after they were last written, and not since SettledBefore, that
-- no treatment still refers to
DELETE FROM notification_events
WHERE notification_events.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(notification_events.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= notification_events.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT notification_event_id FROM event_groups WHERE notification_event_id IS NOT NULL);

-- name: DeletePackagedProfileSwitches :execrows
-- The profile switches from before Before whose day was packaged after they were last written, and not since
-- SettledBefore, that no treatment still refers to, but those that may still be in force: the newest with no end, and
-- one with an end that is not before Before. The basal of the days that are still rows is read from them.
DELETE FROM profile_switches
WHERE profile_switches.occurred_at < sqlc.arg('Before')
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'events' AND date = date(profile_switches.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= profile_switches.updated_at AND cold_days.updated_at <= sqlc.arg('SettledBefore'))
AND id NOT IN (SELECT profile_switch_id FROM event_groups WHERE profile_switch_id IS NOT NULL)
AND NOT (duration_minutes > 0 AND profile_switches.occurred_at + duration_minutes * 60000 >= sqlc.arg('Before'))
AND NOT (duration_minutes = 0 AND profile_switches.occurred_at >= (SELECT COALESCE(MAX(newest.occurred_at), 0) FROM profile_switches AS newest
    WHERE newest.duration_minutes = 0 AND newest.deleted_at IS NULL AND newest.occurred_at < sqlc.arg('Before')));

-- name: MarkColdDaysStale :exec
-- A record of the day was deleted for good, or moved to another day, which leaves no row that says the day changed:
-- its cold days are marked, by a time written of 0, to be packaged again from its rows
UPDATE cold_days SET updated_at = 0 WHERE date = sqlc.arg('Date');

-- name: StaleColdDays :many
-- The dates from From that are marked to be packaged again
SELECT DISTINCT date FROM cold_days WHERE updated_at = 0 AND date >= sqlc.arg('FromDate') ORDER BY date LIMIT 5000;

-- name: ClearStaleColdDays :exec
-- A run that packaged the marked days again clears the mark of a cold day it did not write, as one of a series the
-- day no longer has a value of
UPDATE cold_days SET updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER) WHERE updated_at = 0 AND date >= sqlc.arg('FromDate');

-- name: SummaryDataWrittenSince :many
-- When the data a day's summary is built from, which a report reads from the day's cold days, that was written, changed or deleted after Since, and happened before
-- Before, happened, by the 15 minutes each is in, as every timezone's midnight starts one. The + keeps SQLite on
-- updated_at's index, as few rows are written since, while nearly all happened before
SELECT DISTINCT CAST(written.occurred_at / 900000 * 900000 AS INTEGER) AS occurred_at FROM (
    SELECT glucose_readings.occurred_at FROM glucose_readings
        WHERE glucose_readings.updated_at > @since AND +glucose_readings.occurred_at < @before
    UNION ALL SELECT insulin_doses.occurred_at FROM insulin_doses
        WHERE insulin_doses.updated_at > @since AND +insulin_doses.occurred_at < @before
    UNION ALL SELECT food_intakes.occurred_at FROM food_intakes
        WHERE food_intakes.updated_at > @since AND +food_intakes.occurred_at < @before
    UNION ALL SELECT temp_basals.occurred_at FROM temp_basals
        WHERE temp_basals.updated_at > @since AND +temp_basals.occurred_at < @before
    UNION ALL SELECT temp_targets.occurred_at FROM temp_targets
        WHERE temp_targets.updated_at > @since AND +temp_targets.occurred_at < @before
    UNION ALL SELECT profile_switches.occurred_at FROM profile_switches
        WHERE profile_switches.updated_at > @since AND +profile_switches.occurred_at < @before
    UNION ALL SELECT event_groups.occurred_at FROM event_groups
        WHERE event_groups.updated_at > @since AND +event_groups.occurred_at < @before
    UNION ALL SELECT device_reports.occurred_at FROM device_reports
        WHERE device_reports.updated_at > @since AND +device_reports.occurred_at < @before
) AS written;
