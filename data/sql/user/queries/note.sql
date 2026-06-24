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

-- name: UpdateNote :one
UPDATE note SET
    occurred_at = COALESCE(sqlc.narg(occurred_at), occurred_at),
    content = COALESCE(sqlc.narg(content), content),
    utc_offset_minutes = COALESCE(sqlc.narg(utc_offset_minutes), utc_offset_minutes),
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
WHERE id = ?
RETURNING *; 

--------
-- UPDATE
--------


--------
-- DELETE
--------
