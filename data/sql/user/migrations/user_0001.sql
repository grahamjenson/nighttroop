---------------------------------------------------
-- TABLES FOR USING NIGHTTROOP
---------------------------------------------------
CREATE TABLE devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    totp_secret TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP
);

-- This reports device status above
CREATE TABLE devicestatus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    deviceBattery INTEGER,
    pumpBattery INTEGER,
    pumpStatus TEXT,
    pumpReservoir REAL,
    loopStatus TEXT,
    loopTimestamp INTEGER,
    created_at INTEGER DEFAULT (unixepoch())
);

-- Data souces like dexcom/nightscout (only one per type), has real deletes for saftey
CREATE TABLE data_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Nightscout/Dexcom/Device (Device being a default one, which can be referenced from device)
    type TEXT UNIQUE NOT NULL,
    url TEXT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Rather than having a seperate table data_source_sync_report or something, we will simplify by just having this here
    last_attempted_sync INTEGER,
    last_successful_sync INTEGER,
    consecutive_failed_syncs INTEGER DEFAULT 0,
    total_failed_syncs INTEGER,
    previous_sync_error TEXT -- stored to show to user as feedback

);

---------------------------------------------------
-- TABLES ABOUT PATIENT DATA
---------------------------------------------------

CREATE TABLE sgvs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- The id used extenally to this system (like remote NS id, dexcom id) maybe preceeded by the source like (dexcom-123) to make unique
    eid TEXT, 

    created_by_type TEXT, -- "device", "data_source"
    created_by_id INTEGER, -- id of either device or data_source that created this entry
    
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value INTEGER,
    timestamp INTEGER,

    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE bolus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    value REAL,
    TYPE TEXT,
    timestamp INTEGER,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE carbs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tempbasal (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    START INTEGER,
    duration INTEGER,
    value REAL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temptargets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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



---------------------------------------------------
-- LOGS
-- hooks into the slog logging system and writes certain events here 
-- systematically truncated "DELETE FROM $logsTableName WHERE id < (SELECT MAX(id) - $maxLogRows FROM $logsTableName)"
---------------------------------------------------

CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
     level TEXT,
     -- context is a JSON value about where the log occured, e.g. what background job, what user...
     context TEXT,
     message TEXT,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    details TEXT,
    start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end TIMESTAMP
);