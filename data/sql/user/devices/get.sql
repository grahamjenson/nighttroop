
--IN: Name string
--OUT: Name string, Password string, TotpSecret string, Role string, CreatedAt *time.Time, LastSeenAt *time.Time
SELECT name,
    password,
    totp_secret,
    role, 
    created_at,
    last_seen_at
FROM devices
where name = ?
LIMIT 1;