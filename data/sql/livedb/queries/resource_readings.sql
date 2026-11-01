-- name: CreateResourceReading :exec
INSERT INTO resource_readings (at, cpu_percent, cpu_pegged_seconds, memory_percent, memory_total_bytes, disk_percent, disk_total_bytes)
VALUES (sqlc.arg('At'), sqlc.arg('CPUPercent'), sqlc.arg('CPUPeggedSeconds'), sqlc.arg('MemoryPercent'), sqlc.arg('MemoryTotalBytes'),
    sqlc.arg('DiskPercent'), sqlc.arg('DiskTotalBytes'));

-- name: LatestResourceReading :one
SELECT * FROM resource_readings ORDER BY at DESC LIMIT 1;

-- name: DeleteResourceReadingsBefore :exec
DELETE FROM resource_readings WHERE at < sqlc.arg('Before');

-- name: ResourceReadingsSince :many
-- Oldest first, for a graph
SELECT * FROM resource_readings WHERE at >= sqlc.arg('Since') ORDER BY at;
