-- update the last write at timestamp for a user
--IN: subdomain string
--OUT:
UPDATE users
SET last_write_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;