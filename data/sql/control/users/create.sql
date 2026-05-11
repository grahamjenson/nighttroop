
--IN: Subdomain string, BackupKeyHash string
--OUT:
INSERT INTO users (subdomain, backup_key_hash)
VALUES (?, ?);