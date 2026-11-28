-- name: UpsertColdDay :exec
-- Sealing a day again writes it again, and says when
INSERT INTO cold_days (kind, date, body) VALUES (sqlc.arg('Kind'), sqlc.arg('Date'), sqlc.arg('Body'))
ON CONFLICT(kind, date) DO UPDATE SET body = excluded.body, updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER);

-- name: GetColdDay :one
SELECT * FROM cold_days WHERE kind = sqlc.arg('Kind') AND date = sqlc.arg('Date');

-- name: ColdDaysNewestFirst :many
-- The days of one kind from From to To, as dates, the newest first, to read a time range a few days at a time
SELECT * FROM cold_days WHERE kind = sqlc.arg('Kind') AND date >= sqlc.arg('FromDate') AND date <= sqlc.arg('ToDate')
ORDER BY date DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: ColdDaysPage :many
-- Every cold day, the newest first, for the page that shows what the site keeps of its old days
SELECT * FROM cold_days ORDER BY date DESC, kind
LIMIT CAST(sqlc.arg('Limit') AS INTEGER) OFFSET CAST(sqlc.arg('Skip') AS INTEGER);

-- name: ColdDayTotals :one
-- How many cold days there are, and the bytes of their bodies
SELECT COUNT(*) AS total, CAST(COALESCE(SUM(length(body)), 0) AS INTEGER) AS bytes FROM cold_days;

-- name: HasColdDays :one
-- Whether any day was ever sealed: a site with none has all its days sealed, not only those written lately
SELECT EXISTS (SELECT 1 FROM cold_days);

-- name: CGMDaysWrittenSince :many
-- The UTC days, as days since 1970, that ended before Before and had a CGM reading written, changed or deleted since
-- Since: the days to seal again
SELECT DISTINCT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM glucose_readings
WHERE updated_at >= sqlc.arg('Since') AND reading_type = 'sgv' AND occurred_at < sqlc.arg('Before')
ORDER BY 1 LIMIT 5000;

-- name: CGMReadingsOfDay :many
-- A day's CGM readings, the deleted too, as a deleted one clears its step of a day sealed before
SELECT * FROM glucose_readings
WHERE occurred_at >= sqlc.arg('From') AND occurred_at < sqlc.arg('To') AND reading_type = 'sgv'
ORDER BY occurred_at LIMIT 5000;

-- name: DeleteSealedCGMReadings :execrows
-- The CGM readings from before Before whose day's glucose was sealed after they were last written, so none is deleted
-- that its cold day does not have. One with a note, or that a treatment refers to, is kept: it is an event.
DELETE FROM glucose_readings
WHERE reading_type = 'sgv' AND glucose_readings.occurred_at < sqlc.arg('Before') AND notes IS NULL
AND EXISTS (SELECT 1 FROM cold_days WHERE kind = 'glucose' AND date = date(glucose_readings.occurred_at / 1000, 'unixepoch')
    AND cold_days.updated_at >= glucose_readings.updated_at)
AND id NOT IN (SELECT glucose_reading_id FROM event_groups WHERE glucose_reading_id IS NOT NULL);

-- name: DeviceReportDaysWrittenSince :many
-- The UTC days, as days since 1970, up to Before that had a device report written, changed or deleted since Since:
-- the days whose device series are sealed again
SELECT DISTINCT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM device_reports
WHERE updated_at >= sqlc.arg('Since') AND occurred_at < sqlc.arg('Before')
ORDER BY 1 LIMIT 5000;

-- name: DeviceReportNumbersOfDay :many
-- The numbers of a day's device reports that are sealed as series, the oldest first, without the JSON, which is
-- most of a report; the deleted too, as a deleted one clears its step
SELECT occurred_at, deleted_at, loop_name, loop_iob, loop_iob_basal, loop_cob, pump_reservoir, pump_battery_percent, device_battery_percent
FROM device_reports
WHERE occurred_at >= sqlc.arg('From') AND occurred_at < sqlc.arg('To')
ORDER BY occurred_at LIMIT 5000;

-- name: TreatmentDaysWrittenSince :many
-- The UTC days, as days since 1970, that ended before Before and had an insulin dose, a food intake or a temp basal
-- written, changed or deleted since Since: the days whose treatment series are sealed again
SELECT DISTINCT day FROM (
    SELECT CAST(occurred_at / 86400000 AS INTEGER) AS day FROM insulin_doses WHERE insulin_doses.updated_at >= sqlc.arg('Since') AND insulin_doses.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) FROM food_intakes WHERE food_intakes.updated_at >= sqlc.arg('Since') AND food_intakes.occurred_at < sqlc.arg('Before')
    UNION SELECT CAST(occurred_at / 86400000 AS INTEGER) FROM temp_basals WHERE temp_basals.updated_at >= sqlc.arg('Since') AND temp_basals.occurred_at < sqlc.arg('Before')
) ORDER BY 1 LIMIT 5000;
