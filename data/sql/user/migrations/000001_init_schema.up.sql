---------------------------------------------------
-- TABLES FOR USING NIGHTTROOP
---------------------------------------------------
CREATE TABLE devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    totp_secret TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('admin', 'pwd_phone', 'read')), -- Something like Read/Carer/Full/Admin/LoopPhone
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP
);

-- This reports device status above
CREATE TABLE device_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    timestamp INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    deviceBattery INTEGER,
    pumpBattery INTEGER,
    pumpStatus TEXT,
    pumpReservoir REAL,
    loopStatus TEXT,
    loopTimestamp INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Data souces like dexcom/nightscout (only one per type), has real deletes for saftey
CREATE TABLE sync_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Nightscout/Dexcom
    type TEXT UNIQUE NOT NULL CHECK(type IN ('NIGHTSCOUT', 'DEXCOM')),
    password TEXT NOT NULL, -- or ns token
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    enabled INTEGER DEFAULT 1, -- 0 or 1 for now, can be used to disable without deleting

    -- NIGHTSCOUT
    url TEXT, 
    hostname TEXT,  -- this is not easily dont in SQL

    -- DEXCOM
    username TEXT,
    dexcom_region TEXT NOT NULL CHECK(dexcom_region IN ('USA', 'OTHER')) DEFAULT 'OTHER', -- USA or OTHER
    
    sync_items TEXT, -- comma seperated list of what data to sync, empty or null means all, values like "treatments,entries"...
    frequency INTEGER NOT NULL DEFAULT 60, -- how often to sync in seconds

    -- Rather than having a seperate table data_source_sync_report or something, we will simplify by just having this here
    last_attempted_sync DATETIME,
    last_successful_sync DATETIME,
    
    consecutive_failed_syncs INTEGER DEFAULT 0,
    total_failed_syncs INTEGER DEFAULT 0,
    previous_sync_error TEXT, -- stored to show to user as feedback

    success_ema REAL, -- https://en.wikipedia.org/wiki/Exponential_smoothing
    failure_ema REAL,
    total_ema REAL,
    last_sync_duration INTEGER -- 1 for success, 0 for failure, null for never tried, used for quick checks without having to look at timestamps 
);

-- Events table primarily used to show user issues and results of events like syncing errors or login attempts, stuff like that.
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    details TEXT,
    start_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_at TIMESTAMP
);

---------------------------------------------------
-- TABLES ABOUT PATIENT DATA
---------------------------------------------------

CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- The id used extenally to this system (like remote NS id, dexcom id) maybe preceeded by the source like (dexcom-123) to make unique
    eid TEXT, 

    created_by_type TEXT, -- "device", "data_source"
    created_by_id INTEGER, -- id of either device or data_source that created this entry
    
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value REAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE bolus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    value REAL,
    type TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE carbs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    amount REAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temp_basals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    eid TEXT,
    START INTEGER,
    duration INTEGER,
    value REAL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temp_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    eid TEXT,
    START INTEGER,
    duration INTEGER,
    highTarget INTEGER,
    lowTarget INTEGER,
    -- CONTRAINTS
    UNIQUE(eid)
);



CREATE TABLE profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    storeid TEXT,
    name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- Date stuff
    START INTEGER,
    timestamp INTEGER GENERATED ALWAYS AS (START) VIRTUAL,
    duration INTEGER,
END INTEGER GENERATED ALWAYS AS (START + (duration * 1000)) VIRTUAL,
starttime TEXT GENERATED ALWAYS AS (datetime(START / 1000, 'unixepoch')) VIRTUAL,
endtime TEXT GENERATED ALWAYS AS (datetime(END / 1000, 'unixepoch')
) VIRTUAL,
source TEXT,
percent INTEGER,
dia INTEGER,
units TEXT,
timezone TEXT,

carb_ratios_json TEXT, -- { [1,2] },
insulin_sensitivities_json TEXT -- { [1,2] },
basal_rates_json TEXT -- {[1,2]},
targets_json TEXT, -- { [high, low] }
-- CONTRAINTS
UNIQUE(eid)
);



