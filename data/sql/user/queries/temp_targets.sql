




-------
-- CREATE 
--------


-- name: CreateTempTarget :exec
INSERT INTO temp_targets (
    id, 
    start, 
    duration, 
    highTarget, 
    lowTarget,
    occurred_at,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
    ) 
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?);


--------
-- READ
--------


-- name: AllTempTargets :many
SELECT * FROM temp_targets ORDER BY occurred_at DESC;

-- name: GetTempTarget :one
SELECT * FROM temp_targets WHERE id = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
