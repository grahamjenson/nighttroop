
-- name: LastModifiedCollections :one
SELECT 
     CAST(COALESCE((SELECT updated_at FROM device_reports    WHERE deleted_at IS NULL ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS device_report_updated_at,
     CAST(COALESCE((SELECT updated_at FROM activities        WHERE deleted_at IS NULL ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS activity_updated_at,
     CAST(COALESCE((SELECT updated_at FROM glucose_readings  WHERE deleted_at IS NULL ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS glucose_reading_updated_at,
     CAST(COALESCE((SELECT updated_at FROM event_groups      WHERE deleted_at IS NULL ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS event_group_updated_at,
     CAST(COALESCE((SELECT updated_at FROM food_presets                               ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS food_preset_updated_at,
     CAST(COALESCE((SELECT updated_at FROM profile_presets                            ORDER BY updated_at DESC, is_default DESC  LIMIT 1),0) AS INTEGER) AS profile_preset_updated_at
;




