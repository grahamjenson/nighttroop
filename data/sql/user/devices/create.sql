--IN: Name string, Password string, TotpSecret string, Role string
--OUT:
INSERT INTO devices (name, password, totp_secret, role)
VALUES (?, ?, ?, ?);