-- Passed a list of names to delete we update the table deleted at
UPDATE tenants
SET deleted_at = CURRENT_TIMESTAMP
WHERE subdomain = ?;