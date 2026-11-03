# LoopFollow

[LoopFollow](https://github.com/loopandlearn/LoopFollow) is an iOS app for following someone who uses Loop or
Trio. It is almost only a reader: it polls Nightscout's v1 REST API, and listens on Socket.IO to know when to poll
sooner. Its remote commands (bolus, carbs, overrides, temp targets) go to Apple's push service (APNS), not to
Nightscout; Nightscout only gives it where to send them, in the profile.

This was checked against LoopFollow's `main` in November 2026. Files below are under `LoopFollow/`.

The tests are `tests/api/nightscout/loopfollow/` (`connect.http`, `loop.http`, `trio.http`) and
`tests/socket/nightscout/loopfollow_live.sio`. They upload what Loop and Trio upload, then make LoopFollow's
requests and check the fields it decodes.

## Authentication

- Every request is `GET` with `?token=<token>` in the query (`Helpers/NightscoutUtils.swift`, `constructURL` and
  `createURLRequest`). No `Authorization` or `api-secret` header is sent on reads, and no other header is set.
- Connecting gets `/api/v1/status.json?token=` (`verifyURLAndToken`). It reads only the status code: 200 connects,
  401 is "token required" with no token or "invalid token" with one. On a 200 it also reads `authorized.token` and
  `authorized.permissionGroups` to label the connection "Read", "Read & Write" or "Admin": an exact `*` is Admin
  and an exact `api:treatments:create` is Read & Write. Nighttroop sends `authorized` with the identity's
  permissions, adding `api:treatments:create` for an identity that can write, so a write token shows "OK (Read &
  Write)" and a read token "OK (Read)".
- After a 401, LoopFollow checks whether the token was really the API secret, with
  `GET /api/v2/authorization/subjects` and the SHA-1 of it in `api-secret` (`verifyAPISecret`), and offers to make a
  read-only token from it on a 200 (`provisionReadOnlyToken`, `POST /api/v2/authorization/subjects`). Nighttroop
  has no API secret: it takes the SHA-1 of a device's token (#512), which a token that got a 401 is not, so this is a
  401 and LoopFollow does not offer it; give it a Nighttroop token.
- A read-only token is enough for everything LoopFollow does.
- Socket.IO sends the token as the `token` connect param and as `secret` in `authorize`.

## Glucose units

LoopFollow reads `sgv` and `mbg` as mg/dL whatever the units, as Nightscout always stores them
(`Extensions/ShareClientExtension.swift`, `ShareGlucoseData`). Nighttroop sends them in the identity's units, so
**LoopFollow's identity must be set to mg/dL**; in mmol/L a 6.7 reading is shown as 7 mg/dL. LoopFollow is not one of
Add Device's clients, so a device added for it as Other has the site's units; a site admin can set its own units to
mg/dL from its row on Users & Devices (#523). The profile is fine either way, as LoopFollow reads its `units`.

## Calls

Times LoopFollow sends are in three formats: milliseconds; UTC with no zone, `yyyy-MM-dd'T'HH:mm:ss`
(`Helpers/DateTime.swift`, `getDateTimeString`); and ISO 8601 with or without fractional seconds. Entries,
profiles, ages and the database size are decoded with Swift `Codable`, so one field of the wrong type fails the whole
response. Treatments and device statuses are read with `JSONSerialization`, where a wrong type only drops that
field, except for the force casts below, which crash the app.

### Entries

- When: every few minutes (`Controllers/Nightscout/BGData.swift`, `webLoadNSBGData`), and for stats
  (`Stats/StatsDataFetcher.swift`, `fetchBGData`).
- Request: `GET /api/v1/entries.json?count=<days × 4 × 288>&find[date][$gte]=<ms>&find[type][$ne]=cal`, with
  `days` the days shown (1 by default, one more for yesterday's line). Stats send
  `find[dateString][$gte]=<ISO 8601>` instead.
- Response: an array decoded as `ShareGlucoseData`: `sgv` (a number, else `mbg`, else `glucose`, rounded to an Int),
  `date` (ms, required) and `direction` (string, optional). Newest first.

### Treatments

- When: with the BG (`Controllers/Nightscout/Treatments.swift`, `WebLoadNSTreatments`), for stats, and for the
  treatments list (`Treatments/TreatmentsView.swift`).
- Request: `GET /api/v1/treatments.json?find[created_at][$gte]=<a day ago>&find[created_at][$lte]=<6 hours ahead>
  &count=5000`, both UTC with no zone. Stats send ISO times with fractional seconds; the list pages back with
  `find[created_at][$lt]` and `count=100`.
- Ages: `find[eventType]=Site Change&find[created_at][$lte]=<now>&count=1` (`CAge.swift`), and the same for
  `Sensor Start` (`SAge.swift`) and `Insulin Change` (`IAge.swift`) with `find[created_at][$gte]` 60 days ago. Each
  decodes `[{created_at: String}]`, required.
- Response: read with `JSONSerialization`, by `eventType`:
  - every one: `created_at` or `timestamp` (strings, parsed by `NightscoutUtils.parseDate`, which drops the zone and
    reads them as UTC). Bolus, carbs and notes force cast `timestamp` to a String when it is there, so it must be a
    string or missing, never a number or null. `id` (Trio's) is used to drop duplicates.
  - `Temp Basal`: `absolute` (Double, required) and `duration` (minutes).
  - `Correction Bolus`, `Bolus`, `External Insulin`: `insulin` (Double); `automatic` true makes it an SMB. `SMB`:
    `insulin`. `Meal Bolus`: `insulin` and `carbs`.
  - `Carb Correction`: `carbs` (Double) and `absorptionTime` (Int minutes).
  - `Temporary Override` (Loop) and `Exercise` (Trio): `duration` (minutes, 5 by default), `durationType`
    (`indefinite`), `correctionRange` as `[Int]` or `targetBottom`/`targetTop` as Int, `insulinNeedsScaleFactor`,
    `enteredBy`, and the name from `notes` or `reason`.
  - `Temporary Target`: `targetBottom`/`targetTop` (Double), `duration`, `reason` and `enteredBy` (required).
  - `Note`: `notes`. `BG Check`: `glucose` (Double) and `units`. `Suspend Pump` and `Resume Pump`: the time.
  - `Site Change`/`Pump Site Change`, `Sensor Start` and `Insulin Change`: `created_at`.

### Device status

- When: about every 5 minutes, slower when the socket is connected (`Controllers/Nightscout/DeviceStatus.swift`).
  The treatments list also takes the one nearest a treatment.
- Request: `GET /api/v1/devicestatus.json?count=1`; the list sends `find[created_at][$gte]` and `[$lte]` 30 minutes
  either side, with fractional seconds, and `count=30`.
- Response: the first record, read with `JSONSerialization`. `pump`, `loop` and `openaps` are force cast to objects,
  so each must be an object or missing, never null.
  - `pump.clock` (ISO, no fractional seconds) is the last loop time; the reservoir is only read with it.
    `pump.reservoir`, `pump.manufacturer`, `pump.model` and `pump.pumpID` give the reservoir (an Omnipod above 50U
    shows "50+"), `pump.battery.percent` the pump battery, and `pump.bolusIncrement` or a `Dash` model the bolus
    step.
  - `uploader.battery` and `uploader.isCharging`: the phone's battery.
  - Loop (`DeviceStatusLoop.swift`): `loop.failureReason` (any value, even null, shows a failure), `loop.enacted`,
    `loop.iob.iob`, `loop.cob.cob`, `loop.predicted.values` (`[Double]`, force cast), `loop.recommendedBolus`,
    `loop.recommendedTempBasal.timestamp` (force cast to a String), `loop.timestamp`, and the top level
    `override.active`, `override.multiplier` and `override.currentCorrectionRange.minValue`/`maxValue`.
  - Trio and OpenAPS (`DeviceStatusOpenAPS.swift`): `device`, `openaps.failureReason`, `openaps.iob.iob`, and from
    `openaps.suggested` (else `enacted`): `deliverAt` or `timestamp`, `ISF`, `CR`, `COB` (else from `reason`),
    `sensitivityRatio`, `eventualBG`, `current_target`, `TDD`, and `predBGs.ZT`, `IOB`, `COB` and `UAM` as
    `[Double]`.

### Profile

- When: every 10 minutes (30 with the socket), and when the socket says the profile changed
  (`Task/ProfileTask.swift`, `Controllers/Nightscout/Profile.swift`).
- Request: `GET /api/v1/profiles.json?count=1&find[startDate][$lte]=<a minute ahead, with fractional seconds>`.
  The remote settings' diagnostics get `?count=1000` (`Remote/Settings/RemoteSettingsViewModel.swift`).
- Response: decoded as `NSProfile` (`Controllers/Nightscout/NSProfile.swift`), so these are required: `store`,
  `defaultProfile`, `units`, and in every store `basal`, `sens` and `carbratio` (each `{value: Double, time: String,
  timeAsSeconds: Double}`), `timezone` and `units`. Its store must be called `Default` or `default`.
  - Loop APNS: `loopSettings.deviceToken`, `bundleIdentifier`, and `overridePresets` (`name`, `duration` as Int,
    `targetRange` as `[Double]`, `insulinNeedsScaleFactor`, `symbol`).
  - Trio Remote Control: the top level `deviceToken`, `bundleIdentifier`, `isAPNSProduction`, `teamID`,
    `expirationDate` (string) and `overridePresets` (`name`, `duration`, `percentage`, `target` in mg/dL).

### Database size

`GET /api/v2/properties/dbsize` (`Controllers/Nightscout/DBSize.swift`), decoded as `{dbsize?: {totalDataSize,
dataPercentage: Int, details: {maxSize}}}`. Nighttroop has no `dbsize` and sends `{}`, so the row is hidden.

### Nightscout tab

The Nightscout tab is a web view of `<site>/?token=` (`ViewControllers/NightScoutViewController.swift`). Nighttroop
does not serve the Nightscout web app, so it sends the tab on to the site's own pages, where the user signs in.

## Socket

`Controllers/Nightscout/NightscoutSocketManager.swift`, when "Enable WebSocket" is on, with Socket.IO-Client-Swift
16 (EIO=4), long polling then websocket.

- Connects to the default namespace with `token` as a connect param.
- Emits `authorize` with `{client: "LoopFollow", history: 1, secret: <token>}` and no ack.
- `connected` marks it authenticated, which polls the device status three times slower and the profile every 30
  minutes.
- `dataUpdate` (`NightscoutSocketDataHandler.swift`): with no `delta` (the first one) it fetches everything; with
  `delta: true` it fetches only the collections in it: `sgvs` or `mbgs`, `devicestatus`, `treatments` and
  `profiles`. It reads nothing else from it.

## Writes

LoopFollow writes nothing to Nightscout. `NightscoutUtils.executePostRequest` (a POST with a JWT from
`status.json`'s `authorized.token`, to `/api/v2/notifications/loop` for overrides) has no callers. Remote commands
go to APNS directly (`Remote/TRC/PushNotificationManager.swift`, `Remote/LoopAPNS/LoopAPNSService.swift`), with the
device token, bundle and team from the profile.

## Quirks

- LoopFollow treats a zone in `created_at` as UTC, so times must be sent in UTC; Nighttroop always sends
  `created_at` in UTC.
- A BG Check is also kept as a meter reading, so it is an `mbg` entry too, and LoopFollow draws it in the BG line.
- A treatment's `timestamp` is its `created_at`, a string as LoopFollow needs.
- Nighttroop returns the identity's name as `enteredBy`, and drops a treatment's sent `timestamp`, `automatic` and
  `durationType`.
- A search returns at most 1000 records, the newest, by design, and a profile search with no filters only the latest
  store.

## Status

| Interaction | Nighttroop |
| --- | --- |
| `status.json` to connect, 200 and 401 | Works |
| `status.json` `authorized`, for "Read & Write" | Fixed in this PR |
| API secret check and token provisioning (`/api/v2/authorization/subjects`) | Not supported (no API secret); a 401 |
| Entries, `find[date]`, `find[dateString]`, `find[type][$ne]=cal` | Works with the identity in mg/dL |
| `sgv` in mg/dL for an identity in mmol/L | Not supported, TODO P1 |
| More than 1000 entries or treatments | Not supported: a search gives at most the newest 1000, by design |
| Treatments with zoneless `find[created_at]` times | Works since #516 (they were a 400) |
| `find[created_at]` on when a treatment or device status happened | Works since #514 (it was when it was stored) |
| Loop `Temporary Override`, Trio `SMB`, `Suspend Pump`, `Resume Pump` | Works since #515 and #516 (a batch with one was a 400) |
| Override `durationType: indefinite` | Shown as a week long: kept with `duration` 10080 since #515, without `durationType`, TODO P4 |
| Bolus `automatic` | Not supported, TODO P2 |
| Treatment `enteredBy` | Not supported (the identity's name), TODO P4 |
| Site, sensor and insulin ages | Works (with the zoneless fix) |
| Device status Loop and OpenAPS objects, IOB, COB, predictions | Works |
| Device status `pump.clock`, `manufacturer`, `model` | Works since #516 (they were dropped) |
| Device status `pump.pumpID` | Works since #514 while the status as sent is kept, as the latest is |
| Device status `pump.bolusIncrement` | Fixed in this PR, while the status as sent is kept, as the latest is |
| Device status `uploader.battery` | Works since #516 (Loop's was `uploaderBattery` only, Trio's was dropped) |
| Device status `uploader.isCharging` | Fixed in this PR (it was dropped) |
| Device status Loop `override` | Works since #516 (it was dropped) |
| Profile, schedules and Loop's `loopSettings` | Works |
| Profile Trio remote control settings and override presets | Works since #516 (they were dropped) |
| Profile history for the remote diagnostics | Not supported (only the latest), TODO P4 |
| `dbsize` property | Not supported (sends `{}`, the row is hidden) |
| Nightscout tab (`/`) | Fixed: sent on to the site's pages |
| Socket.IO connect, `authorize` without ack, `connected`, `dataUpdate` and deltas | Works |
