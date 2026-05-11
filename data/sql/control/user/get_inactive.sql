-- Inactive Users are users that were once active but have stopped, so the assumption is they have data 

--IN: last_seen_minutes float64, last_write_minutes float64
--OUT: subdomain string
SELECT subdomain
FROM users 
WHERE deleted_at IS NULL AND (last_seen_at < datetime('now', '-' || ? || ' minutes') OR last_write_at < datetime('now', '-' || ? || ' minutes'));