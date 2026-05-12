-------
-- BOLUSES
---------

-- name: AllBoluses :many
SELECT * FROM bolus ORDER BY created_at DESC; 

-- name: CreateBolus :exec
INSERT INTO bolus (eid, value, type, created_at) VALUES ( ?, ?, ?, ?);

-------
-- ENTRIES
-------

-- name: AllEntries :many
SELECT * FROM entries ORDER BY created_at DESC;

-- name: CreateEntry :exec
INSERT INTO entries (eid, created_by_type, created_by_id, value) VALUES ( ?, ?, ?, ?);


-------
-- CARBS
-------

-- name: AllCarbs :many
SELECT * FROM carbs ORDER BY created_at DESC;


-- name: CreateCarb :exec
INSERT INTO carbs (eid, amount) VALUES ( ?, ?);


-------
-- DATA SOURCES
-------

-- name: AllDataSources :many
SELECT * FROM data_sources ORDER BY created_at DESC;

-- name: CreateDataSource :exec
INSERT INTO data_sources (id, type, url, username, password, created_at, last_attempted_sync, last_successful_sync, consecutive_failed_syncs, total_failed_syncs, previous_sync_error) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

------
-- DEVICES
-------

-- name: AllDevices :many
SELECT * FROM devices ORDER BY created_at DESC;

-- name: CreateDevice :exec
INSERT INTO devices (name, password, totp_secret, role) VALUES (?, ?, ?, ?);

-- name: GetDevice :one
SELECT * FROM devices where name = ? LIMIT 1;

-- name: UpdateDeviceLastSeenAt :exec
UPDATE devices SET last_seen_at = CURRENT_TIMESTAMP WHERE name = ?;

------
-- Devicestatus
-------

-- name: AllDeviceStatus :many
SELECT * FROM device_status ORDER BY created_at DESC;

----
-- Events
-------
-- name: AllEvents :many
SELECT * FROM events ORDER BY start_at DESC;

-- name: CreateEvent :exec
INSERT INTO events (type, details, start_at, end_at) VALUES (?, ?, ?, ?);

------
-- Profile
-------

-- name: AllProfiles :many
SELECT * FROM profile ORDER BY created_at DESC;


------
-- Temp Targets
-------

-- name: AllTempTargets :many
SELECT * FROM temp_targets ORDER BY created_at DESC;

------
-- temp basal
-------

-- name: AllTempBasals :many
SELECT * FROM temp_basals ORDER BY created_at DESC;




