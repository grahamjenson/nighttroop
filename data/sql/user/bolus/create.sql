--IN: id int, eid string, value string, type string, created_at *time.Time
--OUT:
INSERT INTO bolus (id, eid, value, type, created_at) VALUES (?, ?, ?, ?, ?);