-- Input MaxLogRows which will delete the oldest rows
--IN: max_log_rows int
--OUT:
DELETE FROM events WHERE id < (SELECT MAX(id) - ? FROM events)