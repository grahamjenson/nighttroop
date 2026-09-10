-------
-- CREATE 
--------

-- name: CreateSiteIdentity :exec
INSERT INTO site_identities (
    site_id,
    name,
    password,
    totp_secret,
    role,
    created_by_identity_id,
    created_by_identity_name
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
 );




--------
-- READ
--------


-- name: AllSiteIdentities :many
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') ORDER BY created_at DESC;


-- name: GetSiteIdentity :one
SELECT * FROM site_identities WHERE site_id = sqlc.arg('SiteID') AND name = sqlc.arg('Name') LIMIT 1;


--------
-- UPDATE
--------

-- name: UpdateSiteIdentityDateFlags :exec
UPDATE site_identities
SET
    last_seen_at = COALESCE(?,
     last_seen_at),
    last_write_at = COALESCE(?, last_write_at)
WHERE id = ?;


-- name: UpdateSiteIdentitySettings :exec
UPDATE site_identities SET
    -- IDENTITY SETTINGS
    units = sqlc.narg('Units'),

    -- DATE
    timezone = sqlc.narg('Timezone'),
    language = sqlc.narg('Language'),
    date_format = sqlc.narg('DateFormat'),
    time_format = sqlc.narg('TimeFormat'),
    week_start_day = sqlc.narg('WeekStartDay'),

    -- INSULIN
    bolus_dia_hours = sqlc.narg('BolusDiaHours'),
    bolus_onset_minutes = sqlc.narg('BolusOnsetMinutes'),
    bolus_peak_minutes = sqlc.narg('BolusPeakMinutes'),

    basal_dia_hours = sqlc.narg('BasalDiaHours'),
    basal_onset_minutes = sqlc.narg('BasalOnsetMinutes'),
    basal_peak_minutes = sqlc.narg('BasalPeakMinutes'),

    default_insulin_sensitivity = sqlc.narg('DefaultInsulinSensitivity'),
    default_carb_ratio = sqlc.narg('DefaultCarbRatio'),
    default_basal_rate = sqlc.narg('DefaultBasalRate'),

    -- CARBS
    carb_max_absorption_hours = sqlc.narg('CarbMaxAbsorptionHours'),
    carb_min_5m_impact = sqlc.narg('CarbMin5mImpact'),


    -- ALERTS
    alarm_urgent_low_enabled = sqlc.narg('AlarmUrgentLowEnabled'),
    alarm_urgent_low = sqlc.narg('AlarmUrgentLow'),

    alarm_low_enabled = sqlc.narg('AlarmLowEnabled'),
    alarm_low = sqlc.narg('AlarmLow'),
    alarm_low_minutes = sqlc.narg('AlarmLowMinutes'),
    alarm_low_prediction_enabled = sqlc.narg('AlarmLowPredictionEnabled'),

    alarm_high_enabled = sqlc.narg('AlarmHighEnabled'),
    alarm_high = sqlc.narg('AlarmHigh'),
    alarm_high_minutes = sqlc.narg('AlarmHighMinutes'),
    alarm_urgent_high_enabled = sqlc.narg('AlarmUrgentHighEnabled'),
    alarm_urgent_high = sqlc.narg('AlarmUrgentHigh'),

    alarm_fast_rise_enabled = sqlc.narg('AlarmFastRiseEnabled'),
    alarm_fast_rise = sqlc.narg('AlarmFastRise'),
    alarm_fast_rise_minutes = sqlc.narg('AlarmFastRiseMinutes'),

    alarm_fast_drop_enabled = sqlc.narg('AlarmFastDropEnabled'),
    alarm_fast_drop = sqlc.narg('AlarmFastDrop'),
    alarm_fast_drop_minutes = sqlc.narg('AlarmFastDropMinutes'),

    alarm_no_data_enabled = sqlc.narg('AlarmNoDataEnabled'),
    alarm_no_data_minutes = sqlc.narg('AlarmNoDataMinutes'),

    alarm_battery_low_enabled = sqlc.narg('AlarmBatteryLowEnabled'),
    alarm_battery_low_percent = sqlc.narg('AlarmBatteryLowPercent'),
    alarm_battery_low_pwd_enabled = sqlc.narg('AlarmBatteryLowPwdEnabled'),
    alarm_battery_low_pwd_percent = sqlc.narg('AlarmBatteryLowPwdPercent'),

    alarm_low_reservoir_enabled = sqlc.narg('AlarmLowReservoirEnabled'),
    alarm_low_reservoir_units = sqlc.narg('AlarmLowReservoirUnits'),

    -- NOTIFICATIONS
    notifications_enabled = sqlc.narg('NotificationsEnabled')
WHERE id = sqlc.arg('ID') AND site_id = sqlc.arg('SiteID');

--------
-- DELETE
--------