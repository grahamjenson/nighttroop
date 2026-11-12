-- name: PutSettingsDocument :one
-- Replaces the document, or adds it while the site has fewer than Max (no row otherwise). A replaced one's updated_at
-- moves on at least 1ms, so it was created only when updated_at is created_at.
INSERT INTO settings_documents (id, document_json, occurred_at, utc_offset_minutes, created_by_identity_name)
SELECT sqlc.arg('ID'), sqlc.arg('DocumentJson'), sqlc.arg('OccurredAt'), sqlc.arg('UtcOffsetMinutes'), sqlc.arg('CreatedByIdentityName')
WHERE (SELECT COUNT(*) FROM settings_documents) < CAST(sqlc.arg('Max') AS INTEGER)
    OR EXISTS (SELECT 1 FROM settings_documents WHERE id = sqlc.arg('ID'))
ON CONFLICT(id) DO UPDATE SET
    document_json = excluded.document_json,
    occurred_at = excluded.occurred_at,
    utc_offset_minutes = excluded.utc_offset_minutes,
    updated_at = MAX(CAST(unixepoch('subsec') * 1000 AS INTEGER), settings_documents.updated_at + 1)
RETURNING *;

-- name: GetSettingsDocument :one
SELECT * FROM settings_documents WHERE id = ?;

-- name: AllSettingsDocuments :many
SELECT * FROM settings_documents ORDER BY updated_at DESC, id LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: DeleteSettingsDocument :one
DELETE FROM settings_documents WHERE id = ? RETURNING *;
