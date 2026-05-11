
--IN: subdomain string, backup_key_hash string
--OUT:
INSERT INTO users (subdomain, backup_key_hash)
VALUES (?, ?);