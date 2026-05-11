--IN: id int
--OUT: id int, eid string, timestamp *time.Time, device_id int, device_battery int, pump_battery int, pump_status string, pump_reservoir float64, loop_status string, loop_timestamp *time.Time, created_at *time.Time
SELECT id, eid, timestamp, device_id, deviceBattery, pumpBattery, pumpStatus, pumpReservoir, loopStatus, loopTimestamp, created_at FROM devicestatus WHERE id = ?;