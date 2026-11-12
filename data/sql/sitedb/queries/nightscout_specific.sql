
-- name: LastModifiedCollections :one
-- Deleted rows count, as a soft delete moves updated_at forward, at least 1ms, so a history since a client's
-- lastModified has the documents deleted after it
SELECT 
     CAST(COALESCE((SELECT updated_at FROM device_reports                             ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS device_report_updated_at,
     CAST(COALESCE((SELECT updated_at FROM activities                                 ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS activity_updated_at,
     CAST(COALESCE((SELECT updated_at FROM glucose_readings                           ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS glucose_reading_updated_at,
     CAST(COALESCE((SELECT updated_at FROM event_groups                               ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS event_group_updated_at,
     CAST(COALESCE((SELECT updated_at FROM food_presets                               ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS food_preset_updated_at,
     CAST(COALESCE((SELECT updated_at FROM profile_presets                            ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS profile_preset_updated_at,
     CAST(COALESCE((SELECT updated_at FROM settings_documents                         ORDER BY updated_at DESC                   LIMIT 1),0) AS INTEGER) AS settings_document_updated_at
;




