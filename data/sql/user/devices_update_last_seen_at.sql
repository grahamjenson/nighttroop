-- update the last seen at timestamp for a devices
UPDATE devices
SET last_seen_at = CURRENT_TIMESTAMP
WHERE name = ?;