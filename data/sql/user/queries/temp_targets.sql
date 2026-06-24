




-------
-- CREATE 
--------


-- name: CreateTempTarget :exec
INSERT INTO temp_targets (
    id, 
    start,
    utc_offset_minutes, 
    duration_minutes, 
    high_target, 
    low_target,
    raw_json_data, 
    created_by_identity_id,
    created_by_identity_name
    ) 
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?);


--------
-- READ
--------


-- name: AllTempTargets :many
SELECT * FROM temp_targets ORDER BY start DESC;

-- name: GetTempTarget :one
SELECT * FROM temp_targets WHERE id = ?;
--------
-- UPDATE
--------


--------
-- DELETE
--------
