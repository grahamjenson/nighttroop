-------
-- CREATE 
--------
-- name: CreateNote :exec
INSERT INTO note (
        eid,
        occurred_at,
        content,
        raw_json_data,
        created_by_identity_id,
        created_by_identity_name
) VALUES ( ?, ?, ?, ?, ?, ?);

--------
-- READ
--------

-- name: AllNotes :many
SELECT * FROM note ORDER BY occurred_at DESC; 

-- name: GetNoteByEID :one
SELECT * FROM note WHERE eid = ?;

--------
-- UPDATE
--------

--------
-- UPDATE
--------


--------
-- DELETE
--------
