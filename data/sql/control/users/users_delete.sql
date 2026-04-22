-- Passed a list of names to delete we update the table deleted at
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE subdomain IN (SELECT value FROM json_each(?));