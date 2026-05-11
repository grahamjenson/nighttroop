-- Input MaxLogRows which will delete the oldest rows
--IN: MaxLogRows int
--OUT:
DELETE FROM events WHERE id < (SELECT MAX(id) - ? FROM events)