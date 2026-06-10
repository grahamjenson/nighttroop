




-------
-- CREATE 
--------


-- name: CreateTempTarget :exec
INSERT INTO temp_targets (
    eid, 
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

-- name: GetTempTargetByEID :one
SELECT * FROM temp_targets WHERE eid = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
