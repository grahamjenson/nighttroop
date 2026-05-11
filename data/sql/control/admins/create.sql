--IN: Name string,Password string,Role string,TotpSecret string
--OUT:
INSERT INTO admins (name, password, role, totp_secret)
VALUES (?, ?, ?, ?);