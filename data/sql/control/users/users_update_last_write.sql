-- update the last write at timestamp for a user
UPDATE users
SET last_write_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;