-- name: SaveSummaryStale :exec
INSERT INTO summary_stale (at) VALUES (?) ON CONFLICT DO NOTHING;

-- name: GetSummaryStale :many
SELECT at FROM summary_stale ORDER BY at LIMIT 50000;

-- name: DeleteSummaryStale :exec
DELETE FROM summary_stale;
