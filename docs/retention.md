# Retention: hot rows and cold days

**Status:** a proposal, for discussion. The server does not use it. `data/retention` has a small prototype with
tests. This replaces the tiers of `docs/truncation.md` with one general system.

## The problem, measured

A test wrote 7 days of a looping site into a site DB: a CGM, two phones that report, a temp basal each 5 minutes
and an automatic bolus each 10 minutes. SQLite's `dbstat` gave the size of each table, with its indexes.

| Table | Rows a day | Bytes a row | KB a day |
| --- | --- | --- | --- |
| `device_reports` (before the hourly thinning) | 576 | 406 | 229 |
| `event_groups` | 440 | 442 | 190 |
| `glucose_readings` | 290 | 222 | 63 |
| `temp_basals` | 288 | 189 | 53 |
| `insulin_doses` | 152 | 192 | 29 |
| All tables | | | **595** |

The test's loop status is small. A real loop's status is up to 8 KB for each report, so a real site writes more
than this before the thinning.

What the numbers say:

- A looping site writes about 0.6 MB a day. The 20 MB limit is full in about 5 weeks, less what the device report
  thinning takes back.
- **A row costs about 200 bytes, whatever is in it.** A glucose reading is 2 bytes of information. Its ID, its
  times, its names and its 3 indexes are the other 220.
- **A treatment costs two rows.** The `event_groups` row of a temp basal (442 bytes, 12 indexes) is larger than the
  temp basal (189 bytes).
- Thus to delete columns or thin rows is not enough. The cost is the row.

## The idea

A site's data has two forms.

| Form | What it is | For |
| --- | --- | --- |
| **Hot rows** | The tables of today: one row for each record, as the app sent it, with indexes | The last days: what apps sync, and what a person edits |
| **Cold days** | One row for each kind of data for each UTC day | All days before that: what a person looks back at |

To **seal** a day is to change its hot rows into cold days, and then delete the rows. One job does it, for all
tables, from one table of policies.

The result, with the prototype's numbers:

| | Hot rows | Cold day |
| --- | --- | --- |
| A day of CGM glucose | 61 KB | 747 bytes |
| A day of a looping site | about 370 KB, without device reports | about 5 KB |
| A year of a looping site | 135 MB | about 1.8 MB |

## Five rules

1. **Each table has a shape.** The shape is the form its rows take in a cold day, and says what sealing does. A table with no policy fails a test
   (`TestEveryTableHasAPolicy`), so a new table cannot be added without a decision.
2. **A cold day is the truth, not a summary.** Reports, statistics and charts are built from cold days and can be
   built again at any time. There is no "frozen" summary that cannot be made again.
3. **A row is deleted only after its cold day is written and checked** (see [Sealing](#sealing)).
4. **One reader.** Pages, reports and the API ask one function for a series or for the events of a time range. It
   reads hot rows, cold days, or both. No caller knows where the border is.
5. **Each policy says what is lost**, in words a person can read. The site shows these words.

## Shapes

| Shape | What it is | Sealing does | Tables |
| --- | --- | --- | --- |
| **Signal** | A number a machine writes on a clock | Folds the rows into series on a grid. Deletes the rows. | `glucose_readings`, `temp_basals` |
| **Event** | What a person did or decided, a few each day | Moves the rows into the day's event list, with the columns the policy keeps. Folds numbers into series. | `insulin_doses`, `food_intakes`, `activities`, `temp_targets`, `profile_switches`, `notification_events`, `event_groups` |
| **Telemetry** | What a device says about itself | Folds its numbers into series. Deletes the rows. | `device_reports` |
| **Dictionary** | A value many records repeat | Kept one time, by its hash, while a record refers to it | `loop_settings`, and each profile of a profile switch |
| **Config** | A list a person manages | Never sealed | `food_presets`, `profile_presets`, `general_settings` |
| **Derived** | Built from the other tables | A cache. Built again when needed. | `summary_period` |

`data/retention/policy.go` has the policy of each table: its shape, its series, the columns its events keep, and
what is lost.

## Cold days

One table holds all cold data:

```sql
CREATE TABLE cold_days (
    kind TEXT NOT NULL CHECK(length(kind) <= 32),  -- a series ('glucose'), 'events', or 'seal'
    date TEXT NOT NULL CHECK(length(date) = 10),   -- the UTC day
    body_json TEXT NOT NULL CHECK(length(body_json) <= 65536 AND json_valid(body_json)),
    PRIMARY KEY (kind, date)
) STRICT, WITHOUT ROWID;
```

- The key is `(kind, date)`, so a year of one series is 365 rows that are side by side on disk.
- **A cold day is a UTC day.** Each day has the same steps. A change of the site's timezone, or of the clocks,
  changes nothing that is stored. A local day is read from the two UTC days it is in.

### A series

A series is one number through the day on a grid: one value for each step, or none.

| Series | From | Step | Kept to | A step holds |
| --- | --- | --- | --- | --- |
| `glucose` | `glucose_readings` (CGM) | 5 min | 1 mg/dL | The newest reading of the best source |
| `iob`, `iob_basal`, `cob` | `device_reports` | 5 min | 0.01 U, 0.1 g | The newest report |
| `insulin_automatic`, `insulin_manual` | `insulin_doses` | 1 hour | 0.001 U | The total |
| `basal_delivered`, `basal_suspended_minutes` | `temp_basals` and the profile | 1 hour | 0.001 U | The total |
| `carbs` | `food_intakes` | 1 hour | 0.1 g | The total |
| `pump_reservoir`, `pump_battery`, `phone_battery` | `device_reports` | 1 hour | | The newest report |

`PackGrid` (`data/retention/grid.go`) packs a day as JSON: `{"v":1,"d":[140,2,-1,0,null,3,...]}`. `d` has one
entry for each step: `null` for no value, or the difference from the value before. Glucose changes a few mg/dL in
5 minutes, so most entries are 1 or 2 characters.

| A day of glucose (276 readings, an hour with none) | Bytes |
| --- | --- |
| As rows, with indexes | 61,272 |
| JSON, each value (`[140,142,141,...]`) | 1,122 |
| **JSON, differences (proposed)** | **747** |
| Packed bytes, not JSON | 316 |

- It is JSON, so the column has `json_valid`, as each JSON column has, and a person can read a day with a query.
- Packed bytes are half the size, but cannot be checked or read. JSON is the proposal. See Open question 8.
- SQLite's JSONB is not smaller here: it keeps a number as its digits.
- A new series is one line in the policy table. It needs no column and no migration.

### The day's events

`kind = 'events'` is the list of the day's events: meals, manual boluses, finger pricks, notes, temp targets,
profile switches, site and sensor changes. Each event keeps its ID, its time, its values, its notes and who made it.
A treatment and its `event_groups` row become one entry.

- About 120 bytes for each event, not 630 (the row and its group, with indexes).
- Automatic records (temp basals, automatic boluses) are not in the list. Their totals are in the hourly series.

### Statistics

`Stats` (`data/retention/grid.go`) is count, sum, sum of squares, minimum, maximum and a count for each 5 mg/dL.

- Two `Stats` merge into one that is exact. A week is the merge of its days, a year the merge of its weeks.
- The counts give the time in **any** range. If a person changes their target range, old periods are right for
  the new range. `summary_period` stores `time_in_range` for the range of the day it was built, and cannot do this.
- A day's `Stats` comes from its glucose grid in microseconds, so it is not stored. A month can store its `Stats`
  as a cache, so a chart of 10 years reads 120 rows.

## Sealing

A job seals the oldest day that is not sealed and is older than the site's **hot days**. For each day, in one
transaction:

1. Read the day's hot rows.
2. Build each series and the event list from the policies.
3. If cold days are there already (late data, see below), merge: a step of a series takes the new value, an event
   with the same ID replaces the old one.
4. **Check.** The cold day must agree with the rows: the same total insulin, the same total carbs, the same
   count of 5-minute steps with glucose. If one check fails, delete nothing and record an error.
5. Write the cold days, and a `seal` row with the counts of rows folded and deleted for each table.
6. Delete the rows, and the day's soft deleted rows.

- The job can stop after each day and continue later.
- The `seal` rows are the record of what was done. A page shows them: "14 March: 288 readings became 1 glucose
  line. 9 events kept."

### Late data and edits

| Case | What happens |
| --- | --- |
| An app sends a record for a sealed day (a backfill, a new connector with 2 years of history) | It is written as a hot row, as today. The next run seals the day again and merges it. |
| A person edits or deletes an event of a sealed day | The day's event list is written again with the change. |
| A person deletes a CGM reading of a sealed day | Its step of the glucose series is cleared. |
| A backfill of years of history | Its rows are sealed in the same job. A connector could write cold days directly later. |

Thus a cold day is never "frozen". It is only compact.

## Hot days

One number says how many days are hot rows: the env var `RETENTION_HOT_DAYS` (default 14,
`retention.DefaultHotDays`). A site over its size limit has fewer.

| Site | Hot days | Why |
| --- | --- | --- |
| Default | 14 | More than the sync window of each app we know, and a person edits recent days |
| Over its size limit | Fewer, step by step | The limit holds without deleting information |
| Floor | 3 | Loops read hours; followers read 1 to 2 days |

- Device reports keep their own, shorter schedule (as today): as sent for a day, then thinner, then only series.
- A site could have its own number later (Open question 9).

## What apps get

Apps read through the Nightscout API. Most read the last hours or days, which are hot rows.

| Request | Hot days | Sealed days |
| --- | --- | --- |
| `entries` (CGM) | The rows | **Made from the glucose series:** one entry for each 5-minute step, with an ID made from its time |
| `treatments`: meals, manual boluses, notes, changes | The rows | Made from the day's event list, with their own IDs |
| `treatments`: temp basals, automatic boluses | The rows | One for each hour from the hourly series, or none (to decide) |
| `devicestatus` | The rows | None |
| API v3 `history` (changes since a time) | As today | Sealing is not a change, so nothing is sent again |
| A write or a delete by ID | As today | An event: found in its day's list. A CGM entry: by its time. |

- A tool that reads months (Nightscout Reporter, a clinic's export) gets every 5-minute reading and every meal and
  bolus, as today. It does not get each temp basal.
- A request that makes entries from many sealed days costs CPU. It needs a limit for each request, and the
  existing rate limits.

## What a person sees

- **Reports** (AGP, time in range, daily dose): the same, for any age, and now right for a changed target range.
- **Day view of an old day:** the glucose line, IOB and COB, hourly insulin, basal and carbs, and each meal, bolus
  and note. Not each temp basal or automatic bolus.
- **Data pages:** hot rows as today. Older days show the event list, and a note for what is in series.
- **Download Data:** the cold days as CSV files, one for each series, and the events.
- **A "Storage" page:** what each policy keeps and loses, and the `seal` records.

## What is lost

| What | When | Why it is acceptable |
| --- | --- | --- |
| A CGM reading's exact second, raw signal, and app | When its day is sealed | The 5-minute value is what all reports and standards use |
| The readings of a second source of the same glucose | When sealed | One reading for each 5 minutes is kept |
| Each temp basal and automatic bolus | When sealed | The hourly totals are exact |
| The loop's reasons and predictions | After 1 day (as today) | Only useful to look at the last hours |
| Each device report | By its schedule (as today) | IOB, COB, reservoir and batteries are series |
| The app's own IDs of a treatment | When sealed | No app we know uploads again after 14 days |
| Which app and device made an event | When sealed | The identity is kept |

Nothing a person entered is lost. See Open question 1 for a way to lose nothing at all.

## Size of a site

A looping site, with the measured numbers:

| Part | Size |
| --- | --- |
| 14 hot days at about 370 KB (without device reports) | about 5 MB |
| Device reports, thinned as today | about 2 MB |
| Cold days, each year (about 5 KB a day: 4 series each 5 minutes at 0.75 KB, hourly series, events) | about 1.8 MB |
| **After 1 year** | **about 9 MB** |
| **After 7 years** | **about 20 MB** |

- Today the same site is 20 MB after about 5 weeks.
- A site with a CGM and no loop is about 1 MB hot and 0.7 MB for each year.
- Without `iob_basal` (Open question 5), a looping site is about 1.5 MB for each year.
- `summary_period` is about 2.5 MB a year today (lists of 288 numbers as JSON). It would be larger than the cold
  data it is built from. It must use packed grids, or become a small cache.

## Hot rows can be cheaper too

Sealing does not help the first 14 days. Two changes, each on its own, make hot rows smaller:

1. **No `event_groups` row for an automatic record.** A temp basal and an automatic bolus with no notes carry
   nothing a person wrote. Their group row is 190 KB a day, a third of the site. The few fields an app needs back
   (its ID, `entered_by`) can be columns of the record.
2. **Fewer indexes on `event_groups`.** It has 12. Each partial unique index on a member's ID could be one index on
   one `member_id` column.

## Build order

Each step is one PR, and the site works after each one. Sealing is off until an env var turns it on.

| Step | What | Reads change |
| --- | --- | --- |
| 1 | `cold_days`, `PackGrid`, the seal job for **glucose only**, the checks, the `seal` record | The reader for glucose; summaries use it |
| 2 | The Nightscout API makes `entries` from sealed days | `entries` |
| 3 | Device reports fold into `iob`, `cob`, reservoir and battery series | The day view |
| 4 | Events: the day's list, and the hourly insulin, basal and carbs series | Data pages, `treatments`, daily dose |
| 5 | Hot days by size pressure; the Storage page; Download Data as CSV | |
| 6 | `summary_period` becomes a cache of packed grids and `Stats`; `frozen_at` goes | Reports |
| 7 | Cheaper hot rows (the two changes above) | |

Step 1 alone takes glucose from 63 KB a day to 0.75 KB, with the least risk: the glucose line of a day is already
what `summary_period` stores.

## Open questions

1. **Lose nothing: an archive.** Before a day's rows are deleted, write them, as sent, to a compressed file for
   each month in the site's S3 folder. "Sealed" then means "not in the site DB", not "gone", and a month could be
   read back. It needs S3, storage for each site, and a way to restore. Is this wanted, as an option of a site?
2. **Hot days:** is 14 right? Is 3 a safe floor for each app (AAPS, Trio, Loop, xDrip+)?
3. **Temp basals to apps:** for a sealed day, send one temp basal for each hour, or none?
4. **Hourly or 15 minutes** for insulin and basal? 15 minutes is 4 times the bytes of a small number (about 150
   bytes more a day) and shows how the loop acted.
5. **`iob_basal`:** keep it, or only total IOB? It is about 770 bytes a day.
6. **Profiles:** store each distinct profile one time by hash (a dictionary), and a switch refers to it. This
   needs one new column or table. Is the change wanted before the DB is frozen?
7. **Events for 20 years:** 10 events a day are 0.4 MB a year. Is that acceptable for ever, or do events older
   than N years become counts?
8. **JSON or packed bytes** for a series? JSON (proposed) is checked and readable, at about 1.8 MB a year. Packed
   bytes are about 1.1 MB a year, and need an exception to the rule for JSON columns.
9. **A site's own setting:** can a person choose "keep more days as rows" if their DB has space?
