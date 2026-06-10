-------
-- CREATE 
--------

-- name: CreateTempBasal :exec
INSERT INTO temp_basals (
        eid,
        start,
        duration,
        value,
        occurred_at,
        raw_json_data, 
        created_by_identity_id,
        created_by_identity_name
    )
VALUES ( ?, ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllTempBasals :many
SELECT * FROM temp_basals ORDER BY occurred_at DESC;


-- name: GetTempBasalByEID :one
SELECT * FROM temp_basals WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------
