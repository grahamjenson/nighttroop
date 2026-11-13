# LoopFollow

## Overview

[LoopFollow](https://github.com/loopandlearn/LoopFollow) is an iOS app to follow a Loop or Trio user.

- It almost only reads: it polls the Nightscout v1 REST API, and listens on Socket.IO to poll sooner.
- Remote commands (bolus, carbs, overrides, temp targets) go to the Apple push service (APNS), not to Nightscout.
  Nightscout only gives where to send them, in the profile.
- Checked against LoopFollow's `main` in November 2026. Files are under `LoopFollow/`.
- Tests: `tests/api/nightscout/loopfollow/` (`connect.http`, `loop.http`, `trio.http`) and
  `tests/socket/nightscout/loopfollow_live.sio`. They upload as Loop and Trio do, then check the fields LoopFollow
  decodes.

## Set up

1. Add a device as Other (it works on an mmol/L site too). For remote commands, use a Read & Write token
   (`device_read_write`). Else, read-only is sufficient.
2. Give LoopFollow the token. Nighttroop has no API secret.
3. Connect. LoopFollow shows "OK (Read & Write)" or "OK (Read)".
4. Optional: turn on "Enable WebSocket" to poll sooner.

## Authentication

- Every request: `GET` with `?token=<token>`, no headers (`Helpers/NightscoutUtils.swift`, `constructURL`,
  `createURLRequest`).
- Connect: `GET /api/v1/status.json?token=` (`verifyURLAndToken`).
  - 200 connects. 401 is "token required" with no token, "invalid token" with one.
  - On a 200, it labels the connection from `authorized.token` and `authorized.permissionGroups`: an exact `*` is
    "Admin", an exact `api:treatments:create` "Read & Write", else "Read".
  - Nighttroop sends `authorized` with the identity's permissions, plus `api:treatments:create` if it can write.
- After a 401, LoopFollow checks if the token was the API secret: `GET /api/v2/authorization/subjects` with its
  SHA-1 in `api-secret` (`verifyAPISecret`). On a 200, it offers a read-only token (`provisionReadOnlyToken`,
  `POST /api/v2/authorization/subjects`).
  - Nighttroop takes the SHA-1 of a device's token (#512). A token that got a 401 is not one, so this is a 401.
- The profile's device token for remote commands goes only to an identity that can change the site's data.
- Socket.IO sends the token as the `token` connect param, and as `secret` in `authorize`.

## Units and dates

- LoopFollow reads `sgv` and `mbg` as mg/dL, as Nightscout stores them (`Extensions/ShareClientExtension.swift`,
  `ShareGlucoseData`). The Nightscout API always gives mg/dL. The profile too, with its `units`.
- Times sent: milliseconds; UTC with no zone, `yyyy-MM-dd'T'HH:mm:ss` (`Helpers/DateTime.swift`,
  `getDateTimeString`); ISO 8601, with or without fractional seconds.
- Entries, profiles, ages and database size use Swift `Codable`: one field of the wrong type fails the response.
- Treatments and device statuses use `JSONSerialization`: a wrong type drops only that field. The force casts below
  crash the app.

## Uploads

None. `NightscoutUtils.executePostRequest` (a POST with a JWT from `status.json`'s `authorized.token`, to
`/api/v2/notifications/loop` for overrides) has no callers. Remote commands go to APNS directly
(`Remote/TRC/PushNotificationManager.swift`, `Remote/LoopAPNS/LoopAPNSService.swift`), with the device token, bundle
and team from the profile.

## Downloads

### Entries

- When: every few minutes (`Controllers/Nightscout/BGData.swift`, `webLoadNSBGData`), and for stats
  (`Stats/StatsDataFetcher.swift`, `fetchBGData`).
- Request: `GET /api/v1/entries.json?count=<days × 4 × 288>&find[date][$gte]=<ms>&find[type][$ne]=cal`.
  - `days`: the days shown (1 by default, one more for yesterday's line).
  - Stats send `find[dateString][$gte]=<ISO 8601>` instead.
- Response: newest first, as `ShareGlucoseData`: `sgv` (a number, else `mbg`, else `glucose`, rounded to an Int),
  `date` (ms, required), `direction` (string, optional).

### Treatments

- When: with the BG (`Controllers/Nightscout/Treatments.swift`, `WebLoadNSTreatments`), for stats, and for the
  treatments list (`Treatments/TreatmentsView.swift`).
- Request: `GET /api/v1/treatments.json?find[created_at][$gte]=<a day ago>&find[created_at][$lte]=<6 hours ahead>
  &count=5000`, both UTC with no zone.
  - Stats send ISO times with fractional seconds.
  - The list pages back with `find[created_at][$lt]` and `count=100`.
- Ages: `find[eventType]=Site Change&find[created_at][$lte]=<now>&count=1` (`CAge.swift`). The same for
  `Sensor Start` (`SAge.swift`) and `Insulin Change` (`IAge.swift`), with `find[created_at][$gte]` 60 days ago.
  Each decodes `[{created_at: String}]`, required.
- Response: read with `JSONSerialization`, by `eventType`:

| `eventType` | Fields read |
| --- | --- |
| all | `created_at` or `timestamp`: strings, parsed by `NightscoutUtils.parseDate`, which drops the zone and reads them as UTC. `id` (Trio's) drops duplicates |
| `Temp Basal` | `absolute` (Double, required), `duration` (minutes) |
| `Correction Bolus`, `Bolus`, `External Insulin` | `insulin` (Double). `automatic` true makes it an SMB |
| `SMB` | `insulin` |
| `Meal Bolus` | `insulin`, `carbs` |
| `Carb Correction` | `carbs` (Double), `absorptionTime` (Int minutes) |
| `Temporary Override` (Loop), `Exercise` (Trio) | `duration` (minutes, 5 by default), `durationType` (`indefinite`), `correctionRange` as `[Int]` or `targetBottom`/`targetTop` as Int, `insulinNeedsScaleFactor`, `enteredBy`, name from `notes` or `reason` |
| `Temporary Target` | `targetBottom`/`targetTop` (Double), `duration`, `reason`, `enteredBy` (required) |
| `Note` | `notes` |
| `BG Check` | `glucose` (Double), `units` |
| `Suspend Pump`, `Resume Pump` | the time |
| `Site Change`/`Pump Site Change`, `Sensor Start`, `Insulin Change` | `created_at` |

**Caution:** Bolus, carbs and notes force cast `timestamp` to a String when it is there. It must be a string or
missing, never a number or null.

### Device status

- When: about every 5 minutes, slower with the socket connected (`Controllers/Nightscout/DeviceStatus.swift`). The
  treatments list also takes the one nearest a treatment.
- Request: `GET /api/v1/devicestatus.json?count=1`. The list sends `find[created_at][$gte]` and `[$lte]`, 30 minutes
  either side, with fractional seconds, and `count=30`.
- Response: the first record, read with `JSONSerialization`.

**Caution:** `pump`, `loop` and `openaps` are force cast to objects. Each must be an object or missing, never null.

| Source | Fields read |
| --- | --- |
| Pump | `pump.clock` (ISO, no fractional seconds) is the last loop time. The reservoir is read only with it. `pump.reservoir`, `pump.manufacturer`, `pump.model`, `pump.pumpID` give the reservoir (an Omnipod above 50U shows "50+"). `pump.battery.percent`: pump battery. `pump.bolusIncrement` or a `Dash` model: bolus step |
| Phone | `uploader.battery`, `uploader.isCharging` |
| Loop (`DeviceStatusLoop.swift`) | `loop.failureReason` (any value, even null, shows a failure), `loop.enacted`, `loop.iob.iob`, `loop.cob.cob`, `loop.predicted.values` (`[Double]`, force cast), `loop.recommendedBolus`, `loop.recommendedTempBasal.timestamp` (force cast to a String), `loop.timestamp`; top level `override.active`, `override.multiplier`, `override.currentCorrectionRange.minValue`/`maxValue` |
| Trio, OpenAPS (`DeviceStatusOpenAPS.swift`) | `device`, `openaps.failureReason`, `openaps.iob.iob`. From `openaps.suggested` (else `enacted`): `deliverAt` or `timestamp`, `ISF`, `CR`, `COB` (else from `reason`), `sensitivityRatio`, `eventualBG`, `current_target`, `TDD`, and `predBGs.ZT`, `IOB`, `COB`, `UAM` as `[Double]` |

### Profile

- When: every 10 minutes (30 with the socket), and when the socket says the profile changed (`Task/ProfileTask.swift`,
  `Controllers/Nightscout/Profile.swift`).
- Request: `GET /api/v1/profiles.json?count=1&find[startDate][$lte]=<a minute ahead, with fractional seconds>`.
  - The remote settings' diagnostics get `?count=1000` (`Remote/Settings/RemoteSettingsViewModel.swift`).
- Response: decoded as `NSProfile` (`Controllers/Nightscout/NSProfile.swift`).
  - Required: `store`, `defaultProfile`, `units`. In every store: `basal`, `sens` and `carbratio` (each `{value:
    Double, time: String, timeAsSeconds: Double}`), `timezone`, `units`.
  - The store must be called `Default` or `default`.
  - Loop APNS: `loopSettings.deviceToken`, `bundleIdentifier`, `overridePresets` (`name`, `duration` as Int,
    `targetRange` as `[Double]`, `insulinNeedsScaleFactor`, `symbol`).
  - Trio Remote Control: top level `deviceToken`, `bundleIdentifier`, `isAPNSProduction`, `teamID`, `expirationDate`
    (string), `overridePresets` (`name`, `duration`, `percentage`, `target` in mg/dL).

### Database size and Nightscout tab

- Database size: `GET /api/v2/properties/dbsize` (`Controllers/Nightscout/DBSize.swift`), decoded as
  `{dbsize?: {totalDataSize, dataPercentage: Int, details: {maxSize}}}`. Nighttroop has no `dbsize` and sends `{}`,
  so the row is hidden.
- Nightscout tab: a web view of `<site>/?token=` (`ViewControllers/NightScoutViewController.swift`). Nighttroop does
  not serve the Nightscout web app. It sends the tab on to the site's own pages, where the user signs in.

## Socket.IO

`Controllers/Nightscout/NightscoutSocketManager.swift`, when "Enable WebSocket" is on. Socket.IO-Client-Swift 16
(EIO=4), long polling, then websocket.

1. LoopFollow connects to the default namespace, `token` as a connect param.
2. It emits `authorize` `{client: "LoopFollow", history: 1, secret: <token>}`, no ack.
3. `connected`: authenticated. Device status polls three times slower, the profile every 30 minutes.
4. `dataUpdate` (`NightscoutSocketDataHandler.swift`): with no `delta` (the first), it fetches all. With
   `delta: true`, only its collections: `sgvs` or `mbgs`, `devicestatus`, `treatments`, `profiles`. Nothing
   else is read.

## Quirks

- LoopFollow reads any zone in `created_at` as UTC. Nighttroop always sends `created_at` in UTC.
- Nighttroop also keeps a BG Check as an `mbg` entry. LoopFollow draws it in the BG line.
- A treatment's `timestamp` is its `created_at`, a string, as LoopFollow needs.
- Nighttroop returns `enteredBy` and a bolus's `automatic` as sent (the identity's name if no `enteredBy` was sent).
  It drops a treatment's sent `timestamp` and `durationType`.
- A search returns at most the newest 1000 records, by design. A v1 profile search with no count returns the newest
  10 stores, as Nightscout does.

## Status on Nighttroop

| Interaction | Nighttroop |
| --- | --- |
| **Connect** | |
| `status.json` to connect, 200 and 401 | Works |
| `status.json` `authorized`, for "Read & Write" | Fixed |
| API secret check, token provisioning (`/api/v2/authorization/subjects`) | Not supported (no API secret): a 401 |
| **Entries and treatments** | |
| Entries, `find[date]`, `find[dateString]`, `find[type][$ne]=cal` | Works with the identity in mg/dL |
| `sgv` in mg/dL for an identity in mmol/L | Not supported, TODO P1 |
| More than 1000 entries or treatments | Not supported: at most the newest 1000, by design |
| Treatments with zoneless `find[created_at]` times | Works since #516 (were a 400) |
| `find[created_at]` on when a treatment or device status happened | Works since #514 (was when it was stored) |
| Loop `Temporary Override`, Trio `SMB`, `Suspend Pump`, `Resume Pump` | Works since #515 and #516 (a batch with one was a 400) |
| Override `durationType: indefinite` | Fixed: kept with `duration` 10080 since #515. Given back with `durationType` `indefinite` while the override as sent is kept |
| Bolus `automatic` | Fixed: kept. An automatic bolus is an SMB, and not in the bolus stats |
| Treatment `enteredBy` | Fixed (was the identity's name) |
| Site, sensor and insulin ages | Works (with the zoneless fix) |
| **Device status** | |
| Loop and OpenAPS objects, IOB, COB, predictions | Works |
| `pump.clock`, `manufacturer`, `model` | Works since #516 (were dropped) |
| `pump.pumpID` | Works since #514, while the status as sent is kept (as the latest is) |
| `pump.bolusIncrement` | Fixed, while the status as sent is kept (as the latest is) |
| `uploader.battery` | Works since #516 (Loop's was `uploaderBattery` only, Trio's was dropped) |
| `uploader.isCharging` | Fixed (was dropped) |
| Loop `override` | Works since #516 (was dropped) |
| **Profile** | |
| Profile, schedules, Loop's `loopSettings` | Works |
| Trio remote control settings and override presets | Works since #516 (were dropped) |
| `deviceToken`s for a read-only token | Left out, as they push to the phone. Remote commands need a Read & Write token |
| Profile history for the remote diagnostics | Fixed: `count` gives that many stores, newest first (was only the latest) |
| **Other** | |
| `dbsize` property | Not supported (sends `{}`, the row is hidden) |
| Nightscout tab (`/`) | Fixed: sent on to the site's pages |
| Socket.IO connect, `authorize` without ack, `connected`, `dataUpdate` and deltas | Works |
| xDrip+ and Juggluco boluses and carbs (sent as `<none>`) | Fixed: returned as Meal Bolus, Carb Correction or Correction Bolus. Before, LoopFollow showed a Note dot only |
