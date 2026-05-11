--IN: id int, type string, url string, username string, password string, created_at *time.Time, last_attempted_sync *time.Time, last_successful_sync *time.Time, consecutive_failed_syncs int, total_failed_syncs int, previous_sync_error string
--OUT:
INSERT INTO datasources (id, type, url, username, password, created_at, last_attempted_sync, last_successful_sync, consecutive_failed_syncs, total_failed_syncs, previous_sync_error) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);