-------
-- CREATE 
--------
-- name: CreateNote :exec
INSERT INTO note (
        id,
        occurred_at,
        utc_offset_minutes,
        content,
        raw_json_data,
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllNotes :many
SELECT * FROM note ORDER BY occurred_at DESC; 

-- name: GetNote :one
SELECT * FROM note WHERE id = ?;

--------
-- UPDATE
--------

--------
-- UPDATE
--------


--------
-- DELETE
--------
