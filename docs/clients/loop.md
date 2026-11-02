# Loop

[Loop](https://github.com/LoopKit/Loop) is an iOS automated insulin delivery app. It talks to Nightscout through its
Nightscout service plugin, [NightscoutService](https://github.com/LoopKit/NightscoutService), which uses the client
in [NightscoutKit](https://github.com/LoopKit/NightscoutKit). This page is how they talk to Nightscout, checked
against NightscoutKit `6efa02d4` (main), NightscoutService `2cef4896`, Loop `b05f7be3`, LoopKit `b5c5cfbd` and
OmniBLE `29358d59` (dev), for anyone checking Loop works with Nighttroop.

Loop only uses REST API v1. It never uses API v3, a JWT or Socket.IO. Remote commands reach Loop as Apple pushes,
not through the Nightscout API.

Files cited below are `NightscoutClient.swift` (NightscoutKit `Sources/NightscoutKit/`), the models in
NightscoutKit's `Models/`, `NightscoutService.swift` and the `Extensions/` of NightscoutService's
`NightscoutServiceKit/`.

## Setting it up

In Loop's Settings, Services, Nightscout, the user enters a **Site URL** and an **API Secret**. With Nighttroop, use
the site's Nightscout host as the URL, `https://<site>.ns.<host>`, and the PWD phone's Nightscout token as the API
secret.

Only the URL's scheme, host and port are used: its path and any query, such as `?token=`, are dropped
(`NightscoutClient.url(with:)`, line 43). So the token must go in the API Secret field.

## Authentication

Every request (`callNS`, line 519, and `checkAuth`, line 692) sends:

- `api-secret: <lowercase SHA-1 hex of the API Secret field>` (`String+Crypto.swift`)
- `Content-Type: application/json` and `Accept: application/json`, also on GETs and DELETEs

There is no `token` query, `secret` query or `Authorization` header. Nightscout accepts the SHA-1 of its
`API_SECRET`, or of an access token. Nighttroop finds the device whose token has that SHA-1 (added with xDrip+, #512): it keeps the SHA-256
of each device token's SHA-1, set when the token is made or its password is reset. A device made before this change
needs its password reset once to get one.

When the service is saved, `verifyConfiguration` (`NightscoutService.swift:105`) calls `checkAuth`:
`GET /api/v1/experiments/test`. 200 is OK. 401 shows "Unauthorized". Any other status shows the status and body.
Nighttroop answers `{"status":"ok"}`, but only for an identity that can upload treatments. So a read only device gets
a 403 rather than being saved and then failing every upload.

Loop does not call `status.json`, `verifyauth` or `versions`.

## Status codes and responses

`callNS` treats any status other than 200 as an error, including 201. Every POST needs back a JSON array with as many
objects as it sent (`postToNS`, line 479), or Loop keeps the data to upload again. From each object it reads `_id` as
a string, or uses `"NA"` if there is none. A 200 with an empty body counts as an empty array. No headers are read:
not `Last-Modified`, `ETag` or `Content-Type`.

## Dates Loop sends and parses

Loop sends dates as ISO 8601 in UTC with no fractional seconds, e.g. `2026-11-02T09:00:00Z`
(`TimeFormat.timestampStrFromDate`). It parses dates with or without fractional seconds, but they must have a `Z` or
an offset (`TimeFormat.dateFromTimestamp`). A glucose entry's `date` is a number of milliseconds, sent as a double
that can have a fraction (`GlucoseEntry.swift`).

## Glucose: `POST /api/v1/entries`

`uploadGlucoseData` (`NightscoutService.swift:317`) posts, in batches of up to 1000, an array of
`GlucoseEntry.dictionaryRepresentation`:

| Field | Type | Value |
|---|---|---|
| `date` | number | ms since the epoch, may have a fraction |
| `dateString` | string | ISO date |
| `type` | string | `sgv`, or `mbg` for a value the user entered |
| `sgv` / `mbg` | number | mg/dL, may have a fraction. No `units` field |
| `device` | string | e.g. `Dexcom G7 ...` or `loop://<iPhone name>` |
| `trend`, `direction` | int, string | 1 to 9, `DoubleUp` to `RateOutOfRange`, when known |
| `trendRate` | number | mg/dL per minute, when known |
| `isCalibration` | bool | |

A nil field is left out. Loop does not keep the returned `_id`s.

## Treatments: `POST`, `PUT` and `DELETE /api/v1/treatments`

All treatments have `created_at` and `timestamp` (the same ISO date), `enteredBy` and `eventType`, and may have
`notes`, `insulinType` and `syncIdentifier` (`NightscoutTreatment.dictionaryRepresentation`).

| What | When | `eventType` and fields |
|---|---|---|
| Carbs | `uploadCarbData`, line 197 | `Carb Correction`: `carbs` (integer g), `absorptionTime` (minutes), `foodType`, `userEnteredAt`, `userLastModifiedAt`, `syncIdentifier` (a UUID) |
| Bolus | `uploadDoseData`, line 240 | `Correction Bolus`: `type` (`normal`, or `square` for 30 minutes or longer), `insulin`, `programmed`, `unabsorbed` 0, `duration` (minutes), `automatic` |
| Temp basal | `uploadDoseData` | `Temp Basal`: `temp` `absolute`, `rate` and `absolute` (U/h), `duration` (minutes, may have a fraction), `amount`, `automatic` |
| Suspend | `uploadDoseData` | `Temp Basal` with `rate` 0 and `reason` `suspend` |
| Override | `uploadTemporaryOverrideData`, line 157 | `Temporary Override`: `_id` (its UUID), `reason` (e.g. `🏃 Running`, `Pre-Meal`, `Custom Override`), `duration` (minutes, may have a fraction) or `durationType` `indefinite`, `correctionRange` ([low, high] mg/dL), `insulinNeedsScaleFactor`, `remoteAddress` |
| Pump events | `uploadPumpEventData`, line 328 | `Site Change` for a new pod or infusion set, `Note` for an alarm, with `notes` the event's title |
| CGM events | `uploadCgmEventData`, line 391 | `Sensor Start` with `notes` `SensorID: <id>` |
| Failed remote command | `commandSourceV1(_:uploadError:)`, line 487 | `Note`, `enteredBy` the command, `notes` the error and the push's JSON |

`enteredBy` is `loop://<iPhone name>`, or `Loop` / `Loop (via remote command)` for an override.

How each is changed or removed:

- **Carbs** are posted with no `_id`. Loop keeps each returned `_id` against the carb's `syncIdentifier`
  (`ObjectIdCache`, kept a while). An edit is `PUT /api/v1/treatments` with that `_id` in the body (`modifyTreatments`,
  line 74). A delete is `DELETE /api/v1/treatments/<_id>` with no body (`deleteTreatmentsByObjectId`, line 134). The
  response bodies are not read.
- **Doses** are posted with no `_id`, so they "can be either create or update" (`DoseEntry.swift`). LoopKit keeps a
  dose that is still going, such as a temp basal or a bolus being delivered, as mutable. When it ends, or ends early,
  the same dose, with the same `syncIdentifier` and date, is updated and posted again with its final `duration` and
  `amount` or `insulin` (`InsulinDeliveryStore.addDoseEntries`, LoopKit `InsulinDeliveryStore.swift:425`). Nightscout
  replaces the first upload. A finished dose that is deleted is deleted by its cached `_id`; a mutable one is not
  deleted.
- **Overrides** are posted with `_id` the override's UUID, and posted again with the same `_id` when they change,
  such as being cancelled early with a shorter `duration`. A deleted override is `DELETE /api/v1/treatments/<UUID>`
  (`deleteTreatmentsById`, line 106).

### A dose's `syncIdentifier`

A dose's `syncIdentifier` is its pump event's raw data in hex (LoopKit `NewPumpEvent.swift:33`):

- **Omnipod** (OmniBLE and OmniKit): the hex of `"<doseType> <units> <ISO date>"` (`UnfinalizedDose.uniqueKey`,
  OmniBLE `UnfinalizedDose.swift:53`), 60 to 70 characters. For example a temp basal of 0.5 U/h is
  `74656d70426173616c20302e3520323032...`, "tempBasal 0.5 2026-11-02T08:52:00Z".
- **Medtronic** (MinimedKit): the pump history record's bytes in hex, about 16 to 26 characters, or, for a dose not
  yet read from the history, its UUID's 16 bytes in hex, 32 characters.

So an Omnipod identifier's first 36 characters are only the dose type, units and the start of the date: two temp
basals of the same rate share them ("tempBasal 0.5 202"), and two boluses of the same amount in the same month share
them ("bolus 1.0 2026-11-"). A create dedups by its identifier, so Nighttroop keeps an identifier over 36 characters
as the first 36 hex characters of its SHA-256, rather than truncating it, which would keep only the first of each.
`tests/api/nightscout/loop/treatments.http` posts two such temp basals and checks both are kept.

The same dose is posted again with the same `syncIdentifier` when it ends, so a dedup that keeps the first upload,
by identifier or by date, keeps its planned duration rather than what was delivered. Replacing the stored dose on a
repeat upload, as Nightscout does, would fix both.

## Device status: `POST /api/v1/devicestatus`

`uploadDosingDecisionData` (`NightscoutService.swift:276`) posts an array of `DeviceStatus.dictionaryRepresentation`
for each bolus decision, and each automatic dose decision with it:

- `device` `loop://<iPhone name>`, `created_at`
- `loop`: `name`, `version`, `timestamp`, `iob {iob, timestamp}`, `cob {cob, timestamp}`,
  `predicted {startDate, values}` (mg/dL), `automaticDoseRecommendation {timestamp, tempBasalAdjustment {rate,
  duration}, bolusVolume}`, `recommendedBolus`, `enacted {rate, duration, timestamp, received, bolusVolume}`,
  `failureReason`
- `pump`: `clock`, `pumpID`, `manufacturer`, `model`, `battery {percent}`, `suspended`, `bolusing`, `reservoir`,
  `secondsFromGMT`, `reservoir_display_override`, `reservoir_level_override`
- `uploader`: `name`, `timestamp`, `battery`, `isCharging`
- `override`: `timestamp`, `active`, and when active `name`, `currentCorrectionRange {minValue, maxValue}`,
  `duration` (seconds left, not minutes), `multiplier`

Loop does not read any of it back. Nighttroop keeps the `loop` object, the pump's battery and reservoir and the
uploader's battery; the rest is dropped.

## Profile: `POST /api/v1/profile` and `GET /api/v1/profile/current`

`uploadSettingsData` (`NightscoutService.swift:361`) posts an array of `ProfileSet.dictionaryRepresentation` to
`/api/v1/profile` (singular), with no `_id`: `defaultProfile`, `startDate`, `mills` (a string), `units`,
`enteredBy` `Loop`, `isAPNSProduction`, `loopSettings` (`dosingEnabled`, `overridePresets`, `minimumBGGuard`,
`preMealTargetRange`, `maximumBasalRatePerHour`, `maximumBolus`, `deviceToken`, `bundleIdentifier`,
`dosingStrategy`, `scheduleOverride`), and `store` with one profile: `dia` (hours), `carbs_hr` and `delay` as `"0"`,
`timezone` as `ETC/GMT±N`, `units`, and `sens`, `carbratio`, `basal`, `target_low` and `target_high` as
`[{time, value, timeAsSeconds}]`.

When Loop is set up from Nightscout, `fetchStoredTherapySettings` (line 370) gets `/api/v1/profile/current`, which
it reads as one object, not an array (`fetchCurrentProfile`, line 166). `ProfileSet(rawValue:)` needs `_id`,
`startDate`, `units`, `enteredBy`, `defaultProfile`, `store` and `loopSettings` with `dosingEnabled` and
`overridePresets`. Each profile needs `timezone` in the `ETC/GMT±N` form it sent, `dia`, and each schedule's
`timeAsSeconds` and `value`.

`updateProfile` (`PUT /api/v1/profile`) is in NightscoutKit but not used.

## NightscoutKit's reads (not used by Loop itself)

NightscoutKit has reads that Loop does not call, but other apps built on it may:

- `fetchGlucose`: `GET /api/v1/entries?find[dateString][$gte]=<ISO>&find[dateString][$lte]=<ISO>&count=<n>`. Needs
  `_id`, `date` (number) and `sgv` or `mbg`, which it reads as mg/dL.
- `fetchTreatments`: `GET /api/v1/treatments?find[created_at][$gte]=...&find[created_at][$lte]=...&count=<n>`. Needs
  `_id`, `eventType`, `timestamp` and `enteredBy`, and per type the fields it posts (e.g. a bolus's `type`,
  `insulin`, `programmed` and `duration`; a temp basal's `rate`, `temp` and `duration`).
- `fetchDeviceStatus`: `GET /api/v1/devicestatus?find[created_at][$gte]=...&find[created_at][$lte]=...&count=<n>`.
  Needs `_id`, `created_at`, `device`, and reads `loop`.
- `fetchProfiles`: `GET /api/v1/profiles?find[startDate][$gte]=...&find[startDate][$lte]=...&count=<n>`.
- `startOverride`, `cancelOverride`, `deliverBolus` and `deliverCarbs`: `POST /api/v2/notifications/loop` with
  string values (`eventType` `Temporary Override`, `Temporary Override Cancel`, `Remote Bolus Entry` or
  `Remote Carbs Entry`, with `otp`), which Nightscout sends to Loop as an Apple push.

## Quirks

- The site URL's path and query are dropped, so a token cannot be put in the URL.
- Any status but 200 is an error, so a 201 from a POST makes Loop upload again.
- A POST response must have exactly as many objects as were sent, or Loop uploads them again.
- An override's `_id` is a UUID with upper case letters and dashes, not a Mongo ObjectId.
- A dose and an override are posted again, with the same date, when they change.
- An override's `duration` in the device status is seconds; in the treatment it is minutes.
- `carbs` is read back as an integer, and `absorptionTime` is needed to read a carb entry back.

## Nighttroop status

| Interaction | Status |
|---|---|
| `api-secret` as the SHA-1 of a device's token | Works since #512 (it was a 401 for everything) |
| `GET /api/v1/experiments/test` | Fixed in this PR (was a 400, "invalid collection"); 403 for a read only device |
| `POST /api/v1/entries`, as many objects back, `_id` | Works |
| `POST /api/v1/treatments` carbs, doses, pump and CGM events, notes | Works |
| `PUT /api/v1/treatments` with `_id` in the body | Works |
| `DELETE /api/v1/treatments/<_id>` | Works |
| `POST` a `Temporary Override` | Fixed in this PR (was a 400, unknown event type); kept as a temp target, indefinite as a week |
| An override's UUID `_id`, posted again when it changes, deleted by it | Fixed in this PR: a v1 POST with an `_id` replaces the one with it, or is created with it |
| A dose posted again when it ends, with its final duration or amount | Not supported: the first upload is kept. TODO P1 |
| Omnipod `syncIdentifier`s once creates dedup by identifier | Fixed in this PR: an identifier over 36 characters is kept as a hash of it, so they no longer share their first 36 characters; test in `loop/treatments.http` |
| `POST /api/v1/devicestatus`, `loop` object | Works |
| Device status `override`, `uploader`, pump `clock`, `model`, `suspended`, `bolusing` | Not supported: dropped. TODO P2 (Loop does not read them back) |
| `POST /api/v1/profile`, `loopSettings` and `store` | Works |
| `GET /api/v1/profile/current` as one object with `enteredBy` | Fixed in this PR (was an array, and had no `enteredBy`) |
| NightscoutKit `find[created_at]` on treatments and device statuses | Fixed in this PR (it was the upload time, not when they happened) |
| NightscoutKit `fetchGlucose` | Works on an mg/dL site; on an mmol/L site `sgv` is in mmol/L. TODO P2 |
| NightscoutKit `fetchTreatments` reading bolus and temp basal fields and `timestamp` | Not supported: no `timestamp`, `programmed` or `temp`. TODO P3 |
| NightscoutKit `fetchProfiles` | Works |
| `POST /api/v2/notifications/loop` (remote commands) | Not supported. TODO P2 |

The tests are in `tests/api/nightscout/loop/`: `auth.http`, `entries.http`, `treatments.http`, `devicestatus.http`
and `profile.http`. They send Loop's headers and bodies, and check what Loop reads back.
