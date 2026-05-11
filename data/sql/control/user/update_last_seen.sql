-- update the last seen at timestamp for a user
--IN: subdomain string
--OUT:
UPDATE users
SET last_seen_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;