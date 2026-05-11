-- finds deleted_at is null users where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago

--IN: last_seen_minutes float64, last_write_minutes float64
--OUT: subdomain string
SELECT subdomain
FROM users 
WHERE deleted_at IS NULL 
    AND (
            (last_seen_at IS NULL AND created_at < datetime('now', '-' || ? || ' minutes')) 
            OR 
            (last_seen_at IS NOT NULL AND last_write_at IS NULL AND created_at < datetime('now', '-' || ? || ' minutes'))
        
        );