---------------------------------------------------
-- SITE PATIENT DATA Tables
---------------------------------------------------

-----------
-- EVENTS (they happen at a time on a time line)
-----------

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    reading REAL NOT NULL,
    reading_type TEXT NOT NULL CHECK(reading_type IN ('sgv', 'mbg')) DEFAULT 'sgv', -- sgv, mbg etc

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE insulin_doses (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    dose REAL NOT NULL,
    dose_type TEXT NOT NULL CHECK(dose_type IN ('bolus', 'basal')) DEFAULT 'bolus',


    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE food_intakes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT CHECK(length(name) <= 64), -- this could be a reference to a preset

    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    energy REAL,
    portions REAL,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


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
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


CREATE TABLE notification_events (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    notification_type TEXT CHECK(length(notification_type) <= 32) NOT NULL DEFAULT 'announcement', -- note, warning, error, critical, info
    title TEXT CHECK(length(title) <= 128) NOT NULL,
    severity INTEGER CHECK(severity >= 0 AND severity <= 5) NOT NULL DEFAULT 0, -- 0 is lowest, 5 is highest
    content TEXT CHECK(length(content) <= 1024),

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

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
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE temp_targets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL,
    high_target REAL NOT NULL,
    low_target REAL NOT NULL,

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


-- Profile Switches are tricksy, they have a percent, but the values stored in here already contain that percent
CREATE TABLE profile_switches (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    -- Profile Switch Specific
    duration_minutes INTEGER NOT NULL default 0, -- e.g. 0 means it is forever unless another profile overrides it
    percent REAL NOT NULL default 100, 

    -- Profile Stuff
    dia REAL, -- in hours
    units TEXT CHECK(length(units) <= 32), -- mg/dL, mmol/L
    timezone TEXT CHECK(length(timezone) <= 32),
    
    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048), -- { [high, low] }
    
    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE component_events (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    event_type TEXT NOT NULL CHECK(length(event_type) <= 32), -- same event type as event_group

    -- Default Stuff
    notes TEXT CHECK(length(notes) <= 1024),
    reason TEXT CHECK(length(reason) <= 64),
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);



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
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


------------
-- Reports Reports are not events, they are a snapshot of the state of a device at a point in time
------------
CREATE TABLE device_reports (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- uploader
    -- device is created_by_device_name
    device_battery_percent REAL,
    device_charging_state TEXT CHECK(length(device_charging_state) <= 32),

    -- PUMP
    pump_state TEXT CHECK(length(pump_state) <= 64),
    pump_reservoir REAL,
    pump_battery_percent REAL,
    pump_charging_state TEXT CHECK(length(pump_charging_state) <= 32),

    -- LOOP
    loop_iob REAL,
    loop_iob_basal REAL,
    loop_iob_treatment REAL,
    loop_cob REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),


    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


----------
-- PRESETS are user defined templates for the event types
----------

CREATE TABLE food_presets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    category TEXT CHECK(length(category) <= 64),
    subcategory TEXT CHECK(length(subcategory) <= 64),

    carbs REAL,
    fat REAL,
    protein REAL,
    fibre REAL,
    energy REAL,
    portions REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
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
    units TEXT CHECK(length(units) <= 32), -- mg/dL, mmol/L
    timezone TEXT CHECK(length(timezone) <= 32), -- e.g. UTC, America/New_York, etc

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 2048), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 2048), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 2048), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 2048), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 2048), -- { [high, low] }

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly

    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    -- Constraints
    UNIQUE (name, profile_store_id)
);


----------
-- SETTINGS is going to be one row per identity_name with a single default row for the site
-- I think you can update any setting with the id = "<user_identity_name>" and the default row is id = "default"
-- admin user can update the default settings (probably not via API though)
-- users can only see their own + default settings, as they will merge
----------

CREATE TABLE site_settings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    -- SGVs
    units TEXT CHECK (units IN ('mmol/L', 'mg/dL')) DEFAULT 'mmol/L', -- mg/dL, mmol/L

    -- DATE
    timezone TEXT CHECK(length(timezone) <= 32) DEFAULT 'UTC',
    language TEXT CHECK(length(language) <= 32) DEFAULT 'en',
    date_format TEXT CHECK(length(date_format) <= 32) DEFAULT 'YYYY-MM-DD',
    time_format TEXT CHECK(length(time_format) <= 32) DEFAULT 'HH:mm:ss',
    week_start_day TEXT CHECK(length(week_start_day) <= 32) DEFAULT 'monday',

    -- INSULIN
    bolus_dia_hours REAL DEFAULT 6,
    bolus_onset_minutes REAL DEFAULT 20,
    bolus_peak_minutes REAL DEFAULT 60,

    basal_dia_hours REAL DEFAULT 24,
    basal_onset_minutes REAL DEFAULT 60,
    basal_peak_minutes REAL DEFAULT 120,

    default_insulin_sensitivity REAL DEFAULT 10,
    default_carb_ratio REAL DEFAULT 10,
    default_basal_rate REAL DEFAULT 1,

    -- CARBS
    carb_max_absorption_hours REAL DEFAULT 6,
    carb_min_5m_impact REAL DEFAULT 0,

    -- ALERTS
    alarm_urgent_low INTEGER DEFAULT 56,

    alarm_low INTEGER DEFAULT 72,
    alarm_low_minutes INTEGER DEFAULT 5,
    alarm_low_predicted INTEGER DEFAULT 1,

    alarm_high INTEGER DEFAULT 180,
    alarm_high_minutes INTEGER DEFAULT 30,
    
    alarm_urgent_high INTEGER DEFAULT 270,

    alarm_fast_rise INTEGER DEFAULT 18,
    alarm_fast_rise_minutes INTEGER DEFAULT 15,

    alarm_fast_drop INTEGER DEFAULT 18,
    alarm_fast_drop_minutes INTEGER DEFAULT 15,

    alarm_no_data_minutes INTEGER DEFAULT 20,

    alarm_battery_low_percent INTEGER DEFAULT 20,
    alarm_battery_low_pwd_percent INTEGER DEFAULT 20,

    alarm_low_reservoir_units  INTEGER DEFAULT 25,

    -- NOTIFICATIONS
    notifications_enabled INTEGER DEFAULT 0,
    notify_urgent_low INTEGER DEFAULT 0,
    notify_low INTEGER DEFAULT 0,
    notify_high INTEGER DEFAULT 0,
    notify_urgent_high INTEGER DEFAULT 0,
    notify_fast_rise INTEGER DEFAULT 0,
    notify_fast_drop INTEGER DEFAULT 0,
    
    
    -- Default Stuff
    updated_at INTEGER DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
);

INSERT INTO site_settings (id) VALUES ("default");


