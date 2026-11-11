-- credentials_changed_at is when a control admin's password or TOTP last changed, or it was made: JWTs issued before
-- it are refused. SQLite cannot add a column with a default of now, so CreateControlIdentity sets it, and admins made
-- before have 0, which refuses none of their JWTs.
ALTER TABLE control_identities ADD COLUMN credentials_changed_at INTEGER NOT NULL DEFAULT 0;
