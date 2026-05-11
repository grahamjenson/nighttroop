-- update the last seen at timestamp for a user
--IN: Subdomain string
--OUT:
UPDATE users
SET last_seen_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;