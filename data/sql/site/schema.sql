CREATE TABLE bolus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    value REAL,
    type TEXT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE carbs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    amount REAL NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    details TEXT,
    start_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    end_at INTEGER
);

CREATE TABLE identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    totp_secret TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    last_seen_at INTEGER,
    enabled INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE identity_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    timestamp INTEGER NOT NULL,
    identity_id INTEGER NOT NULL,
    battery_level INTEGER,
    pump_battery INTEGER,
    pump_status TEXT,
    pump_reservoir REAL,
    loop_status TEXT,
    loop_timestamp INTEGER,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER))
);

CREATE TABLE profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eid TEXT,
    storeid TEXT,
    name TEXT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    -- Date stuff
    START INTEGER,
    timestamp INTEGER GENERATED ALWAYS AS (START) VIRTUAL,
    duration INTEGER,
    END INTEGER GENERATED ALWAYS AS (START + (duration * 1000)) VIRTUAL,
    starttime TEXT GENERATED ALWAYS AS (datetime(START / 1000, 'unixepoch')) VIRTUAL,
    endtime TEXT GENERATED ALWAYS AS (datetime(END / 1000, 'unixepoch')) VIRTUAL,
    
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

CREATE TABLE sgvs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- The id used extenally to this system (like remote NS id, dexcom id) maybe preceeded by the source like (dexcom-123) to make unique
    eid TEXT, 

    created_by_type TEXT, -- "device", "data_source"
    created_by_id INTEGER, -- id of either device or data_source that created this entry
    
    
    -- UNIT IS ALWAYS MG/DL BECAUSE IT IS HIGHER RESOLUTION
    value REAL NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE sqlite_sequence(name,seq);

CREATE TABLE temp_basals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    eid TEXT,
    START INTEGER,
    duration INTEGER,
    value REAL,
    -- CONTRAINTS
    UNIQUE(eid)
);

CREATE TABLE temp_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    eid TEXT,
    START INTEGER,
    duration INTEGER,
    highTarget INTEGER,
    lowTarget INTEGER,
    -- CONTRAINTS
    UNIQUE(eid)
);

