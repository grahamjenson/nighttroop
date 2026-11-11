-- name: CreateControlIdentity :exec
INSERT INTO control_identities (
    name,
    password_hash,
    totp_secret_enc,
    role,
    created_by_identity_id,
    created_by_identity_name
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
 );



-- name: AllControlIdentities :many
-- Control has at most 20 admins, see controlapi.maxControlAdmins
SELECT * FROM control_identities ORDER BY created_at DESC LIMIT 100;


-- A server without an admin is set up, whatever its admins are named, see control_setup.go
-- name: CountControlAdmins :one
SELECT COUNT(*) FROM control_identities WHERE role = 'admin';

-- name: GetControlIdentity :one
SELECT * FROM control_identities WHERE name = sqlc.arg('Name') LIMIT 1;
-- name: GetControlIdentityByID :one
SELECT * FROM control_identities WHERE id = sqlc.arg('ID');

-- A control admin manages the others. The last admin is never deleted or made read only, which the WHERE checks as it
-- writes, so two admins removing each other at once cannot leave none.

-- name: UpdateControlIdentityRole :execresult
UPDATE control_identities
SET role = sqlc.arg('Role'), updated_by_identity_id = sqlc.arg('UpdatedByIdentityID'), updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE control_identities.id = sqlc.arg('ID') AND (control_identities.role != 'admin' OR (SELECT COUNT(*) FROM control_identities AS admins WHERE admins.role = 'admin') > 1);

-- name: DeleteControlIdentity :execresult
DELETE FROM control_identities
WHERE control_identities.id = sqlc.arg('ID') AND (control_identities.role != 'admin' OR (SELECT COUNT(*) FROM control_identities AS admins WHERE admins.role = 'admin') > 1);

-- name: UpdateControlIdentityPassword :exec
UPDATE control_identities
SET password_hash = sqlc.arg('PasswordHash'), updated_by_identity_id = sqlc.arg('UpdatedByIdentityID'), updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID');

-- name: UpdateControlIdentityTOTP :exec
UPDATE control_identities
SET totp_secret_enc = sqlc.arg('TotpSecretEnc'), updated_by_identity_id = sqlc.arg('UpdatedByIdentityID'), updated_by_identity_name = sqlc.arg('UpdatedByIdentityName')
WHERE id = sqlc.arg('ID');
