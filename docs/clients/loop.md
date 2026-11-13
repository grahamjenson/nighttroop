# Loop

## Overview

[Loop](https://github.com/LoopKit/Loop) is an iOS automated insulin delivery app. This page tells how Loop uses the
Nightscout API, and the Nighttroop status of each interaction.

- Loop uses [NightscoutService](https://github.com/LoopKit/NightscoutService), its Nightscout service plugin, and
  the client in [NightscoutKit](https://github.com/LoopKit/NightscoutKit).
- Checked against NightscoutKit `6efa02d4` (main), NightscoutService `2cef4896`, Loop `b05f7be3`, LoopKit `b5c5cfbd`
  and OmniBLE `29358d59` (dev).
- REST API v1 only: no API v3, JWT or Socket.IO. Remote commands reach Loop as Apple pushes.
- Files: `NightscoutClient.swift` (NightscoutKit `Sources/NightscoutKit/`), NightscoutKit's `Models/`, and
  `NightscoutService.swift` and `Extensions/` in NightscoutService's `NightscoutServiceKit/`.
- Tests: `tests/api/nightscout/loop/` (`auth.http`, `entries.http`, `treatments.http`, `devicestatus.http`,
  `profile.http`). They send Loop's headers and bodies, and check what Loop reads back.

## Set up

1. On the site's Users & Devices, add a device with the Loop client.
   - It is the PWD phone, in mg/dL. Nighttroop shows these steps with its URL and token (#524, #527).
2. In Loop, open Settings, Services, Nightscout.
3. In **Site URL**, enter the site's Nightscout host, `https://<site>.ns.<host>`.
4. In **API Secret**, enter the PWD phone's Nightscout token.

**Note:** Put the token in **API Secret**. Loop uses only the URL's scheme, host and port, and drops the path and
query, such as `?token=` (`NightscoutClient.url(with:)`, line 43).

## Authentication

- Every request (`callNS`, line 519; `checkAuth`, line 692) sends:
  - `api-secret: <lowercase SHA-1 hex of the API Secret field>` (`String+Crypto.swift`).
  - `Content-Type: application/json` and `Accept: application/json`, also on GETs and DELETEs.
  - No `token` query, `secret` query or `Authorization` header.
- Nightscout accepts the SHA-1 of its `API_SECRET`, or of an access token.
- Nighttroop finds the device whose token has that SHA-1 (added with xDrip+, #512).
  - It keeps the SHA-256 of each device token's SHA-1, set when the token is made or its password is reset.
  - A device made before this change needs one password reset.
- On save, `verifyConfiguration` (`NightscoutService.swift:105`) calls `checkAuth`: `GET /api/v1/experiments/test`.
  - 200 is OK. 401 shows "Unauthorized". Other statuses show the status and body.
  - Nighttroop returns `{"status":"ok"}` only for an identity that can upload readings (`api:entries:create`). A read
    only device gets a 401, so Loop does not save it and then fail every upload.
- Loop does not call `status.json`, `verifyauth` or `versions`.

## Responses and dates

- `callNS`: any status other than 200 is an error, 201 too. Loop uploads the data again.
- A POST response must be a JSON array with as many objects as sent (`postToNS`, line 479). Else Loop uploads again.
- Loop reads each object's `_id` as a string, or uses `"NA"`. A 200 with an empty body is an empty array.
- Loop reads no headers (`Last-Modified`, `ETag`, `Content-Type`).
- Loop sends dates in UTC ISO 8601 with no fractional seconds, for example `2026-11-02T09:00:00Z`
  (`TimeFormat.timestampStrFromDate`).
- Loop parses dates with or without fractional seconds, with a `Z` or an offset (`TimeFormat.dateFromTimestamp`).
- A glucose entry's `date` is milliseconds, a double that can have a fraction (`GlucoseEntry.swift`).

## Uploads

### Batches

- One POST holds all data since the last successful upload, up to 1000 records (`glucoseDataLimit`,
  `doseDataLimit`, `carbDataLimit`, `pumpEventDataLimit`).
- Loop moves on only after a 200 (LoopKit `RemoteDataServicesManager.uploadGlucoseData`).
- Nighttroop takes a v1 POST of up to 5000 records, for the backlog of a Loop that already ran or was offline for
  hours. Each 100 records past the first cost a write against the site's and identity's rate limits.

### Glucose: `POST /api/v1/entries`

`uploadGlucoseData` (`NightscoutService.swift:317`) posts `GlucoseEntry.dictionaryRepresentation` arrays. Nil fields
are left out. Loop does not keep the returned `_id`s.

| Field | Type | Value |
|---|---|---|
| `date` | number | ms since the epoch, may have a fraction |
| `dateString` | string | ISO date |
| `type` | string | `sgv`, or `mbg` for a value the user entered |
| `sgv` / `mbg` | number | mg/dL, may have a fraction. No `units` field |
| `device` | string | for example `Dexcom G7 ...` or `loop://<iPhone name>` |
| `trend`, `direction` | int, string | 1 to 9, `DoubleUp` to `RateOutOfRange`, when known |
| `trendRate` | number | mg/dL per minute, when known |
| `isCalibration` | bool | |

### Treatments: `POST`, `PUT` and `DELETE /api/v1/treatments`

- All: `created_at` and `timestamp` (the same ISO date), `enteredBy`, `eventType`. Optional: `notes`, `insulinType`,
  `syncIdentifier` (`NightscoutTreatment.dictionaryRepresentation`).
- `enteredBy`: `loop://<iPhone name>`, or `Loop` / `Loop (via remote command)` for an override.

| What | Sent by | `eventType` and fields |
|---|---|---|
| Carbs | `uploadCarbData`, line 197 | `Carb Correction`: `carbs` (integer g), `absorptionTime` (minutes), `foodType`, `userEnteredAt`, `userLastModifiedAt`, `syncIdentifier` (a UUID) |
| Bolus | `uploadDoseData`, line 240 | `Correction Bolus`: `type` (`normal`, or `square` for 30 minutes or longer), `insulin`, `programmed`, `unabsorbed` 0, `duration` (minutes), `automatic` |
| Temp basal | `uploadDoseData` | `Temp Basal`: `temp` `absolute`, `rate` and `absolute` (U/h), `duration` (minutes, may have a fraction), `amount`, `automatic` |
| Suspend | `uploadDoseData` | `Temp Basal`, `rate` 0, `reason` `suspend` |
| Override | `uploadTemporaryOverrideData`, line 157 | `Temporary Override`: `_id` (its UUID), `reason` (for example `🏃 Running`, `Pre-Meal`, `Custom Override`), `duration` (minutes, may have a fraction) or `durationType` `indefinite`, `correctionRange` ([low, high] mg/dL), `insulinNeedsScaleFactor`, `remoteAddress` |
| Pump events | `uploadPumpEventData`, line 328 | `Site Change` for a new pod or infusion set; `Note` for an alarm, `notes` its title |
| CGM events | `uploadCgmEventData`, line 391 | `Sensor Start`, `notes` `SensorID: <id>` |
| Failed remote command | `commandSourceV1(_:uploadError:)`, line 487 | `Note`: `enteredBy` the command, `notes` the error and the push's JSON |

Nighttroop keeps these and gives them back as sent, through API v1 and v3. Not `userLastModifiedAt`.

| Field | Kept with | Why |
|---|---|---|
| A bolus's `automatic` | its insulin dose | LoopFollow shows an automatic bolus as an SMB |
| A temp basal's `automatic` | the temp basal | It tells a temp basal Loop set from one the user set |
| `syncIdentifier` (up to 256 characters) | its event group | |
| `insulinType` of a bolus or temp basal | its insulin dose or temp basal | |
| A bolus's `unabsorbed` | its insulin dose | |
| A carb entry's `userEnteredAt` | its food intake (given back in UTC) | |

### Edits and deletes

- **Carbs**: posted with no `_id`. Loop keeps each returned `_id` against the carb's `syncIdentifier`
  (`ObjectIdCache`, kept a while).
  - Edit: `PUT /api/v1/treatments` with that `_id` in the body (`modifyTreatments`, line 74).
  - Delete: `DELETE /api/v1/treatments/<_id>`, no body (`deleteTreatmentsByObjectId`, line 134).
  - Loop does not read the response bodies.
- **Doses**: posted with no `_id`, so they "can be either create or update" (`DoseEntry.swift`).
  - LoopKit keeps a dose that is still going (a temp basal, a bolus in delivery) as mutable.
  - When it ends, or ends early, Loop posts it again: same `syncIdentifier` and date, final `duration` and `amount`
    or `insulin` (`InsulinDeliveryStore.addDoseEntries`, LoopKit `InsulinDeliveryStore.swift:425`).
  - As in Nightscout, this replaces the stored dose and keeps its `_id`. A dose ended early keeps what was
    delivered, not what was planned. A dose sent again unchanged changes nothing.
  - Delete: a finished dose by its cached `_id`. A mutable dose is not deleted.
- **Overrides**: posted with `_id` the override's UUID (upper case, with dashes, not a Mongo ObjectId).
  - Posted again with that `_id` when it changes, for example cancelled early with a shorter `duration`.
  - Delete: `DELETE /api/v1/treatments/<UUID>` (`deleteTreatmentsById`, line 106).

### A dose's `syncIdentifier`

The pump event's raw data in hex (LoopKit `NewPumpEvent.swift:33`):

- **Omnipod** (OmniBLE, OmniKit): hex of `"<doseType> <units> <ISO date>"` (`UnfinalizedDose.uniqueKey`, OmniBLE
  `UnfinalizedDose.swift:53`), 60 to 70 characters. A 0.5 U/h temp basal is `74656d70426173616c20302e3520323032...`,
  "tempBasal 0.5 2026-11-02T08:52:00Z".
- **Medtronic** (MinimedKit): the pump history record's bytes in hex, about 16 to 26 characters. A dose not yet read
  from the history: its UUID's 16 bytes in hex, 32 characters.

**Caution:** The first 36 characters of an Omnipod identifier hold only the dose type, units and start of the date.
Two temp basals of the same rate share them ("tempBasal 0.5 202"), and so do two boluses of the same amount in one
month ("bolus 1.0 2026-11-"). Creates dedup by identifier, so truncation keeps only the first of each.

- Nighttroop keeps an identifier over 36 characters as the first 36 hex characters of its SHA-256.
- `tests/api/nightscout/loop/treatments.http` posts two such temp basals and checks that both are kept.

### Device status: `POST /api/v1/devicestatus`

`uploadDosingDecisionData` (`NightscoutService.swift:276`) posts `DeviceStatus.dictionaryRepresentation` arrays, for
each bolus decision and each automatic dose decision. Loop reads none of it back.

| Object | Fields |
|---|---|
| top level | `device` `loop://<iPhone name>`, `created_at` |
| `loop` | `name`, `version`, `timestamp`, `iob {iob, timestamp}`, `cob {cob, timestamp}`, `predicted {startDate, values}` (mg/dL), `automaticDoseRecommendation {timestamp, tempBasalAdjustment {rate, duration}, bolusVolume}`, `recommendedBolus`, `enacted {rate, duration, timestamp, received, bolusVolume}`, `failureReason` |
| `pump` | `clock`, `pumpID`, `manufacturer`, `model`, `battery {percent}`, `suspended`, `bolusing`, `reservoir`, `secondsFromGMT`, `reservoir_display_override`, `reservoir_level_override` |
| `uploader` | `name`, `timestamp`, `battery`, `isCharging` |
| `override` | `timestamp`, `active`; when active, `name`, `currentCorrectionRange {minValue, maxValue}`, `duration` (seconds left; minutes in the treatment), `multiplier` |

Nighttroop keeps:

- `loop` and `override` (#516).
- Pump: battery, reservoir, `clock`, `manufacturer`, `model` (#516), `pumpID`, `suspended`, `bolusing` (followers
  show insulin delivery suspended from these).
- Uploader: battery, `isCharging`, `name`, `timestamp`.

It gives these back as sent, also after it clears the pump as sent from all but the device's newest status (#514).
`secondsFromGMT` comes back only while that is kept. `reservoir_display_override` and `reservoir_level_override` are
dropped.

### Profile: `POST /api/v1/profile`

`uploadSettingsData` (`NightscoutService.swift:361`) posts `ProfileSet.dictionaryRepresentation` arrays to
`/api/v1/profile` (singular), with no `_id`:

- Top level: `defaultProfile`, `startDate`, `mills` (a string), `units`, `enteredBy` `Loop`, `isAPNSProduction`.
- `loopSettings`: `dosingEnabled`, `overridePresets`, `minimumBGGuard`, `preMealTargetRange`,
  `maximumBasalRatePerHour`, `maximumBolus`, `deviceToken`, `bundleIdentifier`, `dosingStrategy`, `scheduleOverride`.
- `store`, one profile: `dia` (hours), `carbs_hr` and `delay` as `"0"`, `timezone` as `ETC/GMT±N`, `units`, and
  `sens`, `carbratio`, `basal`, `target_low`, `target_high` as `[{time, value, timeAsSeconds}]`.

Remote commands:

- Nighttroop registers `loopSettings.deviceToken`, `bundleIdentifier` and `isAPNSProduction` as a `loop` push
  registration for the uploading identity (`/settings/push-registrations`).
- It pushes remote commands to the newest one, with the site's APNs key for the bundle ID.
- The user adds the key in Push Certificates: Loop is built with the user's own Apple developer team.

## Downloads

### Profile: `GET /api/v1/profile/current`

- When Loop is set up from Nightscout, `fetchStoredTherapySettings` (line 370) reads it as one object, not an array
  (`fetchCurrentProfile`, line 166).
- `ProfileSet(rawValue:)` needs `_id`, `startDate`, `units`, `enteredBy`, `defaultProfile`, `store`, and
  `loopSettings` with `dosingEnabled` and `overridePresets`.
- Each profile needs `timezone` in the `ETC/GMT±N` form sent, `dia`, and each schedule's `timeAsSeconds` and `value`.
- `updateProfile` (`PUT /api/v1/profile`) is in NightscoutKit, not used.

### NightscoutKit reads (not used by Loop)

Other apps built on NightscoutKit can use these:

| Function | Request | Needs |
|---|---|---|
| `fetchGlucose` | `GET /api/v1/entries?find[dateString][$gte]=<ISO>&find[dateString][$lte]=<ISO>&count=<n>` | `_id`, `date` (number), `sgv` or `mbg` (read as mg/dL) |
| `fetchTreatments` | `GET /api/v1/treatments?find[created_at][$gte]=...&find[created_at][$lte]=...&count=<n>` | `_id`, `eventType`, `timestamp`, `enteredBy`, and the fields it posts for each type (a bolus's `type`, `insulin`, `programmed`, `duration`; a temp basal's `rate`, `temp`, `duration`) |
| `fetchDeviceStatus` | `GET /api/v1/devicestatus?find[created_at][$gte]=...&find[created_at][$lte]=...&count=<n>` | `_id`, `created_at`, `device`; reads `loop` |
| `fetchProfiles` | `GET /api/v1/profiles?find[startDate][$gte]=...&find[startDate][$lte]=...&count=<n>` | |
| `startOverride`, `cancelOverride`, `deliverBolus`, `deliverCarbs` | `POST /api/v2/notifications/loop`, string values: `eventType` `Temporary Override`, `Temporary Override Cancel`, `Remote Bolus Entry` or `Remote Carbs Entry`, with `otp` | Nightscout sends it to Loop as an Apple push |

## Quirks

- Loop reads `carbs` back as an integer, and needs `absorptionTime` to read a carb entry back.
- Also see the notes on the URL ([Set up](#set-up)), statuses ([Responses and dates](#responses-and-dates)) and
  posts again ([Edits and deletes](#edits-and-deletes)).

## Status on Nighttroop

Test paths start at `tests/api/nightscout/`.

| Interaction | Status |
|---|---|
| **Authentication** | |
| `api-secret` as the SHA-1 of a device's token | Works since #512 (was a 401 for all) |
| `GET /api/v1/experiments/test` | Works since #513 (was a 400, "invalid collection"). 401, "Unauthorized", for a read only device |
| **Entries** | |
| `POST /api/v1/entries`, as many objects back, `_id` | Works |
| A POST of over 100 entries or treatments | Works: up to 5000 (was a 400 over 100, and Loop never uploaded again) |
| **Treatments** | |
| `POST /api/v1/treatments`: carbs, doses, pump and CGM events, notes | Works |
| `PUT /api/v1/treatments` with `_id` in the body | Works |
| A carb entry edited to another time (`PUT` with its `_id`); an override's start time edited (posted again with its `_id`) | Fixed: moved to the new time (was a 400, "Field date cannot be modified by the client", sent again without end). `loop/treatments.http` |
| `DELETE /api/v1/treatments/<_id>` | Works |
| `POST` a `Temporary Override` | Works since #515 (was a 400, unknown event type). Kept as a temp target; indefinite as a week, given back with `durationType` `indefinite` |
| An override's UUID `_id`: posted again when it changes, deleted by it | Works since #515: a v1 POST with an `_id` replaces the record with it, or creates it |
| `Temporary Override` with `durationType` `indefinite` and `duration` 0, or of 0 minutes | Fixed: indefinite, as with no `duration` (was 0 minutes). `loop/override_indefinite.http` |
| A dose posted again when it ends | Fixed: replaces the first upload. `loop/treatments.http` |
| A dose's or override's `duration` with a fraction of a minute | Fixed: kept to the millisecond in `durationInMilliseconds`, given back with the fraction (a 1.67 minute bolus came back as 1). `loop/treatments.http` |
| Omnipod `syncIdentifier`s | Works since #515: kept as a hash. `loop/treatments.http` |
| `syncIdentifier`, `insulinType`, `userEnteredAt`, `unabsorbed`, a temp basal's `automatic` | Fixed (were dropped). `loop/treatments.http` |
| `utcOffset` of treatments, device statuses, profiles (times in UTC) | Fixed: the site timezone's offset; a device status uses its pump's `secondsFromGMT` (were 0). `loop/utc_offset.http` |
| **Device status** | |
| `POST /api/v1/devicestatus`, `loop` object | Works |
| `override`, uploader battery, pump `clock`, `manufacturer`, `model` | Works since #516 |
| Pump `pumpID`, `suspended`, `bolusing` | Fixed: kept after the status as sent is cleared (were dropped). `loop/devicestatus.http` |
| Uploader `name`, `timestamp`, `isCharging` | Works, also after the status as sent is cleared |
| **Profile** | |
| `POST /api/v1/profile`, `loopSettings`, `store` | Works |
| `loopSettings.scheduleOverride`, the active override | Fixed: kept as an override preset, `targetRange` in mg/dL (was dropped). `loop/profile.http` |
| `GET /api/v1/profile/current` as one object with `enteredBy` | Works since #515 (was an array, no `enteredBy`) |
| A profile's `enteredBy` `Loop`, given back as sent | Fixed (was the uploading identity, for example `pwd_phone`) |
| **NightscoutKit reads and remote commands** | |
| `find[created_at]` on treatments and device statuses | Works since #514 (was the upload time, not the event time) |
| `fetchGlucose` | Works for an identity in mg/dL, as a device added for Loop is (#524). In mmol/L, `sgv` is mmol/L (TODO P1) |
| `fetchTreatments`: bolus and temp basal fields, `timestamp` | Fixed: `timestamp` is `created_at`, `programmed` the insulin given, `temp` `absolute` |
| `fetchProfiles` | Works |
| `POST /api/v2/notifications/loop` (remote commands) | Fixed: pushed to the Loop that last uploaded its profile. `loop/remote_commands.http` |
