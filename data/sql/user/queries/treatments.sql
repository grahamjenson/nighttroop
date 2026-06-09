-------
-- BOLUSES
---------

-- name: AllBoluses :many
SELECT * FROM bolus ORDER BY created_at DESC; 

-- name: CreateBolus :exec
INSERT INTO bolus (eid, value, type) VALUES ( ?, ?, ?);



-------
-- CARBS
-------

-- name: AllCarbs :many
SELECT * FROM carbs ORDER BY created_at DESC;


-- name: CreateCarb :exec
INSERT INTO carbs (eid, amount) VALUES ( ?, ?);

------
-- Temp Targets
-------

-- name: AllTempTargets :many
SELECT * FROM temp_targets ORDER BY created_at DESC;

-- name: CreateTempTarget :exec
INSERT INTO temp_targets (eid, start, duration, highTarget, lowTarget) VALUES ( ?, ?, ?, ?, ?);

------
-- temp basal
-------

-- name: AllTempBasals :many
SELECT * FROM temp_basals ORDER BY created_at DESC;

-- name: CreateTempBasal :exec
INSERT INTO temp_basals (eid, start, duration, value) VALUES ( ?, ?, ?, ?);