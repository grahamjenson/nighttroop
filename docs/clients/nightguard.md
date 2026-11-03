# Nightguard

[Nightguard](https://github.com/nightscout/nightguard) is an iOS and watchOS Nightscout follower, with widgets,
complications, alarms and a few care actions. This is how it talks to Nightscout, read from its source at commit
`608be1b` (2026-10-24). Files are under `nightguard/` and, unless named, in `external/NightscoutService.swift`.

The tests are in `tests/api/nightscout/nightguard/`.

## Setting it up with Nighttroop

- Use the Nightscout URL `https://<site>.ns.<host>` and a device's Nightscout token.
- **Its identity must be in mg/dL.** Nightguard reads every `sgv` and `mbg` as mg/dL and never looks at an
  entry's `units`, so with mmol/L settings its readings would be about 18 times too low, and its low alarms would
  go off. A device added for Nightguard in the site's Add Device has its units set to mg/dL (#524), and an admin
  can set an existing device's units from its row (#523). To see mmol/L in the app, turn on "Manually set Units"
  in its settings and pick mmol/L.

## Authentication

- The token is kept apart from the URL (`UserDefaultsRepository.cleanBaseURL`, `normalizedToken`, which also strips
  a `token=` prefix).
- **v3** (`NightscoutAPIClient.requestV3`, line 158): first `GET /api/v3/version` with no token
  (`determineV3Support`, line 256). A 2xx means v3 is used, a 404 or 405 means v1 is used, and anything else fails
  the read. The answer is cached per server for the life of the app.
- With v3 it gets a JWT from `GET /api/v2/authorization/request/token=<token>` (`obtainJWT`, line 497). It reads
  `token` as a non-empty string and decodes `exp` from the JWT's payload to cache it until a minute before it
  expires. It sends `Authorization: Bearer <jwt>`; on a 401 it drops the JWT, gets another and tries once more.
- With no token it calls v3 without one, and falls back to v1 on a 401.
- **v1** (`performLegacy`, line 447, and `UserDefaultsRepository.getUrlWithPathAndQueryParameters`): the token as a
  `token` query parameter, after the others. No `api-secret` header.
- GET reads fall back from v3 to v1 on a 404 or 405, and on a timeout or no connection. Writes never fall back.
- Every request has `Accept: application/json` and no caching, and `Content-Type: application/json` with a body.
  Query parameters are sorted by name and URL-encoded (`|` as `%7C`, `[` and `]` as `%5B` and `%5D`).
- A setting, "force v1", sends everything to v1.

## Calls

### Units: `GET /api/v1/status.json?token=…`

`readLegacyStatus` (line 1083) and `units(from:)` (line 1104), when the URL is set and on the watch, unless units are
set by hand. Reads `settings.units` as a string: with "mg" in it is mg/dL, anything else mmol/L, and none fails.
It reads no thresholds: alarms are set in the app.

### Glucose: `GET /api/v3/entries` or `GET /api/v1/entries.json`

| Function | v3 query | v1 query |
| --- | --- | --- |
| `readEntriesHead` (874), the 48 h stream | `date$gte=<ms>&fields=identifier,_id,date,dateString,mills,type,sgv,mbg,direction,units&limit=500&sort$desc=date&type$in=sgv\|mbg` | `count=500` |
| `readLatestEntriesHybrid` (947), current value | as above with `limit=<2…100>`, no `date$gte` | `count=<limit>`, sent at the same time as v3 |
| `readCurrentData` (1602) | `fields=identifier,_id,date,dateString,sgv,direction,units&limit=2&sort$desc=date` | `count=2` |
| `readChartData` (824) | `fields=date,sgv&limit=20&sort$desc=date` | `count=20` |
| `readChartDataWithinPeriodOfTime` (1150), a day, yesterday, today, statistics | `date$gte=<ms>&date$lt=<ms>&fields=date,dateString,mills,type,sgv,mbg,direction&sort$desc=date&type$in=sgv\|mbg`, then `skip=<n>` | `count=1440&find[date][$gte]=<ms>&find[date][$lt]=<ms>`, moving `$lt` back to the oldest it has |
| Report (`report/ReportDataProvider.swift`, `fetchPages`, 83) | `date$gte`, `date$lt` as whole ms, `limit=1000&sort$desc=date&type$in=sgv\|mbg`, `skip` | `count=1000&find[date][$gte]&find[date][$lt]` |

- The milliseconds are Swift `Double`s written with `"\(value)"`, so always with a fraction, e.g.
  `1761234567890.0` or `1761234567890.123` (the report's are whole).
- Paging stops on an empty page, a page whose oldest date reaches the start, a page no older than the last, or
  32 pages.
- v3 responses are unwrapped from `result`; both must be a JSON array.
- Each entry (`parseEntryRecords`, 1024; `entryTimestampMillis`, 1368): the time from `date`, then `dateString`,
  then `mills`, as a number of milliseconds (seconds if under 1e11), `{"$date"…}`, or an ISO string. `sgv` and `mbg`
  as numbers or strings of numbers, in mg/dL. An entry with an `sgv` is a CGM reading, one with only an `mbg` a
  meter reading. `direction` as a string (`DoubleUp` … `DoubleDown`), `type`, `identifier` and `_id` as strings.
- The current value is the newest `sgv`, and the delta its difference from the one before.

### IOB, COB, uploader battery and delta: `GET /api/v2/properties?token=…`

`loadPropertiesEnrichment` (1707), after each current value, with no `Accept` header and a 10 s timeout. All of it
is optional and any error is ignored:

- `upbat.display`, a string such as `"88%"`;
- `iob.display`, anything, shown with a `U` after it;
- `cob.display`, a number or a string of one, shown as whole grams;
- `delta.mgdl`, a number or a string of one, and `delta.display`, a string such as `"+6"`.

### Pump: `GET /api/v1/devicestatus.json?count=5&token=…`

`readDeviceStatus` (2084) and `deviceStatusData` (2110), always v1. The first status with a `pump`:
`pump.reservoir` as a number of units, and from `pump.extended` (AAPS) `ActiveProfile` as a string,
`BaseBasalRate` and `TempBasalAbsoluteRate` as numbers (`Double` only) for the temp basal percent, and
`TempBasalRemaining` as a whole number of minutes. A pump with no `extended` shows only the reservoir.

### Cannula, sensor and battery age: treatments

`readLastTreatementEventTimestamp` (1885) for `Site Change` (5 days back), `Sensor Change` (14) and
`Pump Battery Change` (40):

- v3 `GET /api/v3/treatments?created_at$gte=<ISO>&eventType$eq=<type>&fields=created_at,date&limit=1&sort$desc=created_at`
- v1 `GET /api/v1/treatments?count=1&find[created_at][$gte]=<ISO>&find[eventType]=<type>`

The ISO time is `ISO8601DateFormatter`'s, e.g. `2026-11-02T09:15:00Z`. It reads `created_at` as a string
(`yyyy-MM-dd'T'HH:mm:ssZ`, or with `.SSS`), else `date`, and ignores one older than it asked for. With none it
shows the age as from now.

### Temp target: treatments

`readLastTemporaryTarget` (1938), 1 day back:

- v3 `GET /api/v3/treatments?created_at$gte=<ISO>&eventType$eq=Temporary%20Target&limit=1&sort$desc=created_at`
- v1 `GET /api/v1/treatments?count=1&find[created_at][$gte]=<ISO>&find[eventType]=Temporary%20Target`

`TemporaryTarget.parse` (`external/model/TemporaryTarget.swift`): `targetTop` and `targetBottom` as numbers,
in mmol/L if `units` has "mol" in it and otherwise mg/dL, 100 if missing; `duration` as a number of minutes; and
`created_at` as a string. It is active until `created_at` plus `duration`.

### Today's carbs and boluses: treatments

`readLatestTreatements` (2133):

- v3 `GET /api/v3/treatments?created_at$gte=<start of day>&eventType$in=Carb Correction|Meal Bolus|Correction Bolus|Bolus Wizard&limit=500&sort$desc=created_at`
- v1 `GET /api/v1/treatments.json?count=500&find[created_at][$gte]=<start of day>`

It filters by `eventType` and `created_at` again itself, and reads `carbs` and `insulin`.

### Report: treatments and profiles

`report/ReportDataProvider.swift`:

- Treatments (`fetchPages`, 83): v3 `created_at$gte=<ISO>&created_at$lt=<ISO>&limit=1000&sort$desc=created_at`,
  then `skip`; v1 `/api/v1/treatments.json?count=1000&find[created_at][$gte]&find[created_at][$lt]`.
  `decodeTreatments` (208) reads `created_at`, `eventType`, `carbs`, `insulin`, `duration`, `absolute` or `rate`,
  `percent`, `profile`, `targetBottom`, `targetTop`, `units`, `isSMB`, `type`, and `_id` or `identifier`.
- Profiles (`loadProfiles`, 148): `GET /api/v1/profile.json?count=1000`, then
  `find[startDate][$lt]=<ISO>`. `decodeProfiles` (229) reads `startDate`, `store`, `defaultProfile`, `timezone`
  and `units`, and in each profile `dia`, `carbs_hr`, `units`, `timezone`, and `basal`, `carbratio`, `sens`,
  `target_low` and `target_high` as lists of a `value` with a `time` (`HH:MM`) or `timeAsSeconds`.

### Writes: care actions

`createTreatment` (2045) needs a token. Every body also has `date` and `mills` as **strings** of milliseconds
(`Date.toUTCMillis`), `created_at` as an ISO time, and `app` and `enteredBy` as `"nightguard"`:

| Action | Body |
| --- | --- |
| `createTemporaryTarget` (1986) | `eventType: "Temporary Target"`, `duration`, `reason`, `targetBottom`, `targetTop` (whole mg/dL), `units: "mg/dl"` |
| `deleteTemporaryTarget` (1998) | `eventType: "Temporary Target"`, `duration: 0`, `reason: "Canceled"` |
| `createCarbsCorrection` (2016) | `eventType: "Carb Correction"`, `carbs`, `duration: 0` |
| cannula, sensor and battery changes (`createChangeTreatment`, 2036) | `eventType` `Site Change`, `Sensor Change` or `Pump Battery Change`, `notes: ""`, `carbs: null`, `insulin: null`, at the time picked |

It POSTs to `/api/v3/treatments` with the JWT, or to `/api/v1/treatments?now=(Function)&token=…` with v1 (the `now`
is a Swift function's description, not a time). Any 2xx is success; anything else is shown as an error.

## Socket use

None. Nightguard polls; alarms and snoozes are worked out on the phone, and nothing about them is sent to
Nightscout. Its background push service is Nightguard's own, not Nightscout.

## Quirks

- The v3 check must be public: a 401 there fails every read, with no fallback.
- Every date filter it builds has a fraction of a millisecond.
- It sends `eventType$eq`, not `eventType`.
- `date` and `mills` it writes are strings.
- It reads `sgv` as mg/dL whatever the units, and any entry with an `sgv` as a CGM reading.
- v3 and v1 run side by side for the current value, and the fresher answer wins.

## Status with Nighttroop

| Interaction | Status |
| --- | --- |
| `GET /api/v3/version` without a token | Fixed in #514: it needed a token, so every read failed |
| JWT from `/api/v2/authorization/request/token=…`, Bearer on v3 | Works |
| v1 with `?token=` | Works |
| `status.json` `settings.units` | Fixed in #514: it was missing, so units could not be read |
| v3 and v1 entries, with their `fields`, `type$in`, `limit`, `skip`, `count` and sort | Works |
| Date filters with a fraction (`date$gte=…0.25`, `find[date][$lt]=…0.0`) | Fixed in #514: they were a 400 |
| A meter reading's entry with no `sgv` | Fixed in #514: every entry had both, so meter readings were CGM readings |
| Glucose in mmol/L | Not supported: Nightguard reads mg/dL only, so a device added for Nightguard is set to mg/dL (#524) |
| `eventType$eq` on v3 treatments | Fixed in #514: it was ignored, so the newest treatment of any type came back |
| `created_at` filters and sort on treatments | Fixed in #514: they were on when it was uploaded, not when it happened |
| v2 properties `upbat`, `iob`, `cob`, `delta` | Works |
| devicestatus `pump.reservoir` | Works |
| devicestatus `pump.extended` (profile, temp basal) | Fixed in #514: it was dropped |
| Temp target, site, sensor and battery change reads | Works |
| Today's carbs and boluses | Works |
| Report treatments and profiles | Works |
| Care writes through v3 and v1, with string `date` and `mills` | Works (`enteredBy` comes back as the identity, not "nightguard") |
| Socket.IO | Not used |
