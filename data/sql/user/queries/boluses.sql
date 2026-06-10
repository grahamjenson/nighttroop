-------
-- CREATE 
--------
-- name: CreateBolus :exec
INSERT INTO bolus (
        eid,
        occurred_at,
        value,
        type,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllBoluses :many
SELECT * FROM bolus ORDER BY occurred_at DESC; 

-- name: GetBolusByEID :one
SELECT * FROM bolus WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
