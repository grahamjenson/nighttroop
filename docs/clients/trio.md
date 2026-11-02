# Trio and Nighttroop

[Trio](https://github.com/nightscout/Trio) is an iOS closed loop app (oref). It uploads its glucose, pump history,
carbs, temp targets, overrides, loop status and profile to Nightscout, and downloads carbs and temp targets that
others enter. It can also use Nightscout as its CGM. [iAPS](https://github.com/Artificial-Pancreas/iAPS), which Trio
grew from, talks to Nightscout in nearly the same way; where it differs it is noted below.

This was checked against Trio's `main` branch. All its Nightscout calls are in
`Trio/Sources/Services/Network/Nightscout/NightscoutAPI.swift`, called from `NightscoutManager.swift`. iAPS's are in
`FreeAPS/Sources/Services/Network/NightscoutAPI.swift`. The tests are in `tests/api/nightscout/trio/`.

## Connecting and authentication

The user enters a URL and an API secret. Trio keeps only the URL's scheme, host and port: it builds every request
from those and a fixed path (`/api/v1/...`), so a path or a `?token=` in the URL is dropped
(`NightscoutAPI`, each function's `URLComponents`).

Every request sends the secret as `api-secret: <SHA1 hex of the secret>` (`String.sha1()`), as Nightscout expects for
its `API_SECRET`. It sends no `token` parameter and no `Authorization` header, and never asks for a JWT.

Nighttroop takes a device's token in `api-secret`, but not its SHA1, which it cannot check against the token. **So
Trio cannot sign in to Nighttroop yet** (TODO P0). The tests send the token itself in `api-secret`, as the closest to
Trio there is. Without a secret, Trio's connection check is a `GET /api/v1/treatments.json`, which Nighttroop refuses
with a 401, as it needs a token.

## Units

Trio reads glucose (`sgv`) as mg/dL and decodes it as an integer, and temp targets' `targetTop`/`targetBottom` as
mg/dL. Nighttroop returns both in the site's units (or the identity's). **The site, or Trio's identity, must be set
to mg/dL**: in mmol/L a reading of 120 mg/dL comes back as 6.66, which Trio truncates to 6 mg/dL. The tests set the
site to mg/dL. The profile is fine in either: Trio uploads it in the user's units, labelled `mg/dl` or `mmol`, and on
import works out mmol/L from `units` or from targets of 39 or less.

## Dates and types Trio reads back

Trio's JSON decoder (`JSONCoding`, `Formatters.swift` `customISO8601`) reads a date only from a string in ISO 8601
with a time zone, with or without fractional seconds (`2026-11-01T12:00:00Z`, `2026-11-01T12:00:00.123Z`). A number,
or a date without a zone, fails. It writes dates with milliseconds and `Z`. Numbers are Swift `Decimal` or `Int`;
one field of the wrong type, or one required field missing, fails the whole list.

Nighttroop writes `created_at`, `dateString` and `startDate` as ISO 8601 with `Z`, leaving out zero milliseconds,
which Trio reads.

## Uploads

Uploads run when the data changes (`BaseNightscoutManager+Subscribers.swift`), in batches of 100, each a `POST` with
`Content-Type: application/json` and a pretty printed body. Trio marks the records uploaded only on a 2xx (the
profile, only on a 200), and tries again later otherwise. It ignores the response body.

### Glucose: `POST /api/v1/entries.json`

When a reading arrives (`uploadGlucose`, built in `GlucoseStorage.getGlucoseNotYetUploadedToNightscout`). A CGM
reading is `{ id (UUID), sgv, direction, date (ms), dateString, unfiltered, filtered, glucose, type: "sgv" }` and a
manual one `{ id, mbg, date, dateString, type: "mbg" }`, in mg/dL. Works.

### Pump history: `POST /api/v1/treatments.json`

After each pump event (`uploadPumpHistory`, `PumpHistoryStorage.getPumpHistoryNotYetUploadedToNightscout`), each
with `created_at`, `enteredBy: "Trio"` and the event's `id`:

| Pump event | eventType | Fields |
| --- | --- | --- |
| Automatic bolus | `SMB` | `insulin` |
| Bolus | `Bolus` | `insulin` |
| Insulin not from the pump | `External Insulin` | `insulin` |
| Temp basal | `Temp Basal` | `rate`, `absolute` (U/h), `duration` (min) |
| Suspend, resume | `Note` | `notes`: `PumpSuspend` or `PumpResume` |
| Rewind | `Insulin Change` | |
| Site change | `Site Change` | |
| Pump alarm | `Announcement` | `duration: 30`, `notes` |

Nighttroop refused `SMB` and `External Insulin` as unknown types, and as one batch holds every kind, one SMB lost
the whole batch. **Fixed in this PR**: both are kept as insulin doses with their type.

### Carbs: `POST /api/v1/treatments.json`

When carbs are entered (`CarbsStorage.getCarbsNotYetUploadedToNightscout`, and the fat and protein units from
`getFPUsNotYetUploadedToNightscout`): `{ eventType: "Carb Correction", created_at, enteredBy, notes, carbs, fat,
protein, foodType, id }`. Works.

### CGM sensor start: `POST /api/v1/treatments.json`

With the glucose: `{ eventType: "Sensor Start", created_at, enteredBy, notes }`. Works.

### Temp targets: `POST /api/v1/treatments.json`

When one starts (`TempTargetsStorage.getTempTargetsNotYetUploadedToNightscout`): `{ eventType: "Temporary Target",
created_at, duration, enteredBy, notes (its name), targetTop, targetBottom }`, the targets in mg/dL with no `units`.
When one ends early its run is posted again with the same `created_at` and its real `duration`
(`getTempTargetRunsNotYetUploadedToNightscout`). Nightscout replaces the first (its v1 upsert on `created_at` and
`eventType`); Nighttroop takes it as a duplicate and keeps the first, so the temp target keeps its full length
(TODO P1).

### Overrides: `POST /api/v1/treatments.json`

When one starts (`OverrideStorage.getOverridesNotYetUploadedToNightscout`): `{ eventType: "Exercise", created_at,
duration, enteredBy, notes (its name) }`, with `duration: 43200` (30 days) when it has no end. If its duration
changed, Trio first deletes the one uploaded (below). When it ends, Trio deletes the first entry and posts the run
with its real duration (`uploadOverrideRuns`). The delete is a 405 on Nighttroop, so Trio stops there and the run is
never uploaded; an override with no end stays 30 days long (TODO P1, P2).

### Notes: `POST /api/v1/treatments.json`

`uploadNoteTreatment`: `{ eventType: "Note", created_at, enteredBy, notes }`. The connection check posts
`{ eventType: "Note", enteredBy: "Trio", notes: "Trio connected" }` (`checkConnection`). Both work.

### Device status: `POST /api/v1/devicestatus.json`

After each loop (`performUploadDeviceStatus`), with no `created_at` (Nightscout and Nighttroop use the time it
arrives):

- `device: "Trio"`
- `openaps`: `iob` (`iob`, `activity`, `basaliob`, `bolusiob`, `netbasalinsulin`, `bolusinsulin`,
  `iobWithZeroTemp`, `lastBolusTime`, `lastTemp`, `time`), `suggested` and `enacted` (oref's determination: `id`,
  `reason`, `units`, `insulinReq`, `eventualBG`, `sensitivityRatio`, `rate`, `duration`, `IOB`, `COB`, `predBGs`
  with `IOB`/`ZT`/`COB`/`UAM`, `deliverAt`, `carbsReq`, `temp`, `bg`, `reservoir`, `timestamp`, `ISF`,
  `current_target`, `TDD`, `minDelta`, `expectedDelta`, `minGuardBG`, `minPredBG`, `threshold`, `CR`, `received`;
  glucose in the user's units), `version`, `recommendedBolus`, `dosingMode`
- `pump`: `clock`, `battery` (`percent`, `voltage`, `string`, `display`), `reservoir`, `status` (`status`,
  `bolusing`, `suspended`, `timestamp`), `bolusIncrement`
- `uploader`: `battery`, `isCharging`

Works, and `openaps` comes back whole. Nighttroop keeps only the pump's `battery.percent` and `reservoir`, and not
`uploader`; Trio does not read them back, but its followers do (TODO P3).

### Profile: `POST /api/v1/profile.json`

When the user saves a schedule (`uploadProfiles`): `{ defaultProfile: "default", startDate, mills, units, enteredBy,
store: { default: { dia, carbs_hr, delay, timezone, target_low, target_high, sens, basal, carbratio, units } },
bundleIdentifier, deviceToken, isAPNSProduction, overridePresets, teamID, expirationDate }`. Each schedule entry is
`{ time: "HH:mm", value, timeAsSeconds }`. `units` is `mg/dl` or `mmol`. The push fields let LoopFollow and Trio's
remote control send it Apple push notifications; Nighttroop does not keep them. Needs exactly a 200. Works.

## Fetches

Each is a `GET` with `api-secret` and a 60 second timeout. Trio checks for a 2xx, except the glucose fetch, which
takes any status and treats a body it cannot decode as no readings.

### Carbs: `GET /api/v1/treatments.json`

Every minute when downloading is on (`FetchTreatmentsManager`, `fetchCarbs`), since a day ago:

```
find[carbs][$exists]=true
&find[$and][0][enteredBy][$ne]=Trio
&find[$and][1][enteredBy][$ne]=AndroidAPS
&find[$and][2][enteredBy][$ne]=openaps://AndroidAPS
&find[$and][3][enteredBy][$ne]=iAPS
&find[$and][4][enteredBy][$ne]=loop://iPhone
&find[created_at][$gt]=<ISO date>
```

Trio decodes each as `CarbsEntry`: `created_at` (date) and `carbs` (number) required; `_id`, `actualDate`, `fat`,
`protein`, `notes`, `enteredBy`, `isFPU`, `fpuID` optional. It stores those whose time it does not have yet.

Nighttroop ignores `find[carbs][$exists]` and the `enteredBy` filters, so it returns every treatment, and one
without `carbs` fails the whole list: Trio downloads no carbs (TODO P1). Fixing only `$exists` would be worse:
Nighttroop returns `enteredBy` as the identity's name (`pwd_phone`), not the `Trio` sent, so Trio would download
its own carbs and add them again. iAPS sends `find[enteredBy][$ne]` five times instead of `$and`.

### Temp targets: `GET /api/v1/treatments.json`

With the carbs (`fetchTempTargets`):

```
find[eventType]=Temporary+Target&find[duration][$exists]=true&<the same enteredBy filters>&find[created_at][$gt]=<ISO date>
```

Trio decodes each as `TempTarget`: `_id` (string), `created_at` and `duration` (number) required; `targetTop`,
`targetBottom` (mg/dL), `name`, `enteredBy`, `reason`, `isPreset`, `enabled`, `halfBasalTarget` optional. It skips
one whose `created_at` it already has, or whose `reason` is `Cancel`, and turns on the newest.

Works in mg/dL, with `created_at` keeping its milliseconds. As the `enteredBy` filters are ignored, Trio also gets its
own temp targets back; it skips them only when their `created_at` matches its own to the millisecond.

### Glucose, with Nightscout as the CGM: `GET /api/v1/entries/sgv.json`

On the CGM timer (`fetchLastGlucose`): `count=1600&find[dateString][$gte]=<ISO date of its last reading>`. iAPS sends
`count=500&sort$desc=dateString` too.

Trio decodes each as `BloodGlucose`: `dateString` (date) required; `_id`/`id`, `sgv` and `mbg` (integer, or a number
it truncates), `direction` (one of Nightscout's names, e.g. `Flat`, `FortyFiveUp`, `NOT COMPUTABLE`; any other
string fails the whole list), `date` (ms, or worked out from `dateString`), `unfiltered`, `filtered`, `noise`
(integer), `type`, `activationDate`, `sessionStartDate`, `transmitterID`. Works in mg/dL.

### Profile import: `GET /api/v1/profile.json?count=1`

In onboarding (`importSettings`). Trio needs a 2xx with `Content-Type: application/json`, and decodes the first
as `FetchedNightscoutProfileStore`: `_id`, `defaultProfile`, `startDate` (string), `mills` (number), `enteredBy`,
`created_at` (string) and `store` required, and takes `store.default`, which needs `dia`, `carbs_hr` (integer),
`delay`, `timezone`, `units` and the five schedules of `{ time, value, timeAsSeconds }`.

Nighttroop left out `enteredBy`, so the import failed. **Fixed in this PR**: a profile's `enteredBy` is the identity
that uploaded it.

### Announcements, iAPS only: `GET /api/v1/treatments.json`

iAPS takes remote commands (`bolus: 1`, `meal: 20,5,5`, `looping: false`, `override: name` and so on) from
announcements: `find[eventType]=Announcement&find[enteredBy]=remote&find[created_at][$gte]=<ISO date>`. It decodes
`created_at`, `enteredBy` and `notes`, all required. Trio dropped this for Apple push notifications.

Nighttroop ignores `find[enteredBy]`, so iAPS sees every announcement as a possible command, including the pump
alarms it uploads itself (which do not parse as commands) (TODO P1).

## Deletes

When the user deletes a record in Trio, it deletes it from Nightscout by query (`deleteCarbs`, `deleteInsulin`,
`deleteGlucose`, `deleteNightscoutOverride`):

- `DELETE /api/v1/treatments.json?find[id][$eq]=<Trio's id>`, for carbs and insulin
- `DELETE /api/v1/entries.json?find[$or][0][id][$eq]=<id>&find[$or][1][dateString][$eq]=<ISO date>`, for glucose
- `DELETE /api/v1/treatments.json?find[created_at][$eq]=<ISO date>&find[eventType][$eq]=Exercise`, for an override

Nighttroop deletes only by `_id` in the path (`/api/v1/treatments/<_id>`), so each is a 405 and nothing is deleted,
and Trio's `id` is not kept as the identifier (TODO P2). iAPS deletes by `find[created_at][$eq]` and
`find[carbs][$exists]` or similar, the same way.

## Socket.IO

Trio and iAPS use no Socket.IO, and no API v2 or v3.

## Quirks

- Trio rebuilds each URL from the scheme, host and port, so Nightscout under a path does not work.
- `find[eventType]=Temporary+Target` has a literal `+` for the space, which decodes to `Temporary Target`.
- iAPS percent-encodes its `enteredBy` values before `URLComponents` encodes them again, so `iAPS Watch` reaches the
  server as `iAPS%20Watch`.
- `find[created_at]` filters on Nighttroop are on when the record was uploaded, not on Nightscout's `created_at`,
  when it happened (TODO P3). For Trio's "since a day ago" this differs only for records uploaded late.

## Status

| Interaction | Nighttroop |
| --- | --- |
| `api-secret` as the SHA1 of the secret | Not supported, TODO P0 |
| Connection check, `POST` Note | Works |
| Glucose upload | Works |
| Pump history upload with `SMB` and `External Insulin` | Fixed in this PR |
| Carbs, sensor start, note uploads | Works |
| Temp target and override upload | Works |
| Temp target and override run (same `created_at`, new duration) | Not supported, TODO P1 |
| Device status upload | Works; pump clock and status and uploader dropped, TODO P3 |
| Profile upload | Works |
| Profile import | Fixed in this PR (`enteredBy`) |
| Carbs fetch (`find[carbs][$exists]`, `enteredBy` `$ne`) | Not supported, TODO P1 |
| Temp target fetch | Works in mg/dL; own ones come back, TODO P1 |
| Glucose fetch (Nightscout as CGM) | Works in mg/dL |
| iAPS announcements (`find[enteredBy]=remote`) | Not supported, TODO P1 |
| Deletes by query | Not supported, TODO P2 |
| mmol/L site or identity | Not supported by Trio's fetches; set mg/dL |
