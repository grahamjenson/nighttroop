-------
-- CREATE 
--------

-- name: CreateTempBasal :exec
INSERT INTO temp_basals (
        id,
        start,
        duration_minutes,
        value,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
    )
VALUES ( ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllTempBasals :many
SELECT * FROM temp_basals ORDER BY start DESC;


-- name: GetTempBasal :one
SELECT * FROM temp_basals WHERE id = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
