# Retention: hot rows and cold days

**Status:** steps 1 to 4d are built, but the delete of treatment rows (4e). The server packages the CGM glucose of each day, what a loop's phone reports
(IOB, COB, reservoir, batteries), the treatments (basal rate, insulin, carbs), and the day's events into
`cold_days`. The Nightscout API makes `entries` from packaged days. It deletes no row unless `RETENTION_DELETE_ROWS` is
`true`. The other steps are the plan. See [Build order](#build-order).

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
| **Hot rows** | The tables of today: one row for each record, as the app sent it, with indexes | The last 31 days (`RETENTION_HOT_DAYS`): what apps sync, and what a person edits |
| **Cold days** | One row for each kind of data for each UTC day, as packed bytes | All days: what a person looks back at |

To **package** a day is to write its cold days from its hot rows. A packaged day is not final: it is written again
when its rows change. (To **pack** is a different word here: it is how the values of a day become bytes.) A day is packaged after it ends, and again when
its rows change. The rows are deleted when they are older than the hot days. One job does it, for all tables,
from one table of policies.

| | Hot rows | Cold day |
| --- | --- | --- |
| A day of CGM glucose | 61 KB | 173 bytes (tested) |
| A day of a looping site | about 370 KB, without device reports | about 2 KB (estimate) |
| A year of a looping site | 135 MB | about 0.7 MB (estimate) |

## Five rules

1. **Each table has a shape.** The shape is the form its rows take in a cold day, and says what packaging does. A
   table with no policy fails a test (`TestEveryTableHasAPolicy`), so a new table needs a decision.
2. **A cold day is the truth, not a summary.** Reports, statistics and charts are built from cold days and can be
   built again at any time. There is no "frozen" summary that cannot be made again.
3. **A row is deleted only if its cold day was written after the row was last written.**
4. **One reader.** Pages, reports and the API ask one function for a series or for the events of a time range. It
   reads hot rows, cold days, or both. No caller knows where the border is.
5. **Each policy says what is lost**, in words a person can read. The site shows these words.

## Shapes

| Shape | What it is | Packaging does | Tables |
| --- | --- | --- | --- |
| `series` | A number on a clock: a CGM's glucose, a loop's IOB | Folds the rows into series on a grid | `glucose_readings`, `temp_basals`, `device_reports` |
| `events` | What a person did or decided, a few each day | Puts the rows into the day's event list, with the columns the policy keeps. Folds numbers into series. | `insulin_doses`, `food_intakes`, `activities`, `temp_targets`, `profile_switches`, `notification_events`, `event_groups` |
| `dictionary` | A value many records repeat | Kept one time, by its hash, while a record refers to it | `loop_settings` |
| `config` | A list a person manages | Never packaged | `food_presets`, `profile_presets`, `general_settings` |
| `derived` | Built from the other tables | A cache. Built again when needed. | `summary_period` |
| `cold` | The cold days | Already small | `cold_days` |

- A table can give two shapes. `glucose_readings` gives the `glucose` series, and its finger pricks and its
  readings with a note are events.
- `data/retention/policy.go` has the policy of each table: its shape, its series, the columns its events keep,
  and what is lost.

## Cold days

One table holds all cold data:

```sql
CREATE TABLE cold_days (
    kind TEXT NOT NULL CHECK(length(kind) <= 32),  -- a series ('glucose'), or 'events'
    date TEXT NOT NULL CHECK(length(date) = 10),   -- the UTC day
    body BLOB NOT NULL CHECK(length(body) <= 16384),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,                   -- when it was last packaged
    PRIMARY KEY (kind, date)
) STRICT;
```

- The key is `(kind, date)`, so a year of one series is 365 rows that are side by side on disk.
- **A cold day is a UTC day.** Each day has the same steps. A change of the site's timezone, or of the clocks,
  changes nothing that is stored. A local day is read from the two UTC days it is in.
- `body` is packed bytes, not JSON, to be as small as possible.

### The first bytes: version and shape

A cold day is kept for years, so each body says what it is.

| Byte | What | Values |
| --- | --- | --- |
| 1 | Version | `1` |
| 2 | Shape | `1` series, `2` events |
| 3 | How a series was packed | bit 1: change of the change; bit 2: DEFLATE |
| Then, for a series | Step in seconds, scale, number of steps | 3 varints |

- A body is read by what it says it is, not by what the policy is now. If the step of a series changes later,
  old days are still read with their own step.
- To change the form, add a version and read each old version as it was written. Never change what an old
  version means. A day gets the new version only when it is packaged again.
- A body with a version, a shape or a packing the code does not know is an error, not a day of wrong numbers.

### A series

A series is one number through the day on a grid: one value for each step, or none.

| Series | From | Step | Kept to | A step holds |
| --- | --- | --- | --- | --- |
| `glucose` | `glucose_readings` (CGM) | 5 min | 1 mg/dL | The newest reading of the best source |
| `iob`, `iob_basal`, `cob` | `device_reports` | 5 min | 0.01 U, 0.1 g | The newest report |
| `basal_rate` | The profile, its switches, and `temp_basals` | 5 min | 0.001 U/hr | The mean rate in force |
| `insulin_bolus` | `insulin_doses` | 5 min | 0.001 U | The total: a loop's and a person's boluses, an extended one as it is given |
| `insulin_basal` | `insulin_doses` of basal insulin | 5 min | 0.001 U | The total of long acting insulin injected |
| `carbs` | `food_intakes` | 5 min | 0.1 g | The total |
| `pump_reservoir`, `pump_battery`, `phone_battery` | `device_reports` | 1 hour | | The newest report |

Treatments are on the 5-minute grid, not each hour:

- A day in India (+5:30) or Nepal (+5:45) starts in the middle of a UTC hour. Its daily dose is exact from
  5-minute values.
- `basal_rate` keeps what the loop did, not only how much. The basal given in a day is the sum of its values
  divided by 12, with `insulin_basal`.
- A step with no insulin or carbs has no value, so these series are a few bytes on most days.
- The manual boluses are also events (step 4b), so a loop's boluses are `insulin_bolus` less those.

`PackGrid` (`data/retention/grid.go`) packs a day:

1. One bit for each step: it has a value, or not.
2. Each value as a whole number (value × scale), written as its change from the value before, or as the change
   of that change. Small numbers take 1 byte.
3. DEFLATE over the result.
4. It tries each way and keeps the smallest.

| A day of glucose (276 readings, an hour with none) | Bytes |
| --- | --- |
| As rows, with indexes | 61,272 |
| JSON, each value | 1,122 |
| JSON, changes | 747 |
| Packed changes | 313 |
| **Packed changes, DEFLATE (built)** | **173** |

A smooth series is smaller: a day of IOB is 107 bytes in the test. A new series is one line in the policy
table. It needs no column and no migration.

### The day's events

`kind = 'events'` is the list of the day's events: meals, manual boluses, finger pricks, notes, temp targets,
profile switches, site and sensor changes. Each event keeps its ID, its time, its values, its notes and who made it.
A treatment and its `event_groups` row become one entry.

- Automatic records (temp basals, a loop's boluses) are not in the list. They are in the series. A CGM reading is
  in the list only if it has a note.
- **Built** (`packageEventsDay`, `data/db/package_events.go`). The body has shape `2`: the events as a JSON list, with
  DEFLATE when that is smaller. JSON, as an event has other columns for each table.
- An event is `k`, its table, and each column of its row that its table's policy keeps (`Keep` in `policy.go`) and
  that has a value. So the policy table is what decides what is kept. A test checks each `Keep` column is real.
- The columns of an event's treatment (`group_type`, `entered_by`) are added where the event has none of that
  name, and `group_id`, the treatment's own ID, which an app knows it by.
- A day with more events than a cold day holds (16 KB packed, some thousands of events) is not packaged.
- **Not yet:** the delete of event rows (4d). So a day's events are built from its rows alone.

### What is in force when a day starts

A profile set in March is in force in June, and a temp target can go past midnight. If the rows of the days
before are gone, a day cannot look back for them. So a day's events have two lists (version 2 of the body):

```json
{"in_force":[{"k":"profile_switches","id":"…","name":"Sick day","basal_rates_json":"[…]"}],
 "events":[{"k":"food_intakes","id":"…","carbs":45}]}
```

- `in_force` is: the newest profile switch with no end; each newer profile switch with an end not yet reached;
  and the newest temp target, if it is not over.
- So each cold day is read alone.
- A day with no event of its own is packaged if something is in force. The last 2 days that ended are packaged in
  each run for this.
- Version 1 of the body, the list of events alone, is still read.

### A profile switch keeps its profile

A profile switch event has the profile it was made with: its basal rates, ratios, sensitivities, targets, DIA,
units and timezone. It is in the event, not in a table of its own.

- AAPS sends the same profile with each switch, so a day can have the same profile many times, and each day has
  it again in `in_force`. DEFLATE makes the copies in one day small. The copies in other days are not shared.
- This is simple: a day is read alone, and there is one table. The cost is about 0.3 to 0.5 KB a day for a site
  with a profile switch in force (an estimate).

### Statistics

`Stats` (`data/retention/grid.go`) is count, sum, sum of squares, minimum, maximum and a count for each 5 mg/dL.

- Two `Stats` merge into one that is exact. A week is the merge of its days, a year the merge of its weeks.
- The counts give the time in **any** range. If a person changes their target range, old periods are right for
  the new range. `summary_period` stores `time_in_range` for the range of the day it was built, and cannot do this.
- A day's `Stats` comes from its glucose series in microseconds, so it is not stored.

## Packaging

Packaging runs for each site before its backup, where the truncation runs (`packageSite`, `data/db/package.go`).

1. Find the UTC days that ended and had a row written, changed or deleted since the last run. A site that was
   never packaged has all its days.
2. For each day, read its rows and build each series (and, later, the event list) from the policies.
3. If the day's rows can be gone (it is older than the hot days), start from the cold day as it is: a row
   written since takes its step, and a deleted row clears its step. If not, build the day from its rows only.
4. **Check.** Read the new cold day back. It must be equal to what was built. If not, write nothing.
5. Write the cold day.
6. If `RETENTION_DELETE_ROWS` is `true`, delete the rows older than the hot days. Only a row whose cold day was
   written after the row was last written is deleted. Thus no row is deleted that its cold day does not have.

- Each day is one write, so a run can stop after a day and continue later.
- Reports will read series, not rows, for all days but today. Thus a day's series is written after the day
  ends, and its rows stay for the hot days.
- For glucose, a CGM reading with a note, or one a treatment refers to, is not deleted. It is an event.
- **Device reports are packaged each run, today too, before they are thinned.** They are thinned from 1 hour old, so
  a day that waits for its end has only each 10 minutes. Their day always starts from the cold day as it is, so
  it keeps each 5 minutes after the rows are thinned and deleted.
- **Treatments are built where the summaries are built** (`sitedata`), with the same code: the profile in force,
  the temp basals, and an extended bolus split over its time and over midnight. `data/db` cannot import that
  package, so it sets `db.TreatmentDaySeries` when the server starts.
- A day's treatment series are packaged when it ends, for the last 2 days, and again for a day with a treatment
  written. They are built from the rows alone, as no treatment row is deleted yet.
- The phone's battery is the loop's phone's. A carer's phone reports its own battery, which is not packaged.

### Late data and edits

| Case | What happens |
| --- | --- |
| An app sends a record for a packaged day (a backfill, a new connector with 2 years of history) | It is written as a hot row, as today. The next run packages the day again, with what the day had. |
| A person deletes a CGM reading of a packaged day | Its step of the glucose series is cleared. |
| A person edits or deletes an event of a packaged day (step 4) | The day's event list is written again with the change. |

Thus a cold day is never "frozen". It is only compact.

## Hot days

`RETENTION_HOT_DAYS` says how many days are hot rows (default 31, `retention.DefaultHotDays`).

| Site | Hot days | Why |
| --- | --- | --- |
| Default | 31 | A month of each record as sent: more than the sync window of each app we know |
| Over its size limit (step 6) | Fewer, step by step | The limit holds without deleting information |
| Floor | 3 | Loops read hours; followers read 1 to 2 days |

## What apps get

Apps read through the Nightscout API. Most read the last hours or days, which are hot rows.

| Request | Hot days | Packaged days |
| --- | --- | --- |
| `entries` (CGM) | The rows | **Made from the glucose series:** one entry for each 5-minute step, with an ID made from its time |
| `treatments`: meals, manual boluses, notes, changes | The rows | Made from the day's event list, with their own IDs |
| `treatments`: temp basals, automatic boluses | The rows | One for each hour from the totals, or none (to decide) |
| `devicestatus` | The rows | None |
| API v3 `history` (changes since a time) | As today | Packaging is not a change, so nothing is sent again |
| A write or a delete by ID | As today | An event: found in its day's list. A CGM entry: by its time. |

- A tool that reads months (Nightscout Reporter, a clinic's export) gets every 5-minute reading and every meal and
  bolus, as today. It does not get each temp basal.
- A request that makes entries from many packaged days costs CPU. It needs a limit for each request, and the
  existing rate limits.
**Built: `entries`.** `ColdGlucoseReadings` (`data/db/cold_glucose.go`) is the reader of packaged glucose. The
API's search uses it (`SearchWithCold`):

- Only if `RETENTION_DELETE_ROWS` is `true`, and only for the part of a search before the hot days.
- A made entry is at the start of its 5-minute step. Its `_id` is the 8 hex digits of its time and
  `c01dc01dc01dc01d`: the shape of a Nightscout ID, and the same each time.
- A step that still has a row (a reading with a note) is given as the row, not made again.
- A search by date, type or glucose, newest first, has made entries. A search by device, app, ID, or the time a
  record was written has rows only. So has a search sorted oldest first, or by `srvModified`.
- A made entry cannot be read, changed or deleted by its `_id` (404). To do, with the delete of a step.

**Built: `treatments`** (`treatments_cold.go`), by the same rules as `entries`:

- A made treatment is built from the cold events with its ID (`group_id`): a meal and its bolus are one treatment.
  It has its own `_id`, its type, its time and what its events kept.
- A search by date or by `eventType`, newest first, has made treatments. A search by another field (carbs,
  insulin, `enteredBy`, duration) has rows only.
- A loop's boluses and temp basals are not made: they are series, not events. To decide: make them from the
  series, one for each run of the same rate.
- No treatment row is deleted yet (step 4e), so the API reads rows for all days, and a made treatment shows only
  in a test that deletes rows.

`devicestatus` of packaged days is not made.

## What a person sees (when all steps are built)

- **Reports** (AGP, time in range, daily dose): the same, for any age, and right for a changed target range.
- **Day view of an old day:** the glucose line, IOB and COB, the insulin, basal and carbs totals, and each meal,
  bolus and note. Not each temp basal or automatic bolus.
- **Data pages:** hot rows as today. Older days show the event list, and a note for what is in series.
- **Download Data:** the cold days as CSV files, one for each series, and the events.
- **A "Storage" page:** what each policy keeps and loses.

## To look at the cold days

A site's admin opens **Cold Days**, below Usage. It lists each cold day with what its bytes say they are (version,
shape, step) and its size, and opens a day to show each value. It is read only. See `REQUIREMENTS.md`, "Cold Days
page".

## What is lost

| What | When | Why it is acceptable |
| --- | --- | --- |
| A CGM reading's exact second, raw signal, and app | When its row is deleted | The 5-minute value is what all reports and standards use |
| The readings of a second source of the same glucose | When deleted | One reading for each 5 minutes is kept |
| Each temp basal and automatic bolus | When deleted | The rate and the insulin of each 5 minutes are kept |
| The loop's reasons and predictions | After 1 day (as today) | Only useful to look at the last hours |
| Each device report | By its schedule | IOB, COB, reservoir and batteries are series |
| The app's own IDs of a treatment | When deleted | No app we know uploads again after a month |
| Which app and device made an event | When deleted | The identity is kept |

Nothing a person entered is lost. See Open question 1 for a way to lose nothing at all.

## Size of a site

A looping site, with the measured numbers:

| Part | Size |
| --- | --- |
| 31 hot days at about 370 KB (without device reports) | about 11.5 MB |
| Device reports, thinned as today | about 2 MB |
| Cold days, each year (about 2 KB a day) | about 0.7 MB |
| **After 1 year** | **about 14 MB** |
| **After 9 years** | **about 20 MB** |

- Today the same site is 20 MB after about 5 weeks.
- **The 31 hot days are more than half of 20 MB for a looping site.** The two changes below take them to about
  5.5 MB, and then the site is under 20 MB for 15 years or more.
- A site with a CGM and no loop is about 2 MB hot and 0.3 MB for each year.
- The cold sizes other than glucose and IOB are estimates.

## Hot rows can be cheaper too

Packaging does not help the hot days. Two changes, each on its own, make hot rows smaller:

1. **No `event_groups` row for an automatic record.** A temp basal and an automatic bolus with no notes carry
   nothing a person wrote. Their group row is 190 KB a day, a third of the site. The few fields an app needs back
   (its ID, `entered_by`) can be columns of the record.
2. **Fewer indexes on `event_groups`.** It has 12. Each partial unique index on a member's ID could be one index on
   one `member_id` column.

## Build order

Each step is one PR, and the site works after each one.

| Step | What | Reads change | State |
| --- | --- | --- | --- |
| 1 | `cold_days`, the versioned body, packaging of **glucose**, the delete rule | None | **Built** |
| 2 | The reader of packaged glucose; the Nightscout API makes `entries` from packaged days | `entries` | **Built** |
| 2b | The site's glucose page says that older readings are in Reports; a made entry by its `_id`; searches oldest first | Glucose page | |
| 3 | Device reports fold into `iob`, `iob_basal`, `cob`, reservoir and battery series | None yet: the day view reads them in step 5 | **Built** |
| 4a | The treatment series: `basal_rate`, `insulin_bolus`, `insulin_basal`, `carbs` | None yet | **Built** |
| 4b | Events: the day's list | The Cold Days page | **Built** |
| 4c | What is in force at the start of a day; a profile switch keeps its profile | The Cold Days page | **Built** |
| 4d | The Nightscout API makes `treatments` from packaged days | `treatments` | **Built** |
| 4e | The delete of treatment and event rows older than the hot days; a packaged day keeps what it has when its rows are gone | Data pages | |
| 5a | A summary day built from the series of its cold days (`coldSummaryDaily`), for a day in any timezone | None yet | **Built** |
| 5b | Reports (AGP, time in range, daily dose) read the days built from series; a week or a month is built when read | Reports | |
| 5c | **Remove the summaries:** the `summary_period` table, the freeze, the stale marks, the builds | | |
| 6 | Hot days by size pressure. **Remove most of the truncation.** | | |
| 7 | The Storage page; Download Data as CSV; cheaper hot rows | | |

`RETENTION_DELETE_ROWS` can now be `true` on staging: apps get the glucose of packaged days. The site's glucose page
shows only rows, and Reports show old days from the summaries, as before. It becomes the default when each
table's reader is built.

## What this takes the place of

### Summaries (step 5)

Each field of `summary_period` can be built from cold days:

| Field | From |
| --- | --- |
| The day's glucose line; AGP; time in range and the bands; mean, SD, GMI; counts | The `glucose` series |
| Total daily dose, basal and carbs | `basal_rate`, `insulin_bolus`, `insulin_basal`, `carbs` |
| IOB and COB lines | The `iob` and `cob` series |
| Finger prick count | The day's events |

- An AGP of 90 days reads 90 small rows. It needs no stored AGP.
- **To remove:** the `summary_period` table, `frozen_at` and the stale marks, the builds for a new timezone or
  week start, **Regenerate summaries**, and most of `background/summary_daily`.
- **To keep:** the report maths (percentiles, bands). It reads series, not summary rows.
- A site's day is not a UTC day. `db.ColdSeries` reads a time range of a series from the UTC days it is in.
- What a day built from series does not have: a meter reading in its glucose line. Its count is from the rows.
- **A change a person can see:** the mean and SD are of the 5-minute values, not of each reading. A site with a
  1-minute sensor gets a small difference.

### Truncation (step 6)

| Today | With cold days |
| --- | --- |
| Device reports thinned (10 min, 20 min, hourly) and deleted after 30 days; `utils/thin` | **Remove.** Their numbers are series. Device reports get short hot days of their own (2 days). |
| A site over 1.5 times its limit: clear old JSON, delete device reports over 7 days old | **Remove.** The site gets fewer hot days. |
| The same site, if still over: delete readings and treatments over 90 days old | **Remove.** This loses data. Packaging does not. |
| The **Truncate** button | **Keep.** It packages now, with the fewest hot days, and vacuums. It packages now already. |
| Delete loop settings no record refers to | **Keep**, in the package run. |
| Vacuum | **Keep.** |
| The size limit, writes refused at 4 times the limit, the abuse score | **Keep.** Packaging does not help a site that writes too many rows today. |

To check before the thinning goes: does a page or an app read a device report over 2 days old?

## Open questions

1. **Lose nothing: an archive.** Before a day's rows are deleted, write them, as sent, to a compressed file for
   each month in the site's S3 folder. "Deleted" then means "not in the site DB", not "gone", and a month could be
   read back. It needs S3, storage for each site, and a way to restore. Is this wanted, as an option of a site?
2. **Hot days:** is 3 a safe floor for each app (AAPS, Trio, Loop, xDrip+)?
3. **Temp basals to apps:** for a packaged day, send one temp basal for each hour, or none?
4. **`iob_basal`:** keep it, or only total IOB?
5. **Profiles:** each profile switch keeps its profile, and each day has it again when it is in force. This was
   chosen over a table that keeps each profile one time: one table, and a day is read alone.
6. **Events for 20 years:** 10 events a day are the largest part of a cold day. Is that acceptable for ever, or
   do events older than N years become counts?
7. **A site's own setting:** can a person choose more hot days if their DB has space?
