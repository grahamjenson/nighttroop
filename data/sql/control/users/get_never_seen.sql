-- finds deleted_at is null users where 
--    last_seen_at is null and created_at is more than minutes ago and 
--    OR last_seen_at is not null and last_write_at is null and created_at is more than minutes ago

--IN: lastSeenMinutes float64, lastWriteMinutes float64
--OUT: Subdomain string
SELECT subdomain
FROM users 
WHERE deleted_at IS NULL 
    AND (
            (last_seen_at IS NULL AND created_at < datetime('now', '-' || ? || ' minutes')) 
            OR 
            (last_seen_at IS NOT NULL AND last_write_at IS NULL AND created_at < datetime('now', '-' || ? || ' minutes'))
        
        );