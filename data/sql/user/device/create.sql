--IN: name string, password string, totp_secret string, role string
--OUT:
INSERT INTO devices (name, password, totp_secret, role)
VALUES (?, ?, ?, ?);