--IN: Id int, Eid string, Value string, Type string, CreatedAt *time.Time
--OUT:
INSERT INTO bolus (id, eid, value, type, created_at) VALUES (?, ?, ?, ?, ?);