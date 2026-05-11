--IN: id int
--OUT: id int, eid string, start *time.Time, duration *time.Duration, high_target int, low_target int
SELECT id, eid, start, duration, high_target, low_target FROM temptargets WHERE id = ?;