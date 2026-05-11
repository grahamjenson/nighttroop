--IN: id int
--OUT: id int, eid string, created_by_type string, created_by_id int, value float64, created_at *time.Time
SELECT id, eid, created_by_type, created_by_id, value, created_at FROM entries WHERE id = ?;
