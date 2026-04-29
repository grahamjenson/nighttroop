-- Input MaxLogRows which will delete the oldest rows
DELETE FROM events WHERE id < (SELECT MAX(id) - ? FROM events)