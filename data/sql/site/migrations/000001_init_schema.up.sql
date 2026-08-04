---------------------------------------------------
-- SITE PATIENT DATA Tables
---------------------------------------------------

-----------
-- EVENTS (they happen at a time on a time line)
-----------

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    reading REAL NOT NULL,
    reading_type TEXT NOT NULL CHECK(length(reading_type) <= 32) DEFAULT 'sgv', -- sgv, finger, etc

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE insulin_doses (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    dose REAL NOT NULL,
    dose_type TEXT CHECK (dose_type IN ('bolus', 'basal')) DEFAULT 'bolus',

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE food_intakes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    name TEXT CHECK(length(name) <= 64), -- this could be a reference to a preset
    food_type TEXT CHECK(length(food_type) <= 32), -- this could be like hypo treatment, meal, snack, etc

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    calories REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


CREATE TABLE activities (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT CHECK(length(name) <= 64),
    notes TEXT CHECK(length(notes) <= 1024),
    activity_type TEXT CHECK(length(activity_type) <= 32),

    duration_minutes INTEGER,
    calories INTEGER,
    steps INTEGER,
    heart_rate INTEGER,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


CREATE TABLE notes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    content TEXT CHECK(length(content) <= 1024) NOT NULL,
    reason TEXT CHECK(length(reason) <= 64), 

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

-- temp_basals, temp_targets and profile_switches are all ranges over a time period so start is occured_at until duration (unless another overide happens)
-- activity also has a duration
CREATE TABLE temp_basals (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    duration_minutes INTEGER NOT NULL,
    absolute REAL NOT NULL,
    percent REAL, -- This is optional, but absolute is not
    

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE temp_targets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    duration_minutes INTEGER NOT NULL,
    high_target REAL NOT NULL,
    low_target REAL NOT NULL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


-- Profile Switches are tricksy, they have a percent, but the values stored in here already contain that percent
CREATE TABLE profile_switches (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    -- Date stuff
    duration_minutes INTEGER NOT NULL default 0, -- e.g. 0 means it is forever unless another profile overrides it

    -- ratios
    percent REAL NOT NULL default 100, 

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- { [1,2] },
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {[1,2]},
    targets_json TEXT CHECK(length(targets_json) <= 1024), -- { [high, low] }
    
    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


-- STATE TABLES (recording states of things over time)

CREATE TABLE pump_states (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    pump_state TEXT CHECK(length(pump_state) <= 64), -- This is the state of the device, like "ok", "error", "warning", etc
    pump_timestamp INTEGER, -- This is the timestamp of the device state, like the time of the last reading

    pump_insertion_time INTEGER,

    pump_reservoir REAL,

    pump_battery_percent REAL,
    pump_charging_state TEXT CHECK(length(pump_charging_state) <= 32),

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE cgm_states (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    cgm_state TEXT CHECK(length(cgm_state) <= 64), -- This is the state of the device, like "ok", "error", "warning", "insertion" etc
    cgm_timestamp INTEGER, -- This is the timestamp of the device state, like the time of the last reading

    cgm_insertion_time INTEGER,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE loop_states (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    loop_state TEXT CHECK(length(loop_state) <= 64), -- This is the state of the device, like "ok", "error", "warning", etc
    loop_timestamp INTEGER, -- This is the timestamp of the device state, like the time of the last reading

    iob REAL,
    iob_basal REAL,
    iob_treatment REAL,

    cob REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE device_states (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    device_state TEXT CHECK(length(device_state) <= 64), -- This is the state of the device, like "ok", "error", "warning", etc
    device_timestamp INTEGER, -- This is the timestamp of the device state, like the time of the last reading

    device_battery_percent REAL,
    device_charging_state TEXT CHECK(length(device_charging_state) <= 32),

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);



-- EventGroup is when multiple events happen together, like a meal with insulin and glucose reading
CREATE TABLE event_groups (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    group_type TEXT NOT NULL CHECK(length(group_type) > 0 AND length(group_type) <= 32),

    glucose_reading_id TEXT UNIQUE REFERENCES glucose_readings(id) ON DELETE RESTRICT,
    insulin_dose_id TEXT UNIQUE REFERENCES insulin_doses(id) ON DELETE RESTRICT,
    food_intake_id TEXT UNIQUE REFERENCES food_intakes(id) ON DELETE RESTRICT,
    temp_basal_id TEXT UNIQUE REFERENCES temp_basals(id) ON DELETE RESTRICT,
    temp_target_id TEXT UNIQUE REFERENCES temp_targets(id) ON DELETE RESTRICT,
    profile_switch_id TEXT UNIQUE REFERENCES profile_switch(id) ON DELETE RESTRICT,

    device_state_id TEXT UNIQUE REFERENCES device_states(id) ON DELETE RESTRICT,
    loop_state_id TEXT UNIQUE REFERENCES loop_states(id) ON DELETE RESTRICT,
    pump_state_id TEXT UNIQUE REFERENCES pump_states(id) ON DELETE RESTRICT,
    cgm_state_id TEXT UNIQUE REFERENCES cgm_states(id) ON DELETE RESTRICT,

    activity_id TEXT UNIQUE REFERENCES activities(id) ON DELETE RESTRICT,
    note_id TEXT UNIQUE REFERENCES notes(id) ON DELETE RESTRICT,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

----------
-- PRESETS are user defined templates for the event types
----------

CREATE TABLE food_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 64),
    food_type TEXT NOT NULL CHECK(length(food_type) <= 32),

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    calories REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

-- e.g. this is just a list of available profiles and user settings to choose from
CREATE TABLE profile_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 64),

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- { [1,2] },
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {[1,2]},
    targets_json TEXT CHECK(length(targets_json) <= 1024), -- { [high, low] }
    
    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


----------
-- SETTINGS is a singleton defined settings used to configure site behavior
----------

CREATE TABLE site_settings (
    id INTEGER PRIMARY KEY,

    -- SGVs
    units TEXT NOT NULL CHECK (units IN ('mmol/L', 'mg/dL')) DEFAULT 'mmol/L', -- mg/dL, mmol/L

    -- DATE
    timezone TEXT NOT NULL CHECK(length(timezone) <= 32) DEFAULT 'UTC',

    -- INSULIN
    bolus_dia_hours REAL NOT NULL default 6,
    bolus_onset_minutes REAL NOT NULL default 20,
    bolus_peak_minutes REAL NOT NULL default 60,

    basal_dia_hours REAL NOT NULL default 24,
    basal_onset_minutes REAL NOT NULL default 60,
    basal_peak_minutes REAL NOT NULL default 120,

    -- CARBS
    carb_max_absorption_hours REAL NOT NULL default 6,
    carb_min_5m_impact REAL NOT NULL default 0,

    -- Default Stuff
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32)
);

INSERT INTO site_settings (id) VALUES (1);


