-- Passed a list of names to delete we update the table deleted at
--IN: subdomains_json_list string
--OUT:
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP
WHERE subdomain IN (SELECT value FROM json_each(?));