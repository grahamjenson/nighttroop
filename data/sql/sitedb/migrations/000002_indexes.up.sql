-- Every All*/Search* query orders by occurred_at DESC with a LIMIT, so an
-- index on it lets SQLite walk rows newest first and stop instead of scanning
-- and sorting the whole table (the Nightscout API dedups every create this way).
CREATE INDEX glucose_readings_occurred_at_idx ON glucose_readings(occurred_at);
CREATE INDEX insulin_doses_occurred_at_idx ON insulin_doses(occurred_at);
CREATE INDEX food_intakes_occurred_at_idx ON food_intakes(occurred_at);
CREATE INDEX activities_occurred_at_idx ON activities(occurred_at);
CREATE INDEX notification_events_occurred_at_idx ON notification_events(occurred_at);
CREATE INDEX temp_basals_occurred_at_idx ON temp_basals(occurred_at);
CREATE INDEX temp_targets_occurred_at_idx ON temp_targets(occurred_at);
CREATE INDEX profile_switches_occurred_at_idx ON profile_switches(occurred_at);
CREATE INDEX component_events_occurred_at_idx ON component_events(occurred_at);
CREATE INDEX event_groups_occurred_at_idx ON event_groups(occurred_at);
CREATE INDEX device_reports_occurred_at_idx ON device_reports(occurred_at);
CREATE INDEX profile_presets_occurred_at_idx ON profile_presets(occurred_at);
CREATE INDEX food_presets_created_at_idx ON food_presets(created_at);

-- LastModifiedCollections takes the newest updated_at of each collection
CREATE INDEX glucose_readings_updated_at_idx ON glucose_readings(updated_at);
CREATE INDEX activities_updated_at_idx ON activities(updated_at);
CREATE INDEX event_groups_updated_at_idx ON event_groups(updated_at);
CREATE INDEX device_reports_updated_at_idx ON device_reports(updated_at);
CREATE INDEX food_presets_updated_at_idx ON food_presets(updated_at);

-- Type lookups and deletes by type, a device's reports, and a profile store's presets
CREATE INDEX glucose_readings_reading_type_idx ON glucose_readings(reading_type);
CREATE INDEX event_groups_group_type_idx ON event_groups(group_type);
CREATE INDEX device_reports_device_occurred_at_idx ON device_reports(created_by_device_name, occurred_at);
CREATE INDEX profile_presets_profile_store_id_idx ON profile_presets(profile_store_id);
