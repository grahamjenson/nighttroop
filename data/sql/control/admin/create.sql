--IN: name string, password string, role string, totp_secret string
--OUT:
INSERT INTO admins (name, password, role, totp_secret)
VALUES (?, ?, ?, ?);