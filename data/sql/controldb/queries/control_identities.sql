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
SELECT * FROM control_identities ORDER BY created_at DESC;


-- name: GetControlIdentity :one
SELECT * FROM control_identities WHERE name = sqlc.arg('Name') LIMIT 1;
-- name: GetControlIdentityByID :one
SELECT * FROM control_identities WHERE id = sqlc.arg('ID');
