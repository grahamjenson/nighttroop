--IN: Name string
--OUT: Name string, Password string, Role string, TotpSecret string, LastSeenAt *time.Time, CreatedAt *time.Time
SELECT name, password, role, totp_secret, last_seen_at, created_at FROM admins WHERE name = ?;