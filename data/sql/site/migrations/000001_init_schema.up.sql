---------------------------------------------------
-- SITE PATIENT DATA Tables
---------------------------------------------------

-----------
-- EVENTS (they happen at a time on a time line)
-----------

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    mgdl REAL NOT NULL,
    reading_type TEXT NOT NULL CHECK(reading_type IN ('sgv', 'mbg')) DEFAULT 'sgv', -- sgv, mbg etc
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L -- the reading is always store in mgdl, but the original unit CAN also be stored
    measurement_method TEXT CHECK(length(measurement_method) <= 32), -- cgm, meter, manual, etc

    delta REAL,
    direction TEXT CHECK(length(direction) <= 16),
    noise REAL,
    filtered REAL,
    unfiltered REAL,
    rssi REAL,


    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW glucose_readings_vw AS SELECT glucose_readings.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN glucose_readings ON TRUE;


CREATE TABLE insulin_doses (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    dose REAL NOT NULL,
    dose_type TEXT CHECK(length(dose_type) <= 32),
    
    split_now REAL,
    split_ext REAL,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW insulin_doses_vw AS SELECT insulin_doses.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN insulin_doses ON TRUE;

CREATE TABLE food_intakes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT CHECK(length(name) <= 64), -- this could be a reference to a preset

    food_type TEXT CHECK(length(food_type) <= 32), -- Type of food
    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),

    absorption_time_minutes INTEGER,
    pre_bolus REAL,

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    energy REAL,
    portions REAL,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW food_intakes_vw AS SELECT food_intakes.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN food_intakes ON TRUE;

CREATE TABLE activities (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    name TEXT CHECK(length(name) <= 64),
    activity_type TEXT NOT NULL CHECK(length(activity_type) <= 32) default 'exercise',

    duration_minutes INTEGER,
    calories INTEGER,
    steps INTEGER,
    heart_rate INTEGER,


    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW activities_vw AS SELECT activities.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN activities ON TRUE;

CREATE TABLE notification_events (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    notification_type TEXT CHECK(length(notification_type) <= 32) NOT NULL DEFAULT 'announcement', -- note, warning, error, critical, info
    title TEXT CHECK(length(title) <= 128) NOT NULL,
    severity INTEGER CHECK(severity >= 0 AND severity <= 5) NOT NULL DEFAULT 0, -- 0 is lowest, 5 is highest
    content TEXT CHECK(length(content) <= 1024),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW notification_events_vw AS SELECT notification_events.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN notification_events ON TRUE;

-- temp_basals, temp_targets and profile_switches are all ranges over a time period so start is occured_at until duration (unless another overide happens)
-- activity also has a duration
CREATE TABLE temp_basals (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL,
    absolute REAL NOT NULL,
    percent REAL, -- This is optional, but absolute is not
    
    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW temp_basals_vw AS SELECT temp_basals.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN temp_basals ON TRUE;

CREATE TABLE temp_targets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL,
    high_target REAL NOT NULL,
    low_target REAL NOT NULL,
    insulin_needs_scale_factor REAL,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW temp_targets_vw AS SELECT temp_targets.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN temp_targets ON TRUE;

-- Profile Switches are tricksy, they have a percent, but the values stored in here already contain that percent
CREATE TABLE profile_switches (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    -- Profile Switch Specific
    duration_minutes INTEGER NOT NULL default 0, -- e.g. 0 means it is forever unless another profile overrides it
    percent REAL default 100, 
    timeshift INTEGER default 0, -- in minutes, e.g. 60 means the profile is shifted forward by 1 hour

    -- Profile Stuff
    dia REAL, -- in hours
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L used for targets
    timezone TEXT CHECK(length(timezone) <= 32),
    
    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048                     AND json_valid(carb_ratios_json)), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048 AND json_valid(insulin_sensitivities_json)), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048                     AND json_valid(basal_rates_json)), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048                   AND json_valid(targets_high_json)), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048                     AND json_valid(targets_low_json)), -- { [high, low] }
    
    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW profile_switches_vw AS SELECT profile_switches.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN profile_switches ON TRUE;

CREATE TABLE component_events (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    event_type TEXT NOT NULL CHECK(length(event_type) <= 32), -- same event type as event_group

    sensor_code TEXT CHECK(length(sensor_code) <= 64),
    transmitter_id TEXT CHECK(length(transmitter_id) <= 64),

    pump_id TEXT CHECK(length(pump_id) <= 32),
    pump_serial TEXT CHECK(length(pump_serial) <= 32),
    pump_type TEXT CHECK(length(pump_type) <= 64),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

CREATE VIEW component_events_vw AS SELECT component_events.* FROM (SELECT NULL WHERE FALSE) dummy FULL JOIN component_events ON TRUE;


-- EventGroup is when multiple events happen together, like a meal with insulin and glucose reading
CREATE TABLE event_groups (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    group_type TEXT NOT NULL CHECK(length(group_type) > 0 AND length(group_type) <= 32),

    glucose_reading_id TEXT UNIQUE REFERENCES glucose_readings(id) ON DELETE RESTRICT,
    insulin_dose_id TEXT UNIQUE REFERENCES insulin_doses(id) ON DELETE RESTRICT,
    food_intake_id TEXT UNIQUE REFERENCES food_intakes(id) ON DELETE RESTRICT,
    temp_basal_id TEXT UNIQUE REFERENCES temp_basals(id) ON DELETE RESTRICT,
    temp_target_id TEXT UNIQUE REFERENCES temp_targets(id) ON DELETE RESTRICT,
    profile_switch_id TEXT UNIQUE REFERENCES profile_switches(id) ON DELETE RESTRICT,

    component_event_id TEXT UNIQUE REFERENCES component_events(id) ON DELETE RESTRICT,

    activity_id TEXT UNIQUE REFERENCES activities(id) ON DELETE RESTRICT,
    notification_event_id TEXT UNIQUE REFERENCES notification_events(id) ON DELETE RESTRICT,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    duration_minutes INTEGER, -- this is needed for notes, and is common enough
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);


------------
-- Reports Reports are not events, they are a snapshot of the state of a device at a point in time
------------
CREATE TABLE device_reports (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- uploader
    -- device is created_by_device_name
    device_battery_percent REAL,
    is_device_charging INTEGER DEFAULT 0,

    -- PUMP
    pump_state TEXT CHECK(length(pump_state) <= 64),
    pump_reservoir REAL,
    pump_battery_percent REAL,
    pump_charging_state TEXT CHECK(length(pump_charging_state) <= 32),

    -- Raw Loop Settings
    loop_name TEXT CHECK( loop_name IN ('openaps', 'loop')),
    configuration_settings_json TEXT CHECK(length(configuration_settings_json) <= 8192 AND json_valid(configuration_settings_json)),
    loop_settings_json TEXT CHECK(length(loop_settings_json) <= 8192 AND json_valid(loop_settings_json)), -- some loops (e.g. Loop) store settings on the profile

    -- Extracted Loop Settings that will be useful for reporting and analysis
    loop_iob REAL,
    loop_iob_basal REAL,
    loop_iob_treatment REAL,
    loop_cob REAL,

    loop_predictions_start_time INTEGER,
    loop_predictions_default TEXT CHECK(length(loop_predictions_default) <= 2048 AND json_valid(loop_predictions_default)), -- iob for openaps, normal for loop
    loop_predictions_zt TEXT CHECK(length(loop_predictions_zt) <= 2048 AND json_valid(loop_predictions_zt)),
    
    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);


----------
-- PRESETS are user defined templates for the event types
----------

CREATE TABLE food_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),

    food_type TEXT CHECK(length(food_type) <= 32),     
    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),

    absorption_time_minutes INTEGER,
    pre_bolus REAL,

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    energy REAL,
    portions REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64)
);

-- e.g. this is just a list of available profiles and user settings to choose from
CREATE TABLE profile_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),

    -- Store stuff
    profile_store_id TEXT NOT NULL,
    is_default INTEGER NOT NULL DEFAULT 0, -- 0 for false, 1 for true

    -- Profile Stuff
    dia REAL, -- in hours
    units TEXT CHECK(units IN ('mmol/L', 'mg/dL')), -- mg/dL, mmol/L
    timezone TEXT CHECK(length(timezone) <= 32), -- e.g. UTC, America/New_York, etc

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048), -- { [high, low] }

    -- loop settings
    loop_settings_json TEXT CHECK(length(loop_settings_json) <= 2048 AND json_valid(loop_settings_json)), -- some loops (e.g. Loop) store settings on the profile

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192  AND json_valid(raw_json_data)), -- This is the largest user derived string, and should be culled quickly

    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 64),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 64),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 64),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 64),

    -- Constraints
    UNIQUE (name, profile_store_id)
);

