--IN: name string
--OUT: name string, password string, role string, totp_secret string, last_seen_at *time.Time, created_at *time.Time
SELECT name, password, role, totp_secret, last_seen_at, created_at FROM admins WHERE name = ?;