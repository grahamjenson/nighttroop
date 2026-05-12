CREATE TABLE admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    totp_secret TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP
);

CREATE TABLE data_source_block_list(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT, -- in case it is a dexcom account or something
    url TEXT,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    details TEXT,
    start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end TIMESTAMP
);

CREATE TABLE sqlite_sequence(name,seq);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subdomain TEXT UNIQUE NOT NULL,
    batch TEXT GENERATED ALWAYS AS (SUBSTR(subdomain, 1, 1)) VIRTUAL NOT NULL,
    
    backup_key_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    max_db_size_kb_limit INTEGER DEFAULT 3000, -- default 3Mb, can be changed in control plane
    

    -- last_seen_at is default null because on creation they have not been seen, will be deleted quickly is not logged in sooner
    last_seen_at TIMESTAMP,
    last_write_at TIMESTAMP,
    deleted_at TIMESTAMP
);

