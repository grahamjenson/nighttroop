---------------------------------------------------
-- TABLES FOR USING NIGHTTROOP
---------------------------------------------------
CREATE TABLE settings (
    key TEXT PRIMARY KEY CHECK(length(key) <= 64),
    value JSON NOT NULL CHECK(length(value) <= 8192),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
);

-- This table is like users and devices, as like a sugarpixel isnt a user, it is a device.
CREATE TABLE identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE CHECK(length(name) <= 32),
    password TEXT NOT NULL CHECK(length(password) <= 128),
    totp_secret TEXT NOT NULL CHECK(length(totp_secret) <= 128),
    role TEXT NOT NULL CHECK(length(role) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    last_seen_at INTEGER,
    last_write_at INTEGER,
    deleted_at INTEGER, -- soft delete here, but on delete will wipe out password and such
    enabled INTEGER NOT NULL DEFAULT 1
);


-- Events table primarily used to show user issues and results of events like syncing errors or login attempts, stuff like that.
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT CHECK(length(type) <= 32),
    details TEXT CHECK(length(details) <= 1024),
    start_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    end_at INTEGER
);

---------------------------------------------------
-- TABLES ABOUT PATIENT DATA
---------------------------------------------------

-- This reports device status above
CREATE TABLE device_status (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    
    identity_id INTEGER NOT NULL,
    

    -- Specific Details about the device, especially the PWD device
    battery_level INTEGER,
    charging_state TEXT CHECK(length(charging_state) <= 32),

    -- pump sub device
    pump_battery INTEGER,
    pump_status TEXT CHECK(length(pump_status) <= 32),
    pump_reservoir REAL,
    pump_insertion_time INTEGER, -- This is sometimes recorded as an event which is stuper annoying

    -- loop
    loop_status TEXT CHECK(length(loop_status) <= 32),
    loop_timestamp INTEGER,

    -- cgm
    cgm_status TEXT CHECK(length(cgm_status) <= 32),
    cgm_insertion_time INTEGER, -- This is sometimes recorded as an event which is stuper annoying

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 8192), -- This is the largest user derived string, and should be culled quickly
    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL  -- id of identity that created this record
);

CREATE TABLE glucose_readings (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value REAL NOT NULL,
    type TEXT NOT NULL  CHECK(length(type) <= 32) DEFAULT 'sgv', -- sgv, finger, etc

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly
    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record


);

CREATE TABLE bolus (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    value REAL NOT NULL,
    type TEXT NOT NULL CHECK(length(type) <= 32) DEFAULT 'bolus',

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly
    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record

);

CREATE TABLE carbs (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(length(type) <= 32) DEFAULT 'carbs',

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly
    deleted_at INTEGER,
    utc_offset_minutes INTEGER DEFAULT 0,
    occurred_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record

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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record

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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
);



CREATE TABLE profile (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT CHECK(length(name) <= 64),
    
    -- Date stuff
    start INTEGER,
    duration_minutes INTEGER,
    
    source TEXT CHECK(length(source) <= 64),
    percent INTEGER,
    dia INTEGER,
    units TEXT CHECK(length(units) <= 32),
    timezone TEXT CHECK(length(timezone) <= 32),

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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
);


--
-- Stuff To make NS compatable (but not core)
--

CREATE TABLE activity (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT CHECK(length(name) <= 64),
    type TEXT CHECK(length(type) <= 32),
    notes TEXT CHECK(length(notes) <= 1024),
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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
);

CREATE TABLE food (
    id TEXT PRIMARY KEY CHECK(length(id) <= 32),

    name TEXT CHECK(length(name) <= 64),
    carbs REAL,
    calories REAL,

    -- Default Stuff
    raw_json_data TEXT CHECK(length(raw_json_data) <= 1024), -- This is the largest user derived string, and should be culled quickly
    deleted_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
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
    created_by_identity_name TEXT NOT NULL CHECK(length(created_by_identity_name) <= 32), -- denormalized name for easier lookup (also survives delete)
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
);