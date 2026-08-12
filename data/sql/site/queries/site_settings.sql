
--------
-- READ
--------

-- name: GetSiteSettings :one
SELECT * FROM site_settings WHERE id = 1;

--------
-- UPDATE
--------

-- name: UpdateSiteSettings :one
UPDATE site_settings SET
    units = COALESCE(sqlc.narg('Units'), units),
    timezone = COALESCE(sqlc.narg('Timezone'), timezone),
    bolus_dia_hours = COALESCE(sqlc.narg('BolusDiaHours'), bolus_dia_hours),
    bolus_onset_minutes = COALESCE(sqlc.narg('BolusOnsetMinutes'), bolus_onset_minutes),
    bolus_peak_minutes = COALESCE(sqlc.narg('BolusPeakMinutes'), bolus_peak_minutes),
    basal_dia_hours = COALESCE(sqlc.narg('BasalDiaHours'), basal_dia_hours),
    basal_onset_minutes = COALESCE(sqlc.narg('BasalOnsetMinutes'), basal_onset_minutes),
    basal_peak_minutes = COALESCE(sqlc.narg('BasalPeakMinutes'), basal_peak_minutes),
    carb_max_absorption_hours = COALESCE(sqlc.narg('CarbMaxAbsorptionHours'), carb_max_absorption_hours),
    carb_min_5m_impact = COALESCE(sqlc.narg('CarbMin5mImpact'), carb_min_5m_impact),
    
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER),
    updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID')
RETURNING *; 

--------
-- DELETE
--------
