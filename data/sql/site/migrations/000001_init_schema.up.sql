---------------------------------------------------
-- PATIENT DATA Tables
-- The design of the tables are such that they should not need external references, like a spreadsheet
-- Standing alone they should be enough to calculate and predict values.
---------------------------------------------------

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value REAL NOT NULL,
    type TEXT NOT NULL CHECK(length(type) <= 32) DEFAULT 'sgv', -- sgv, finger, etc

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

CREATE TABLE insulin (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(length(type) <= 32) DEFAULT 'bolus', -- bolus, basal...
    dia_hours REAL NOT NULL default 6,
    onset_minutes REAL NOT NULL default 20,
    peak_minutes REAL NOT NULL default 60,

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

CREATE TABLE food (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    name TEXT CHECK(length(name) <= 64),

    carbs REAL NOT NULL DEFAULT 0,
    fat REAL NOT NULL DEFAULT 0,
    protein REAL NOT NULL DEFAULT 0,
    fibre REAL NOT NULL DEFAULT 0,
    calories REAL NOT NULL DEFAULT 0,
    
    glycemic_index REAL NOT NULL DEFAULT 100,
    onset_minutes REAL NOT NULL DEFAULT 20,
    absorption_minutes REAL NOT NULL DEFAULT 120,

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

CREATE TABLE temp_basals (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    start INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL,
    value REAL NOT NULL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE temp_targets (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    start INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL,
    high_target REAL NOT NULL,
    low_target REAL NOT NULL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);


-- Profiles are tricksy, they can have be set as a percent
CREATE TABLE profile (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT NOT NULL CHECK(length(name) <= 64),
    
    -- Date stuff
    start INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL default 0, -- e.g. 0 means it is forever unless another profile overrides it
    
    percent INTEGER,

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


-- This reports device status each device like cgm, pump phone, loop, mums phone, school phone, pixel can report its status
CREATE TABLE device_status (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    type TEXT NOT NULL CHECK(length(type) <= 32), -- pump, cgm, loop, phone

    -- Specific Details about the device, especially the PWD device
    battery_percentage REAL,
    charging_state TEXT CHECK(length(charging_state) <= 32),
    status TEXT CHECK(length(status) <= 32), -- This is the status of the device, like "ok", "error", "warning", etc

    insertion_time INTEGER, -- This is sometimes recorded as an event which is stuper annoying
    timestamp INTEGER, -- This is the timestamp of the device status, like the time of the last reading

    -- loop
    iob REAL,
    iob_basal REAL,
    iob_treatment REAL,
    
    cob REAL,

    -- pump specific values
    pump_reservoir REAL,

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

CREATE TABLE activity (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT CHECK(length(name) <= 64),
    type TEXT CHECK(length(type) <= 32),

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


CREATE TABLE note (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    content TEXT CHECK(length(content) <= 1024) NOT NULL,

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
-- Treatment is a group of other types, joined together so that they can be treated as a single event, e.g. a bolus wizard having a glucose reading and a bolus
----------

CREATE TABLE treatment (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    type TEXT NOT NULL CHECK(length(type) <= 32),

    glucose_reading_id TEXT UNIQUE CHECK(length(glucose_reading_id) <= 32),
    insulin_id TEXT UNIQUE CHECK(length(insulin_id) <= 32),
    food_id TEXT UNIQUE CHECK(length(food_id) <= 32),
    temp_basal_id TEXT UNIQUE CHECK(length(temp_basal_id) <= 32),
    temp_target_id TEXT UNIQUE CHECK(length(temp_target_id) <= 32),
    profile_id TEXT UNIQUE CHECK(length(profile_id) <= 32),
    device_status TEXT UNIQUE CHECK(length(device_status) <= 32),
    activity TEXT UNIQUE CHECK(length(activity) <= 32),
    note_id TEXT UNIQUE CHECK(length(note_id) <= 32),

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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32),

    -- Contraints
    FOREIGN KEY (glucose_reading_id) REFERENCES glucose_readings(id),
    FOREIGN KEY (insulin_id) REFERENCES insulin(id),
    FOREIGN KEY (food_id) REFERENCES food_intakes(id),
    FOREIGN KEY (temp_basal_id) REFERENCES temp_basals(id),
    FOREIGN KEY (temp_target_id) REFERENCES temp_targets(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id),
    FOREIGN KEY (device_status) REFERENCES device_status(id),
    FOREIGN KEY (activity) REFERENCES activity(id),
    FOREIGN KEY (note_id) REFERENCES notes(id)
);

----------
-- Presets are user defined templates for the above types, and can be used to create new instances of the above types
----------

CREATE TABLE presets_food (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 64),
    category TEXT NOT NULL CHECK(length(category) <= 32) DEFAULT 'meal',

    carbs REAL NOT NULL DEFAULT 0,
    fat REAL NOT NULL DEFAULT 0,
    protein REAL NOT NULL DEFAULT 0,
    calories REAL NOT NULL DEFAULT 0,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

CREATE TABLE presets_activity (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 64),
    category TEXT NOT NULL CHECK(length(category) <= 32) DEFAULT 'exercise',

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

-- e.g. this is just a list of available profiles and user settings to choose from
CREATE TABLE presets_profile (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 64),

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- { [1,2] },
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {[1,2]},
    targets_json TEXT CHECK(length(targets_json) <= 1024), -- { [high, low] }
    
    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly

    deleted_at INTEGER,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    created_by_app_name TEXT NOT NULL CHECK(length(created_by_app_name) <= 32),
    created_by_device_name TEXT NOT NULL CHECK(length(created_by_device_name)  <= 32),
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);

-- Maybe "preset" is the wrong word here, more like settings, but I like the reuse of terminology
-- These also should not belong on profile.
CREATE TABLE site_settings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    -- SGVs
    units TEXT NOT NULL CHECK (units IN ('mmol/L', 'mg/dL')) DEFAULT 'mmol/L', -- mg/dL, mmol/L

    -- DATE
    timezone TEXT NOT NULL CHECK(length(timezone) <= 32) DEFAULT 'UTC',
    
    -- insulin stuff
    bolus_insulin_name TEXT NOT NULL CHECK(length(bolus_insulin_name) <= 32) DEFAULT 'novorapid',
    bolus_dia_hours REAL NOT NULL default 6,
    bolus_onset_minutes REAL NOT NULL default 20,
    bolus_peak_minutes REAL NOT NULL default 60,

    -- If basal insulin is different from bolus insulin
    basal_insulin_name TEXT NOT NULL CHECK(length(basal_insulin_name) <= 32) DEFAULT 'lantus',
    basal_dia_hours REAL NOT NULL default 24,
    basal_onset_minutes REAL NOT NULL default 60,
    basal_peak_minutes REAL NOT NULL default 360,

    -- Default Stuff
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    updated_by_identity_name TEXT CHECK(length(updated_by_identity_name) <= 32),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32)
);