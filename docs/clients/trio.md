# Trio and Nighttroop

## Overview

[Trio](https://github.com/nightscout/Trio) is an iOS closed loop app (oref).

- Uploads: glucose, pump history, carbs, temp targets, overrides, loop status, profile.
- Downloads: carbs and temp targets that others enter. It can also use Nightscout as its CGM.
- [iAPS](https://github.com/Artificial-Pancreas/iAPS), which Trio grew from, is nearly the same. Differences are noted.
- No Socket.IO, API v2 or API v3, in Trio or iAPS.
- Checked against Trio's `main` branch:
  - Trio: `Trio/Sources/Services/Network/Nightscout/NightscoutAPI.swift`, called from `NightscoutManager.swift`.
  - iAPS: `FreeAPS/Sources/Services/Network/NightscoutAPI.swift`.
- Tests: `tests/api/nightscout/trio/`. They send the token's SHA1 in `api-secret`, as Trio does.

## Set up

1. On the site, use Add Device and select Trio as the client (#524).
   - This makes a PWD Phone device (#527) in mg/dL, with a setup guide.
2. In Trio, enter the site URL.
3. Enter the device's token as the API secret.

**Note:** Trio keeps only the URL's scheme, host and port, and adds a fixed path (`/api/v1/...`). It drops a path or
a `?token=` in the URL (`NightscoutAPI`, each function's `URLComponents`). Thus Nightscout under a path does not work.

## Authentication

- Every request sends `api-secret: <SHA1 hex of the secret>` (`String.sha1()`), as Nightscout expects for its
  `API_SECRET`.
- No `token` parameter, no `Authorization` header, no JWT.
- Nighttroop takes a device's token in `api-secret`, or its SHA1 (40 hex digits), since #512.
  - It finds the SHA1 by a digest kept for each token (not for an admin or an identity with TOTP).
- Without a secret, the connection check is `GET /api/v1/treatments.json`. Nighttroop returns 401: it needs a token.

## Units and dates

- Units:
  - Trio reads `sgv` as mg/dL, decoded as an integer, and `targetTop`/`targetBottom` as mg/dL.
  - The Nightscout API always gives both in mg/dL, whatever the identity's or site's units: 120 mg/dL is 120 on an
    mmol/L site too.
  - Profile: Trio uploads it in the user's units, labelled `mg/dl` or `mmol`. On import, it finds mmol/L from
    `units`, or from targets of 39 or less.
- Dates (`JSONCoding`, `Formatters.swift` `customISO8601`):
  - Trio reads a date only from an ISO 8601 string with a time zone, with or without fractional seconds
    (`2026-11-01T12:00:00Z`, `2026-11-01T12:00:00.123Z`). A number, or a date without a zone, fails.
  - Trio writes dates with milliseconds and `Z`.
  - Nighttroop writes `created_at`, `timestamp`, `dateString` and `startDate` as Nightscout does (JavaScript
    `toISOString`), with three fractional digits: `2026-11-01T12:00:00.000Z`. Thus `deleteGlucose` finds a reading
    by the `.SSS` `dateString` it sent.
- Types: numbers are Swift `Decimal` or `Int`. One field of the wrong type, or one missing required field, fails the
  whole list.

## Uploads

- Sent when data changes (`BaseNightscoutManager+Subscribers.swift`), in batches of 100.
- Each is a `POST` with `Content-Type: application/json` and a pretty printed body. Treatments go to
  `/api/v1/treatments.json`.
- Trio ignores the response body. It marks records uploaded only on a 2xx (the profile, only on a 200), else it tries
  again later.
- A batch with up to 10 records Nighttroop cannot store: the rest are stored with a 200, so Trio marks the batch
  uploaded and does not send the bad records again, as Nightscout stores them all. The whole batch was a 400, sent
  again and refused each time. The site's client errors show the records left out.

| Upload | When (source) | Body |
| --- | --- | --- |
| Glucose, `/api/v1/entries.json` | A reading arrives (`uploadGlucose`, `GlucoseStorage.getGlucoseNotYetUploadedToNightscout`) | CGM: `{ id (UUID), sgv, direction, date (ms), dateString, unfiltered, filtered, glucose, type: "sgv" }`. Manual: `{ id, mbg, date, dateString, type: "mbg" }`. In mg/dL |
| Pump history | Each pump event (`uploadPumpHistory`, `PumpHistoryStorage.getPumpHistoryNotYetUploadedToNightscout`) | See [Pump history](#pump-history) |
| Carbs | Carbs entered (`CarbsStorage.getCarbsNotYetUploadedToNightscout`; fat and protein units from `getFPUsNotYetUploadedToNightscout`) | `{ eventType: "Carb Correction", created_at, enteredBy, notes, carbs, fat, protein, foodType, id }`. `notes` and `foodType` are the same text, can end with a new line, and come back as sent |
| CGM sensor start | With the glucose | `{ eventType: "Sensor Start", created_at, enteredBy, notes }` |
| Temp target | Start (`TempTargetsStorage.getTempTargetsNotYetUploadedToNightscout`) | `{ eventType: "Temporary Target", created_at, duration, enteredBy, notes (its name), targetTop, targetBottom }`, mg/dL, no `units` |
| Temp target run | Early end (`getTempTargetRunsNotYetUploadedToNightscout`) | Same `created_at`, real `duration`. Replaces the first, as the Nightscout v1 upsert on `created_at` and `eventType` does |
| Override | Start (`OverrideStorage.getOverridesNotYetUploadedToNightscout`) | `{ eventType: "Exercise", created_at, duration, enteredBy, notes (its name) }`. `duration: 43200` (30 days) if no end |
| Override run | Duration change or end (`uploadOverrideRuns`) | Trio first deletes the first upload ([Deletes](#deletes)): Nighttroop soft deletes it. Then the run, with the real duration. A run with the same `created_at`, while the first is kept, replaces it |
| Note | `uploadNoteTreatment` | `{ eventType: "Note", created_at, enteredBy, notes }` |
| Connection check | `checkConnection` | `{ eventType: "Note", enteredBy: "Trio", notes: "Trio connected" }` |
| Device status, `/api/v1/devicestatus.json` | Each loop (`performUploadDeviceStatus`) | See [Device status](#device-status) |
| Profile, `/api/v1/profile.json` | Schedule saved (`uploadProfiles`) | See [Profile](#profile). Needs exactly a 200 |

### Pump history

Each event has `created_at`, `enteredBy: "Trio"` and the event's `id`.

| Pump event | `eventType` | Fields |
| --- | --- | --- |
| Automatic bolus | `SMB` | `insulin` |
| Bolus | `Bolus` | `insulin` |
| Insulin not from the pump | `External Insulin` | `insulin` |
| Temp basal | `Temp Basal` | `rate`, `absolute` (U/h), `duration` (min) |
| Suspend, resume | `Note` | `notes`: `PumpSuspend` or `PumpResume` |
| Rewind | `Insulin Change` | |
| Site change | `Site Change` | |
| Pump alarm | `Announcement` | `duration: 30`, `notes` |

Nighttroop refused `SMB` and `External Insulin` as unknown types. One batch holds every kind, so one SMB lost the
whole batch. Fixed in #509: both are kept as insulin doses with their type.

### Device status

No `created_at`: Nightscout and Nighttroop use the arrival time.

| Object | Fields |
| --- | --- |
| `device` | `"Trio"` |
| `openaps` | `iob` (`iob`, `activity`, `basaliob`, `bolusiob`, `netbasalinsulin`, `bolusinsulin`, `iobWithZeroTemp`, `lastBolusTime`, `lastTemp`, `time`); `suggested` and `enacted` (the oref determination: `id`, `reason`, `units`, `insulinReq`, `eventualBG`, `sensitivityRatio`, `rate`, `duration`, `IOB`, `COB`, `predBGs` with `IOB`/`ZT`/`COB`/`UAM`, `deliverAt`, `carbsReq`, `temp`, `bg`, `reservoir`, `timestamp`, `ISF`, `current_target`, `TDD`, `minDelta`, `expectedDelta`, `minGuardBG`, `minPredBG`, `threshold`, `CR`, `received`; glucose in the user's units); `version`, `recommendedBolus`, `dosingMode` |
| `pump` | `clock`, `battery` (`percent`, `voltage`, `string`, `display`), `reservoir`, `status` (`status`, `bolusing`, `suspended`, `timestamp`), `bolusIncrement` |
| `uploader` | `battery`, `isCharging` |

Nighttroop gives back (followers read it, Trio does not):

- `openaps`: whole.
- `pump`: all fields above while the JSON as sent is kept. After that, only the clock, battery percent and reservoir.
- `uploader`: `battery` and `isCharging`.
- If iOS does not know the phone's battery, Trio sends `uploader.battery` -100 (its `batteryLevel` -1 x 100).
  Nighttroop keeps the device status without it.

### Profile

- Body: `{ defaultProfile: "default", startDate, mills, units, enteredBy, store: { default: { dia, carbs_hr, delay,
  timezone, target_low, target_high, sens, basal, carbratio, units } }, bundleIdentifier, deviceToken,
  isAPNSProduction, overridePresets, teamID, expirationDate }`.
- Each schedule entry: `{ time: "HH:mm", value, timeAsSeconds }`. `units` is `mg/dl` or `mmol`.
- The push fields let LoopFollow and Trio's remote control send Apple push notifications to Trio.
- Nighttroop registers `deviceToken`, `bundleIdentifier` and `isAPNSProduction` as a `trio` push registration for
  the uploading identity (`/settings/push-registrations`).

## Downloads

Each is a `GET` with `api-secret` and a 60 second timeout. Trio needs a 2xx, except for glucose: it takes any
status, and a body it cannot decode is no readings.

### Carbs: `GET /api/v1/treatments.json`

Every minute when downloads are on (`FetchTreatmentsManager`, `fetchCarbs`), from a day ago:

```
find[carbs][$exists]=true
&find[$and][0][enteredBy][$ne]=Trio
&find[$and][1][enteredBy][$ne]=AndroidAPS
&find[$and][2][enteredBy][$ne]=openaps://AndroidAPS
&find[$and][3][enteredBy][$ne]=iAPS
&find[$and][4][enteredBy][$ne]=loop://iPhone
&find[created_at][$gt]=<ISO date>
```

- Decoded as `CarbsEntry`:
  - Required: `created_at` (date), `carbs` (number).
  - Optional: `_id`, `actualDate`, `fat`, `protein`, `notes`, `enteredBy`, `isFPU`, `fpuID`.
- Trio stores the entries with a time it does not have yet.
- Nighttroop:
  - Gives only treatments with `carbs`. One without fails the whole list.
  - Leaves out treatments whose `enteredBy` (kept as sent) matches a `$ne`. Thus Trio does not download its own
    carbs and add them again.
- iAPS sends `find[enteredBy][$ne]` five times instead of `$and`, encoded twice (`iAPS%20Watch`). This works too.

### Temp targets: `GET /api/v1/treatments.json`

With the carbs (`fetchTempTargets`):

```
find[eventType]=Temporary+Target&find[duration][$exists]=true&<the same enteredBy filters>&find[created_at][$gt]=<ISO date>
```

- Decoded as `TempTarget`:
  - Required: `_id` (string), `created_at`, `duration` (number).
  - Optional: `targetTop`, `targetBottom` (mg/dL), `name`, `enteredBy`, `reason`, `isPreset`, `enabled`,
    `halfBasalTarget`.
- Trio skips a target whose `created_at` it has, or whose `reason` is `Cancel`. It turns on the newest.
- Nighttroop keeps the milliseconds of `created_at`.

### Glucose, Nightscout as the CGM: `GET /api/v1/entries/sgv.json`

On the CGM timer (`fetchLastGlucose`): `count=1600&find[dateString][$gte]=<ISO date of its last reading>`. iAPS also
sends `count=500&sort$desc=dateString`.

- Decoded as `BloodGlucose`:
  - Required: `dateString` (date).
  - Optional: `_id`/`id`; `sgv` and `mbg` (integer, or a number it truncates); `direction`; `date` (ms, or from
    `dateString`); `unfiltered`, `filtered`, `noise` (integer); `type`, `activationDate`, `sessionStartDate`,
    `transmitterID`.
  - `direction` must be a Nightscout name (`Flat`, `FortyFiveUp`, `NOT COMPUTABLE`, …). Any other string fails the
    whole list.
- Nighttroop gives `direction` as a Nightscout name from any case or arrow (`NOT COMPUTABLE` if unknown), and
  `noise` as a whole number. Thus one uploader's reading cannot stop Trio's glucose.

### Profile import: `GET /api/v1/profile.json?count=1`

In onboarding (`importSettings`):

- Trio needs a 2xx with `Content-Type: application/json`.
- Trio decodes the first item as `FetchedNightscoutProfileStore`. Required: `_id`, `defaultProfile`, `startDate`
  (string), `mills` (number), `enteredBy`, `created_at` (string), `store`.
- Trio takes `store.default`. It needs `dia`, `carbs_hr` (integer), `delay`, `timezone`, `units` and the five
  schedules of `{ time, value, timeAsSeconds }`.
- Nighttroop left out `enteredBy`, so the import failed. Fixed in #509: `enteredBy` is the uploading identity.

### Announcements (iAPS only): `GET /api/v1/treatments.json`

- iAPS takes remote commands from announcements: `bolus: 1`, `meal: 20,5,5`, `looping: false`, `override: name`
  and others.
- Query: `find[eventType]=Announcement&find[enteredBy]=remote&find[created_at][$gte]=<ISO date>`.
- Required: `created_at`, `enteredBy`, `notes`.
- Trio dropped this for Apple push notifications.
- Only announcements `enteredBy` `remote` come back, not the pump alarms iAPS uploads.

## Deletes

When the user deletes a record in Trio, Trio deletes it by query (`deleteCarbs`, `deleteInsulin`, `deleteGlucose`,
`deleteNightscoutOverride`):

| Record | Request |
| --- | --- |
| Carbs, insulin | `DELETE /api/v1/treatments.json?find[id][$eq]=<Trio's id>` |
| Glucose | `DELETE /api/v1/entries.json?find[$or][0][id][$eq]=<id>&find[$or][1][dateString][$eq]=<ISO date>` |
| Override | `DELETE /api/v1/treatments.json?find[created_at][$eq]=<ISO date>&find[eventType][$eq]=Exercise` |

- iAPS deletes the same way, by `find[created_at][$eq]` and `find[carbs][$exists]` or similar.
- Nighttroop:
  - Keeps Trio's `id` as the identifier, so `find[id][$eq]` finds it.
  - Soft deletes what the search finds, at most 1000 a request.
  - Runs each `find[$or]` alternative as its own search. Each needs a date or an id.
  - Returns a 400 for a filter it does not take for a delete, to not delete more than asked.

## Quirks

- `find[eventType]=Temporary+Target` has a literal `+` for the space, which decodes to `Temporary Target`.
- iAPS percent-encodes its `enteredBy` values before `URLComponents` encodes them again. `iAPS Watch` arrives as
  `iAPS%20Watch`.
- `find[created_at]` filters on when the record happened, as in Nightscout (#514). `srvCreated` filters on when it
  was uploaded.

## Status on Nighttroop

| Interaction | Nighttroop |
| --- | --- |
| `api-secret` as the SHA1 of the secret | Works (#512) |
| Connection check, `POST` Note | Works |
| Glucose upload | Works |
| Pump history upload with `SMB` and `External Insulin` | Works (#509) |
| Carbs, sensor start, note uploads | Works |
| Temp target and override upload | Works |
| Temp target and override run (same `created_at`, new duration) | Fixed: replaces the first, as in Nightscout |
| Device status upload | Works. Pump `status.bolusing`, `suspended`, battery `string`, `display` and `uploader.isCharging` fixed in this PR |
| Profile upload | Works |
| Profile import | Works (`enteredBy`, #509) |
| Carbs fetch (`find[carbs][$exists]`, `enteredBy` `$ne`) | Fixed |
| Temp target fetch | Works in mg/dL. Own ones left out since this PR |
| Glucose fetch (Nightscout as CGM) | Works in mg/dL |
| Glucose upload of a reading already stored (Nightscout as CGM), or another app's copy of Trio's reading | Fixed: one reading, with the first upload's values and device. The copy adds only the fields it does not have |
| iAPS announcements (`find[enteredBy]=remote`) | Fixed |
| Deletes by query | Fixed (were a 405) |
| mmol/L site or identity | Not supported by Trio's fetches. A device added for Trio is in mg/dL |
