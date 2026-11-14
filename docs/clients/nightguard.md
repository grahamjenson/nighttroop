# Nightguard

How Nightguard talks to Nightscout, and how it works with Nighttroop.

## Overview

- [Nightguard](https://github.com/nightscout/nightguard): an iOS and watchOS Nightscout follower, with widgets,
  complications, alarms and a few care actions.
- Source: commit `608be1b` (2026-10-24). Files are under `nightguard/`; unnamed ones are
  `external/NightscoutService.swift`, and numbers are its lines.
- Tests: `tests/api/nightscout/nightguard/`.
- No Socket.IO: Nightguard polls. Alarms and snoozes are worked out on the phone, and nothing about them goes to
  Nightscout. Its background push service is its own, not Nightscout.

## Set up

1. Add a device for Nightguard. Nighttroop sets its units to mg/dL (#524).
2. In Nightguard, set the Nightscout URL to `https://<site>.ns.<host>`.
3. Set the token to the device's Nightscout token.

- **Glucose is mg/dL.** Nightguard reads each `sgv` and `mbg` as mg/dL, and never reads an entry's `units`. The
  Nightscout API always gives mg/dL. The app shows it in the status's `settings.units` (the identity's own), so a
  device on an mmol/L site shows mmol/L.
- The "force v1" setting sends everything to v1.

## Authentication

- The token is kept apart from the URL (`UserDefaultsRepository.cleanBaseURL`, `normalizedToken`, which also removes
  a `token=` prefix).
- **Version check** (`NightscoutAPIClient.requestV3`, line 158; `determineV3Support`, line 256): `GET /api/v3/version`
  with no token.
  - 2xx: v3. 404 or 405: v1. Anything else: the read fails.
  - The result is cached for each server for the life of the app.
- **v3:** a JWT from `GET /api/v2/authorization/request/token=<token>` (`obtainJWT`, line 497).
  - Reads `token` (a non-empty string). Decodes `exp` from the JWT's payload, and caches the JWT until a minute
    before then.
  - Sends `Authorization: Bearer <jwt>`. On a 401, gets a new JWT and tries once more.
  - With no token: calls v3 without one, and falls back to v1 on a 401.
- **v1** (`performLegacy`, line 447; `UserDefaultsRepository.getUrlWithPathAndQueryParameters`): the token as a
  `token` query parameter, after the others. No `api-secret` header.
- **Fallback:** GET reads fall back from v3 to v1 on a 404 or 405, a timeout, or no connection. Writes never fall back.
- **Headers:** `Accept: application/json` and no caching. `Content-Type: application/json` with a body.
- **Query:** parameters sorted by name and URL-encoded (`|` as `%7C`, `[` and `]` as `%5B` and `%5D`).

## What syncs

Reads: units, glucose, IOB, COB, uploader battery, delta, pump, cannula, sensor and battery age, temp target, today's
carbs and boluses, and the report. Writes: care actions.

### Units

`GET /api/v1/status.json?token=…` (`readLegacyStatus`, line 1083; `units(from:)`, line 1104).

- When the URL is set, and on the watch, unless the units are set by hand.
- Reads `settings.units` as a string: with "mg" in it, mg/dL; anything else, mmol/L; none, fails.
- Reads no thresholds: alarms are set in the app.

### Glucose

| Function | v3 query (`GET /api/v3/entries`) | v1 query (`GET /api/v1/entries.json`) |
| --- | --- | --- |
| `readEntriesHead` (874), the 48 h stream | `date$gte=<ms>&fields=identifier,_id,date,dateString,mills,type,sgv,mbg,direction,units&limit=500&sort$desc=date&type$in=sgv\|mbg` | `count=500` |
| `readLatestEntriesHybrid` (947), current value | as above with `limit=<2…100>`, no `date$gte` | `count=<limit>`, sent at the same time as v3 |
| `readCurrentData` (1602) | `fields=identifier,_id,date,dateString,sgv,direction,units&limit=2&sort$desc=date` | `count=2` |
| `readChartData` (824) | `fields=date,sgv&limit=20&sort$desc=date` | `count=20` |
| `readChartDataWithinPeriodOfTime` (1150), a day, yesterday, today, statistics | `date$gte=<ms>&date$lt=<ms>&fields=date,dateString,mills,type,sgv,mbg,direction&sort$desc=date&type$in=sgv\|mbg`, then `skip=<n>` | `count=1440&find[date][$gte]=<ms>&find[date][$lt]=<ms>`, moving `$lt` back to the oldest it has |
| Report (`report/ReportDataProvider.swift`, `fetchPages`, 83) | `date$gte`, `date$lt` as whole ms, `limit=1000&sort$desc=date&type$in=sgv\|mbg`, `skip` | `count=1000&find[date][$gte]&find[date][$lt]` |

- **Milliseconds:** Swift `Double`s written with `"\(value)"`, so always with a fraction. For example
  `1761234567890.0` or `1761234567890.123`. The report's are whole.
- **Paging stops** on an empty page, a page whose oldest date reaches the start, a page no older than the last, or
  32 pages.
- **Responses:** v3 responses are unwrapped from `result`. Both must be a JSON array.
- **Each entry** (`parseEntryRecords`, 1024; `entryTimestampMillis`, 1368):
  - Time: from `date`, then `dateString`, then `mills`. A number of milliseconds (seconds if under 1e11),
    `{"$date"…}`, or an ISO string.
  - `sgv` and `mbg`: numbers or strings of numbers, in mg/dL. An entry with an `sgv` is a CGM reading. One with only
    an `mbg` is a meter reading.
  - `direction` (`DoubleUp` … `DoubleDown`), `type`, `identifier` and `_id`: strings.
- **Current value:** the newest `sgv`. The delta is its difference from the one before.

### IOB, COB, uploader battery and delta

`GET /api/v2/properties?token=…` (`loadPropertiesEnrichment`, 1707), after each current value. No `Accept` header,
10 s timeout. All fields are optional, and Nightguard ignores any error.

| Field | Value |
| --- | --- |
| `upbat.display` | a string, such as `"88%"`, from the newest device status with a battery in the last 30 minutes |
| `iob.display` | anything, shown with a `U` after it |
| `cob.display` | a number or a string of one, shown as whole grams |
| `delta.mgdl` | a number or a string of one |
| `delta.display` | a string, such as `"+6"` |

The properties also have `pump`, `openaps` and `loop`, in Nightscout's shapes, for watch faces. Nightguard does not
read them.

### Pump

`GET /api/v1/devicestatus.json?count=5&token=…` (`readDeviceStatus`, 2084; `deviceStatusData`, 2110), always v1. From
the first status with a `pump`:

- `pump.reservoir`: a number of units.
- `pump.extended` (AAPS): `ActiveProfile` (a string); `BaseBasalRate` and `TempBasalAbsoluteRate` (numbers, `Double`
  only) for the temp basal percent; `TempBasalRemaining` (whole minutes). Without it, only the reservoir shows.

### Cannula, sensor and battery age

`readLastTreatementEventTimestamp` (1885), for `Site Change` (5 days back), `Sensor Change` (14) and
`Pump Battery Change` (40):

- v3 `GET /api/v3/treatments?created_at$gte=<ISO>&eventType$eq=<type>&fields=created_at,date&limit=1&sort$desc=created_at`
- v1 `GET /api/v1/treatments?count=1&find[created_at][$gte]=<ISO>&find[eventType]=<type>`

- ISO times are from `ISO8601DateFormatter`, for example `2026-11-02T09:15:00Z`.
- Reads `created_at` as a string (`yyyy-MM-dd'T'HH:mm:ssZ`, or with `.SSS`), else `date`. Ignores one older than
  asked for. With none, the age is from now.

### Temp target

`readLastTemporaryTarget` (1938), 1 day back:

- v3 `GET /api/v3/treatments?created_at$gte=<ISO>&eventType$eq=Temporary%20Target&limit=1&sort$desc=created_at`
- v1 `GET /api/v1/treatments?count=1&find[created_at][$gte]=<ISO>&find[eventType]=Temporary%20Target`

`TemporaryTarget.parse` (`external/model/TemporaryTarget.swift`) reads `targetTop` and `targetBottom` (numbers;
mmol/L if `units` has "mol" in it, else mg/dL; 100 if missing), `duration` (minutes) and `created_at` (a string). The
target is active until `created_at` plus `duration`.

### Today's carbs and boluses

`readLatestTreatements` (2133):

- v3 `GET /api/v3/treatments?created_at$gte=<start of day>&eventType$in=Carb Correction|Meal Bolus|Correction Bolus|Bolus Wizard&limit=500&sort$desc=created_at`
- v1 `GET /api/v1/treatments.json?count=500&find[created_at][$gte]=<start of day>`

It filters by `eventType` and `created_at` again, and reads `carbs` and `insulin`.

### Report

`report/ReportDataProvider.swift`. For entries, see [Glucose](#glucose).

- **Treatments** (`fetchPages`, 83):
  - v3 `created_at$gte=<ISO>&created_at$lt=<ISO>&limit=1000&sort$desc=created_at`, then `skip`.
  - v1 `/api/v1/treatments.json?count=1000&find[created_at][$gte]&find[created_at][$lt]`.
  - `decodeTreatments` (208) reads `created_at`, `eventType`, `carbs`, `insulin`, `duration`, `absolute` or `rate`,
    `percent`, `profile`, `targetBottom`, `targetTop`, `units`, `isSMB`, `type`, and `_id` or `identifier`.
- **Profiles** (`loadProfiles`, 148): `GET /api/v1/profile.json?count=1000`, then `find[startDate][$lt]=<ISO>`.
  - `decodeProfiles` (229) reads `startDate`, `store`, `defaultProfile`, `timezone` and `units`.
  - In each profile: `dia`, `carbs_hr`, `units`, `timezone`, and `basal`, `carbratio`, `sens`, `target_low` and
    `target_high` as lists of a `value` with a `time` (`HH:MM`) or `timeAsSeconds`.

### Care actions

`createTreatment` (2045) needs a token. Each body also has `date` and `mills` as **strings** of milliseconds
(`Date.toUTCMillis`), `created_at` as an ISO time, and `app` and `enteredBy` as `"nightguard"`.

| Action | Body |
| --- | --- |
| `createTemporaryTarget` (1986) | `eventType: "Temporary Target"`, `duration`, `reason`, `targetBottom`, `targetTop` (whole mg/dL), `units: "mg/dl"` |
| `deleteTemporaryTarget` (1998) | `eventType: "Temporary Target"`, `duration: 0`, `reason: "Canceled"` |
| `createCarbsCorrection` (2016) | `eventType: "Carb Correction"`, `carbs`, `duration: 0` |
| Cannula, sensor and battery changes (`createChangeTreatment`, 2036) | `eventType` `Site Change`, `Sensor Change` or `Pump Battery Change`, `notes: ""`, `carbs: null`, `insulin: null`, at the time picked |

- v3: `POST /api/v3/treatments` with the JWT.
- v1: `POST /api/v1/treatments?now=(Function)&token=…`. The `now` is a Swift function's description, not a time.
- Any 2xx is a success. Anything else shows as an error.

## Known problems and limits

- The v3 version check must be public. A 401 there fails every read, with no fallback.
- Date filters have a fraction of a millisecond.
- `eventType$eq`, not `eventType`.
- `date` and `mills` are written as strings.
- `sgv` is mg/dL whatever the units. Any entry with an `sgv` is a CGM reading.
- v3 and v1 run in parallel for the current value. The fresher answer wins.

## Status in Nighttroop

- **Works:**
  - JWT from `/api/v2/authorization/request/token=…`, and Bearer on v3.
  - v1 with `?token=`.
  - v3 and v1 entries, with their `fields`, `type$in`, `limit`, `skip`, `count` and sort.
  - v2 properties `upbat`, `iob`, `cob`, `delta`.
  - Device status `pump.reservoir`.
  - Temp target, site, sensor and battery change reads.
  - Today's carbs and boluses.
  - Report treatments and profiles.
  - Care writes through v3 and v1, with string `date` and `mills`. `enteredBy` comes back as the identity, not
    "nightguard".
- **Fixed in #514:**
  - `GET /api/v3/version` without a token. It needed a token, so every read failed.
  - `status.json` `settings.units`. It was missing, so the units could not be read.
  - Date filters with a fraction (`date$gte=…0.25`, `find[date][$lt]=…0.0`). They were 400.
  - A meter reading's entry with no `sgv`. Each entry had both, so meter readings were CGM readings.
  - `eventType$eq` on v3 treatments. Nighttroop ignored it, so the newest treatment of any type came back.
  - `created_at` filters and sort on treatments. They used the upload time, not the time of the treatment.
  - Device status `pump.extended` (profile, temp basal). It was dropped.
- **Not supported:** glucose in mmol/L. Nightguard reads mg/dL only, so Nighttroop sets a device added for Nightguard
  to mg/dL (#524).
- **Not used:** Socket.IO.
