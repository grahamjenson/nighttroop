CREATE TABLE activities (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    name TEXT CHECK(length(name) <= 64),
    notes TEXT CHECK(length(notes) <= 1024),
    activity_type TEXT NOT NULL CHECK(length(activity_type) <= 32) default 'exercise',

    duration_minutes INTEGER,
    calories INTEGER,
    steps INTEGER,
    heart_rate INTEGER,

    -- Default Stuff
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
    note_id TEXT UNIQUE REFERENCES notes(id) ON DELETE RESTRICT,

    -- Default Stuff
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

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    reading REAL NOT NULL,
    reading_type TEXT NOT NULL CHECK(reading_type IN ('sgv', 'mbg')) DEFAULT 'sgv', -- sgv, mbg etc

    -- Default Stuff
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

CREATE TABLE notes (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),
    content TEXT CHECK(length(content) <= 1024),
    reason TEXT CHECK(length(reason) <= 64), 

    -- Default Stuff
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

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 1024), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 1024), -- { [high, low] }

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
    
    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- 
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {"value": "110", "time": "00:00"},
    targets_high_json TEXT CHECK(length(targets_high_json) <= 1024), -- { [high, low] }
    targets_low_json TEXT CHECK(length(targets_low_json) <= 1024), -- { [high, low] }
    
    -- Default Stuff
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

CREATE TABLE temp_basals (
    id TEXT PRIMARY KEY CHECK(length(id) <= 36),

    duration_minutes INTEGER NOT NULL,
    absolute REAL NOT NULL,
    percent REAL, -- This is optional, but absolute is not
    

    -- Default Stuff
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

