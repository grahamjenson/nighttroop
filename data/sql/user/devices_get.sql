SELECT name,
    token
FROM devices
where name = ? OR token = ?
LIMIT 1;