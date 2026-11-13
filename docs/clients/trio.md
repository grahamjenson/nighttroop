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

Nighttroop takes a device's token in `api-secret`, or its SHA1 (40 hex digits), which it finds by a digest of the
SHA1 kept for each token (not for an admin or an identity with TOTP), since #512. So Trio signs in with a device's
token as its API secret. The site's Add Device has Trio as a client (#524), which makes a PWD Phone device (#527) in
mg/dL, with a setup guide. The tests send the token's SHA1 in `api-secret`, as Trio does. Without a secret, Trio's connection check is a `GET /api/v1/treatments.json`, which Nighttroop
refuses with a 401, as it needs a token.

## Units

Trio reads glucose (`sgv`) as mg/dL and decodes it as an integer, and temp targets' `targetTop`/`targetBottom` as
mg/dL. The Nightscout API always gives both in mg/dL, whatever the identity's or site's units, so a reading of 120
mg/dL comes back as 120 on an mmol/L site too. The profile is fine either way: Trio uploads it in the user's units, labelled `mg/dl` or `mmol`, and on import works
out mmol/L from `units` or from targets of 39 or less.

## Dates and types Trio reads back

Trio's JSON decoder (`JSONCoding`, `Formatters.swift` `customISO8601`) reads a date only from a string in ISO 8601
with a time zone, with or without fractional seconds (`2026-11-01T12:00:00Z`, `2026-11-01T12:00:00.123Z`). A number,
or a date without a zone, fails. It writes dates with milliseconds and `Z`. Numbers are Swift `Decimal` or `Int`;
one field of the wrong type, or one required field missing, fails the whole list.

Nighttroop writes `created_at`, `timestamp`, `dateString` and `startDate` as Nightscout does, as JavaScript's
`toISOString` writes them: with three fractional digits, e.g. `2026-11-01T12:00:00.000Z`, so `deleteGlucose` finds a
reading by the `.SSS` `dateString` it sent. Trio reads them.

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
the whole batch. Fixed in #509: both are kept as insulin doses with their type.

### Carbs: `POST /api/v1/treatments.json`

When carbs are entered (`CarbsStorage.getCarbsNotYetUploadedToNightscout`, and the fat and protein units from
`getFPUsNotYetUploadedToNightscout`): `{ eventType: "Carb Correction", created_at, enteredBy, notes, carbs, fat,
protein, foodType, id }`. Works. Its `notes` and `foodType` are the same text, which can end with a new line, and
both come back as sent.

### CGM sensor start: `POST /api/v1/treatments.json`

With the glucose: `{ eventType: "Sensor Start", created_at, enteredBy, notes }`. Works.

### Temp targets: `POST /api/v1/treatments.json`

When one starts (`TempTargetsStorage.getTempTargetsNotYetUploadedToNightscout`): `{ eventType: "Temporary Target",
created_at, duration, enteredBy, notes (its name), targetTop, targetBottom }`, the targets in mg/dL with no `units`.
When one ends early its run is posted again with the same `created_at` and its real `duration`
(`getTempTargetRunsNotYetUploadedToNightscout`). As in Nightscout (its v1 upsert on `created_at` and `eventType`),
it replaces the first, so the temp target keeps its real length.

### Overrides: `POST /api/v1/treatments.json`

When one starts (`OverrideStorage.getOverridesNotYetUploadedToNightscout`): `{ eventType: "Exercise", created_at,
duration, enteredBy, notes (its name) }`, with `duration: 43200` (30 days) when it has no end. If its duration
changed, Trio first deletes the one uploaded (below). When it ends, Trio deletes the first entry and posts the run
with its real duration (`uploadOverrideRuns`). Both work: the delete by query soft deletes the first, and a run
posted with the same `created_at` while the first is kept replaces it.

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

Works, and `openaps` comes back whole. The pump comes back with its `clock`, `reservoir`, `battery` (`percent`,
`voltage`, `string`, `display`), `status` (`status`, `bolusing`, `suspended`, `timestamp`) and `bolusIncrement`
while the JSON it was sent in is kept (then only the clock, battery percent and reservoir), and `uploader` with its
`battery` and `isCharging`. Trio does not read them back, but its followers do. When iOS does not know the phone's
battery Trio sends `uploader.battery` -100 (its `batteryLevel` -1 x 100); the device status is kept without it.

### Profile: `POST /api/v1/profile.json`

When the user saves a schedule (`uploadProfiles`): `{ defaultProfile: "default", startDate, mills, units, enteredBy,
store: { default: { dia, carbs_hr, delay, timezone, target_low, target_high, sens, basal, carbratio, units } },
bundleIdentifier, deviceToken, isAPNSProduction, overridePresets, teamID, expirationDate }`. Each schedule entry is
`{ time: "HH:mm", value, timeAsSeconds }`. `units` is `mg/dl` or `mmol`. The push fields let LoopFollow and Trio's
remote control send it Apple push notifications; Nighttroop registers `deviceToken`, `bundleIdentifier` and
`isAPNSProduction` as a `trio` push registration for the identity that uploaded it (`/settings/push-registrations`).
Needs exactly a 200. Works.

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

Works: Nighttroop gives only treatments with `carbs`, as one without fails the whole list, and leaves out those
whose `enteredBy`, kept as sent, is one of the `$ne`s, so Trio does not download its own carbs and add them again.
iAPS sends `find[enteredBy][$ne]` five times instead of `$and`, encoded twice (`iAPS%20Watch`), which works too.

### Temp targets: `GET /api/v1/treatments.json`

With the carbs (`fetchTempTargets`):

```
find[eventType]=Temporary+Target&find[duration][$exists]=true&<the same enteredBy filters>&find[created_at][$gt]=<ISO date>
```

Trio decodes each as `TempTarget`: `_id` (string), `created_at` and `duration` (number) required; `targetTop`,
`targetBottom` (mg/dL), `name`, `enteredBy`, `reason`, `isPreset`, `enabled`, `halfBasalTarget` optional. It skips
one whose `created_at` it already has, or whose `reason` is `Cancel`, and turns on the newest.

Works in mg/dL, with `created_at` keeping its milliseconds, and without Trio's own temp targets.

### Glucose, with Nightscout as the CGM: `GET /api/v1/entries/sgv.json`

On the CGM timer (`fetchLastGlucose`): `count=1600&find[dateString][$gte]=<ISO date of its last reading>`. iAPS sends
`count=500&sort$desc=dateString` too.

Trio decodes each as `BloodGlucose`: `dateString` (date) required; `_id`/`id`, `sgv` and `mbg` (integer, or a number
it truncates), `direction` (one of Nightscout's names, e.g. `Flat`, `FortyFiveUp`, `NOT COMPUTABLE`; any other
string fails the whole list), `date` (ms, or worked out from `dateString`), `unfiltered`, `filtered`, `noise`
(integer), `type`, `activationDate`, `sessionStartDate`, `transmitterID`. Works in mg/dL. Nighttroop gives each
`direction` as one of those names, from any case or arrow, `NOT COMPUTABLE` for one it does not know, and `noise` as
a whole number, so one uploader's reading cannot stop Trio getting glucose.

### Profile import: `GET /api/v1/profile.json?count=1`

In onboarding (`importSettings`). Trio needs a 2xx with `Content-Type: application/json`, and decodes the first
as `FetchedNightscoutProfileStore`: `_id`, `defaultProfile`, `startDate` (string), `mills` (number), `enteredBy`,
`created_at` (string) and `store` required, and takes `store.default`, which needs `dia`, `carbs_hr` (integer),
`delay`, `timezone`, `units` and the five schedules of `{ time, value, timeAsSeconds }`.

Nighttroop left out `enteredBy`, so the import failed. Fixed in #509: a profile's `enteredBy` is the identity
that uploaded it.

### Announcements, iAPS only: `GET /api/v1/treatments.json`

iAPS takes remote commands (`bolus: 1`, `meal: 20,5,5`, `looping: false`, `override: name` and so on) from
announcements: `find[eventType]=Announcement&find[enteredBy]=remote&find[created_at][$gte]=<ISO date>`. It decodes
`created_at`, `enteredBy` and `notes`, all required. Trio dropped this for Apple push notifications.

Works: only announcements `enteredBy` `remote` come back, not the pump alarms iAPS uploads itself.

## Deletes

When the user deletes a record in Trio, it deletes it from Nightscout by query (`deleteCarbs`, `deleteInsulin`,
`deleteGlucose`, `deleteNightscoutOverride`):

- `DELETE /api/v1/treatments.json?find[id][$eq]=<Trio's id>`, for carbs and insulin
- `DELETE /api/v1/entries.json?find[$or][0][id][$eq]=<id>&find[$or][1][dateString][$eq]=<ISO date>`, for glucose
- `DELETE /api/v1/treatments.json?find[created_at][$eq]=<ISO date>&find[eventType][$eq]=Exercise`, for an override

Works: Trio's `id` is kept as the identifier, so `find[id][$eq]` finds it, and a delete by query soft deletes what
that search finds, at most 1000 a request, each `find[$or]` alternative a search of its own. Each alternative needs
a date or an id, and a filter Nighttroop does not take for a delete is a 400, rather than delete more than was asked
for. iAPS deletes by `find[created_at][$eq]` and `find[carbs][$exists]` or similar, the same way.

## Socket.IO

Trio and iAPS use no Socket.IO, and no API v2 or v3.

## Quirks

- Trio rebuilds each URL from the scheme, host and port, so Nightscout under a path does not work.
- `find[eventType]=Temporary+Target` has a literal `+` for the space, which decodes to `Temporary Target`.
- iAPS percent-encodes its `enteredBy` values before `URLComponents` encodes them again, so `iAPS Watch` reaches the
  server as `iAPS%20Watch`.
- `find[created_at]` filters are on when the record happened, as in Nightscout (#514), and `srvCreated` on when it
  was uploaded.

## Status

| Interaction | Nighttroop |
| --- | --- |
| `api-secret` as the SHA1 of the secret | Works (#512) |
| Connection check, `POST` Note | Works |
| Glucose upload | Works |
| Pump history upload with `SMB` and `External Insulin` | Works (#509) |
| Carbs, sensor start, note uploads | Works |
| Temp target and override upload | Works |
| Temp target and override run (same `created_at`, new duration) | Fixed in this PR: it replaces the first, as in Nightscout |
| Device status upload | Works; pump `status.bolusing`, `suspended`, battery `string`, `display` and `uploader.isCharging` fixed in this PR |
| Profile upload | Works |
| Profile import | Works (`enteredBy`, #509) |
| Carbs fetch (`find[carbs][$exists]`, `enteredBy` `$ne`) | Fixed in this PR |
| Temp target fetch | Works in mg/dL; own ones left out since this PR |
| Glucose fetch (Nightscout as CGM) | Works in mg/dL |
| iAPS announcements (`find[enteredBy]=remote`) | Fixed in this PR |
| Deletes by query | Fixed in this PR (they were a 405) |
| mmol/L site or identity | Not supported by Trio's fetches; a device added for Trio is in mg/dL |
