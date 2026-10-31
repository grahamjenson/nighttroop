-- name: CreateResourceReading :exec
INSERT INTO resource_readings (at, cpu_percent, cpu_pegged_seconds, memory_percent, disk_percent)
VALUES (sqlc.arg('At'), sqlc.arg('CPUPercent'), sqlc.arg('CPUPeggedSeconds'), sqlc.arg('MemoryPercent'), sqlc.arg('DiskPercent'));

-- name: LatestResourceReading :one
SELECT * FROM resource_readings ORDER BY at DESC LIMIT 1;

-- name: DeleteResourceReadingsBefore :exec
DELETE FROM resource_readings WHERE at < sqlc.arg('Before');

-- name: ResourceReadingsSince :many
-- Oldest first, for a graph
SELECT * FROM resource_readings WHERE at >= sqlc.arg('Since') ORDER BY at;
