--IN: id int
--OUT: id int, eid string, start *time.Time, duration *time.Duration, value float64
SELECT id, eid, start, duration, value FROM tempbasals WHERE id = ?;


