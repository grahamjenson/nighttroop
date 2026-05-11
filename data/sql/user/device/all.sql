--IN:
--OUT: name string, password string, totp_secret string, role string, created_at *time.Time, last_seen_at *time.Time
SELECT name, password, totp_secret, role, created_at, last_seen_at FROM devices;