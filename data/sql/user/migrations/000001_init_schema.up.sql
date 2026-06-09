---------------------------------------------------
-- TABLES FOR USING NIGHTTROOP
---------------------------------------------------
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


    enabled INTEGER NOT NULL DEFAULT 1
);

-- This reports device status above
CREATE TABLE identity_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT CHECK(length(eid) <= 32),
    timestamp INTEGER NOT NULL,
    identity_id INTEGER NOT NULL,
    battery_level INTEGER,
    pump_battery INTEGER,
    pump_status TEXT CHECK(length(pump_status) <= 32),
    pump_reservoir REAL,
    loop_status TEXT CHECK(length(loop_status) <= 32),
    loop_timestamp INTEGER,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
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

CREATE TABLE sgvs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- The id used extenally to this system (like remote NS id, dexcom id, Loop id), a means to not duplicate records from a source
    eid TEXT UNIQUE NOT NULL CHECK(length(eid) <= 32), 
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value REAL NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by_identity_id INTEGER NOT NULL -- id of identity that created this record
);

CREATE TABLE bolus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT CHECK(length(eid) <= 32),
    value REAL,
    type TEXT CHECK(length(type) <= 32),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE carbs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT CHECK(length(eid) <= 32),
    amount REAL NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temp_basals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    eid TEXT CHECK(length(eid) <= 32),
    start INTEGER,
    duration INTEGER,
    value REAL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temp_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    eid TEXT CHECK(length(eid) <= 32),
    start INTEGER,
    duration INTEGER,
    highTarget INTEGER,
    lowTarget INTEGER,
    -- CONTRAINTS
    UNIQUE(eid)
);



CREATE TABLE profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT CHECK(length(eid) <= 32),
    storeid TEXT CHECK(length(storeid) <= 64),
    name TEXT CHECK(length(name) <= 64),
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    -- Date stuff
    START INTEGER,
    duration INTEGER,
    
    source TEXT CHECK(length(source) <= 64),
    percent INTEGER,
    dia INTEGER,
    units TEXT CHECK(length(units) <= 32),
    timezone TEXT CHECK(length(timezone) <= 32),

    carb_ratios_json TEXT CHECK(length(carb_ratios_json) <= 1024), -- { [1,2] },
    insulin_sensitivities_json TEXT CHECK(length(insulin_sensitivities_json) <= 1024), -- { [1,2] },
    basal_rates_json TEXT CHECK(length(basal_rates_json) <= 1024), -- {[1,2]},
    targets_json TEXT CHECK(length(targets_json) <= 1024), -- { [high, low] }
    -- CONTRAINTS
    UNIQUE(eid)
);



