SELECT name,
    password,
    totp_secret,
    role, 
    created_at,
    last_seen_at
FROM devices
where name = ?
LIMIT 1;