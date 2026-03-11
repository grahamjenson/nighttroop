CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    TYPE TEXT,
    details TEXT,
    start INTEGER,
    end INTEGER,

    timestamp INTEGER GENERATED ALWAYS AS (START) VIRTUAL,
    starttime TEXT GENERATED ALWAYS AS (datetime(START / 1000, 'unixepoch')) VIRTUAL,
    endtime TEXT GENERATED ALWAYS AS (datetime(END / 1000, 'unixepoch')) VIRTUAL
);

