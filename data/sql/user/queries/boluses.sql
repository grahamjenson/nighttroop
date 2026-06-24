-------
-- CREATE 
--------
-- name: CreateBolus :exec
INSERT INTO bolus (
        id,
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

-- name: GetBolus :one
SELECT * FROM bolus WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
