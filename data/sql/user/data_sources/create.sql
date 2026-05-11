--IN: Id int, Type string, URL string, Username string, Password string, CreatedAt *time.Time, LastAttemptedSync *time.Time, LastSuccessfulSync *time.Time, ConsecutiveFailedSyncs int, TotalFailedSyncs int, PreviousSyncError string
--OUT:
INSERT INTO datasources (id, type, url, username, password, created_at, last_attempted_sync, last_successful_sync, consecutive_failed_syncs, total_failed_syncs, previous_sync_error) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);