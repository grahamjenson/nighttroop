-------
-- CREATE 
--------


--------
-- READ
--------

-- name: AllDeviceStatus :many
SELECT * FROM device_status ORDER BY occurred_at DESC;

-- name: GetDeviceStatusByEID :one
SELECT * FROM device_status WHERE eid = ?;

--------
-- UPDATE
--------


--------
-- DELETE
--------