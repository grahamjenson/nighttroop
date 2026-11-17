-- Each search's totals count the rows not deleted, and their earliest and latest occurred_at. With deleted_at in
-- occurred_at's index, SQLite counts from the index alone, rather than reading every row of the table: a device's
-- statuses are most of a looping site's DB.
DROP INDEX glucose_readings_occurred_at_idx;
CREATE INDEX glucose_readings_occurred_at_idx ON glucose_readings(occurred_at, deleted_at);
DROP INDEX insulin_doses_occurred_at_idx;
CREATE INDEX insulin_doses_occurred_at_idx ON insulin_doses(occurred_at, deleted_at);
DROP INDEX food_intakes_occurred_at_idx;
CREATE INDEX food_intakes_occurred_at_idx ON food_intakes(occurred_at, deleted_at);
DROP INDEX activities_occurred_at_idx;
CREATE INDEX activities_occurred_at_idx ON activities(occurred_at, deleted_at);
DROP INDEX notification_events_occurred_at_idx;
CREATE INDEX notification_events_occurred_at_idx ON notification_events(occurred_at, deleted_at);
DROP INDEX temp_basals_occurred_at_idx;
CREATE INDEX temp_basals_occurred_at_idx ON temp_basals(occurred_at, deleted_at);
DROP INDEX temp_targets_occurred_at_idx;
CREATE INDEX temp_targets_occurred_at_idx ON temp_targets(occurred_at, deleted_at);
DROP INDEX profile_switches_occurred_at_idx;
CREATE INDEX profile_switches_occurred_at_idx ON profile_switches(occurred_at, deleted_at);
DROP INDEX event_groups_occurred_at_idx;
CREATE INDEX event_groups_occurred_at_idx ON event_groups(occurred_at, deleted_at);
DROP INDEX device_reports_device_occurred_at_idx;
CREATE INDEX device_reports_device_occurred_at_idx ON device_reports(created_by_device_name, occurred_at, deleted_at);
