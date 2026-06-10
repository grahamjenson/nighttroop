-------
-- CREATE 
--------



--------
-- READ
--------

-- name: AllProfiles :many
SELECT * FROM profile ORDER BY occurred_at DESC;


-- name: GetProfileByEID :one
SELECT * FROM profile WHERE eid = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
