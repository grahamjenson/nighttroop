# AAPS (AndroidAPS) and Nighttroop

## Overview

How AAPS uses the Nightscout API, and the Nighttroop status of each interaction.

- Source: [nightscout/AndroidAPS](https://github.com/nightscout/AndroidAPS). NSClient v3: tag `4.0.0` (commit
  `f77abd9`, the same as `master` on 2026-11-02). NSClient v1: tag `3.4.2.6` (AAPS 4.0 removed v1).
- Paths start at the repository root. `core/nssdk/.../` is `core/nssdk/src/commonMain/kotlin/app/aaps/core/nssdk/`.
  `nsclientV3/` is `plugins/sync/src/commonMain/kotlin/app/aaps/plugins/sync/nsclientV3/`.
- Tests: `tests/api/nightscout/aaps/*.http` and `tests/socket/nightscout/aaps_*.sio`.

## Set up

1. For a master (the phone with the pump), add a device for AAPS. It is always the PWD Phone (#527), and can write.
2. For a follower (AAPSClient), add a device as Other. A read only token is sufficient.
3. In AAPS, set the URL to the site's Nightscout host.
   - AAPS adds `https://` if there is no scheme, and `/api/` after it (`core/nssdk/.../networking/NsUrl.kt:21`).
4. Set the token to the device's token. In AAPS 3.4 (NSClient v1), put it in the API secret field.

## Glucose units

- AAPS reads `sgv` as mg/dL and ignores `units` (`nsclientV3/extensions/GlucoseValueExtension.kt` `NSSgvV3.toGV`).
- AAPS knows only `mg/dl` and `mmol` (`core/nssdk/.../localmodel/entry/NsUnits.kt`, `GlucoseUnit.fromText` in
  `core/data`). Other units, such as Nighttroop's `mmol/L`, are mg/dL to it.
- The Nightscout API always gives glucose, temp targets, BG checks and profiles in mg/dL, as Nightscout does,
  whatever the units of the identity or site.
- Uploads work in all units: AAPS labels them `mg/dl`.

## NSClient v3 (AAPS 4.0, and 3.x with "Use API v3")

### Authentication

One interceptor handles every request (`core/nssdk/.../networking/NsAuth.kt`):

1. AAPS sends `Date: <now in milliseconds>` (Nighttroop ignores it) and `Authorization: Bearer <jwt>` (lines
   46–47). The JWT is empty until AAPS has one.
2. On a **401 or 403** (line 127), AAPS checks the body for the Nightscout date tolerance message (line 55).
3. AAPS gets a JWT from `GET /api/v2/authorization/request/<token>`: host only, no sub path, no `Authorization`
   header (line 98).
4. AAPS sends the request again, one time, with the new JWT (lines 63–65).

- The refresh response needs `token`, `iat` and `exp` (`core/nssdk/.../remotemodel/RemoteAuthResponse.kt`).
- A 401 or 403 from the refresh shows "invalid access token", with no retry.
- The socket `subscribe` sends the token, not the JWT.

### Request and response format

- Endpoints: `core/nssdk/.../networking/NightscoutApi.kt`. Response handling: `core/nssdk/.../NSAndroidClientImpl.kt`.
- Query keys keep their `$` (`date$gt`, not `date%24gt`).
- A v3 read is `{"result": …}`. A create, patch and delete response is `{identifier, isDeduplication,
  deduplicatedIdentifier, lastModified}` at the top level, not only in `result` (`remotemodel/RemoteStatusResponse.kt`,
  `RemoteCreateUpdateResponse`).
- Lenient JSON (`core/nssdk/.../NsSdkJson.kt`): unknown keys ignored, a quoted number read as a number, an unknown
  `eventType` skipped.
- **Caution:** A fraction in an `Int` or `Long` field fails the full page. AAPS then asks for that page again,
  without end.

### Requests

| When | Request | AAPS reads or needs |
| --- | --- | --- |
| Each sync, first | `GET /api/v3/status` | `version`, `apiVersion`, `srvDate`, `storage.{storage,version}`, `apiPermissions` (`RemoteStatusResponse`): six collections, each a string of `c`, `r`, `u`, `d`. A master uploads only if all are `crud`. A client needs read on all, and create and update on `settings` (`localmodel/ApiPermissions.kt`, `NSClientV3Plugin.kt:534`) |
| Status not full | `GET /api/v2/authorization/debug/check/api:settings:create` and `…:update` | 200 yes, 401 or 403 no. Other statuses fail the status (`NSAndroidClientImpl.kt:119`) |
| Each sync | `GET /api/v3/lastModified` | `collections.{devicestatus,entries,profile,treatments,foods,settings}` as numbers (`remotemodel/LastModified.kt`). `foods`, not Nightscout's `food`, so always 0 |
| Entries, first load | `GET /api/v3/entries?sort=date&date$gt=<ms>&limit=500`: from 100 days ago (`NSClientV3Plugin.kt:185,232`), then from the newest `date`. No other filter, so pages of 500, the oldest first | `type` (must be `sgv`), `sgv`, `date`, `identifier`, `direction`, `utcOffset`, `isValid`, `srvModified` (`mapper/SvgMapper.kt`) |
| Entries, later loads | `GET /api/v3/entries/history/<ms>?limit=500`, from the last `ETag` | The same. `isValid: false` is a deletion. The `ETag` `W/"<ms>"`, read by position (`NsHttpResponse.kt:58`), is the next `<ms>`. Stops when the ETag does not change (`workers/LoadBgRunner.kt`) |
| A reading or calibration | `POST /api/v3/entries` | 200 or 201, and `identifier`. A 400 "Bad or missing utcOffset field" is sent again with `utcOffset: 0` (`NSAndroidClientImpl.kt:226`) |
| Reading changed / invalidated | `PATCH /api/v3/entries/<id>` without `date` and `utcOffset` / `DELETE /api/v3/entries/<id>` | 2xx, or 404 as done (line 259). Other statuses stop the upload queue until it works |
| Treatments, first load | `GET /api/v3/treatments?sort=created_at&created_at$gt=<ISO>&limit=500` | See [Bodies AAPS sends](#bodies-aaps-sends) |
| Treatments, later loads | `GET /api/v3/treatments/history/<ms>?limit=500` | As entries, with the `ETag` |
| A treatment | `POST /api/v3/treatments` | As entries. A 400 is logged and **skipped permanently**: all that Nighttroop refuses is lost. A status other than 200, 201, 400 or 404 stops the upload, to try again (`NSClientV3Plugin.kt:1079–1091`) |
| Treatment changed / invalidated | `PATCH` / `DELETE /api/v3/treatments/<id>` | As entries |
| Each loop | `POST /api/v3/devicestatus` | 200 or 201 |
| Master: the newest device status | `GET /api/v3/devicestatus/history/<now-7min>?limit=500` | `srvModified`, `date`, `device`, `uploaderBattery` (Int), `isCharging`, `pump.{clock,reservoir,battery.percent (Int),status.status,extended}`, `openaps.{suggested,enacted,iob}` (`remotemodel/RemoteDeviceStatus.kt`) |
| Client: all of the last 48 hours | The same, paged from the newest `srvModified` (`workers/LoadDeviceStatusRunner.kt`) | Pump data skipped without `pump.clock` (`nsclientV3/data/NSDeviceStatusHandler.kt:225`) |
| Foods | `GET /api/v3/food?limit=1000`, `POST /api/v3/food`. Never update or delete (`NightscoutApi.kt:216`) | Only `type` `food`; `carbs`, `gi`, `energy`, `protein`, `fat` as Int (`remotemodel/RemoteFood.kt`, `mapper/FoodMapper.kt`) |
| Profiles, first load | `GET /api/v3/profile?sort$desc=date&limit=1` | The store as JSON: `defaultProfile`, `startDate`, `store`, `srvModified` (`workers/LoadProfileStoreRunner.kt`) |
| Profiles, later loads | `GET /api/v3/profile/history/<ms>?limit=10` | **The last** in the list, as the newest |
| Profile changed | `POST /api/v3/profile` with `app: AAPS` added | 200 or 201 |
| Client control (4.0) | `GET/POST/PATCH/PUT/DELETE /api/v3/settings[/<id>]` for `aaps`, `aaps-state`, `aaps_clientcontrol_*`; `/api/v2/authorization/roles` and `/subjects` with an admin token | Client pairing and remote commands |

### Bodies AAPS sends

From `core/nssdk/.../mapper/TreatmentMapper.kt` (`toRemoteTreatment`, line 397) and
`nsclientV3/extensions/*Extension.kt`.

- Nulls left out; `false` and `0` sent. A create sets `app: AAPS`. An update sends `identifier`.
- Every treatment: `date` (ms), `utcOffset` (minutes), `isValid`, `isReadOnly`, `eventType`, and, if known, the
  pump's `pumpId`, `pumpType`, `pumpSerial`, `endId`.

| Record | `eventType` and fields |
| --- | --- |
| **Bolus** | `Meal Bolus`, `Correction Bolus`: `insulin`, `type` (`NORMAL`, `SMB`, `PRIMING`), `isBasalInsulin` |
| **Carbs** | `Carb Correction`, `Meal Bolus`: `carbs`, `duration` (min), `durationInMilliseconds` |
| **Temp Basal** | `type` (`NORMAL`, `EMULATED_PUMP_SUSPEND`, `PUMP_SUSPEND`, `SUPERBOLUS`, `FAKE_EXTENDED`). See [Temp basal percent](#temp-basal-percent) |
| **Temporary Target** | `targetTop`, `targetBottom` in mg/dL with `units: "mg/dl"`, `reason`, `duration`, `durationInMilliseconds` |
| **Profile Switch** | `profile`, `profileJson` (a JSON string), `originalProfileName`, `originalDuration`, `duration`, `durationInMilliseconds`, `timeshift` (ms), `percentage` (Int) |
| **Effective profile switch** | A `Note` with `profileJson`, `originalProfileName`, `originalCustomizedName`, `originalTimeshift`, `originalPercentage`, `originalDuration`, `originalEnd`. Read back as one only with all of them (`TreatmentMapper.kt:198`) |
| **Running mode** | `OpenAPS Offline`: `duration`, `durationInMilliseconds`, `reason`, `mode`, `autoForced`, `reasons`, `originalDuration` |
| **Therapy events** | `Site Change`, `Sensor Change`, `BG Check`, `Note`, `Exercise`, `Announcement`, …: `duration`, `durationInMilliseconds`, `glucose`, `glucoseType`, `units`, `enteredBy`, `location`, `arrow`, `notes` |
| **Bolus Wizard** | `bolusCalculatorResult` (a JSON string), `glucose` |
| **Extended bolus** | `Combo Bolus`: `enteredinsulin`, `relative` (U/hr), `splitNow: 0`, `splitExt: 100`, `isEmulatingTempBasal`. Or a `Temp Basal` with it in `extendedEmulated` |
| **Entries** | `type: "sgv"`, `sgv` (mg/dL), `date`, `device`, `direction`, `noise`, `filtered`, `unfiltered`, `units: "mg/dl"`. Calibration: `type: "mbg"`, `mbg`, `sensorMgdlAtPairing`, `isCalibration: true` |
| **Device status** | `date`, `device` (`openaps://<phone>`), `uploaderBattery`, `isCharging`, `uploader: {}`, `pump.{clock,reservoir,battery,status,extended}`, `openaps.{suggested,enacted,iob}`, `configuration` |

### Temp basal percent

`percent` is the **change** from the basal, in AAPS, Nightscout and Nighttroop.

- AAPS sends (`TB.toNSTemporaryBasal`, `nsclientV3/extensions/TemporaryBasalExtension.kt:27–44`):
  - `rate`: always absolute, U/hr (`convertedToAbsolute`, line 34).
  - Set in U/hr: also `absolute` (line 36).
  - Set as a percent of the basal (Dana, Combo, Insight pumps): `percent = rate - 100`, no `absolute` (line 37).
    50% is `percent: -50`, 120% is `percent: 20`, a zero temp is `percent: -100`.
  - No `isAbsolute` (not a `RemoteTreatment` field). Comments: `NSTemporaryBasal.kt:32` "when sending to NS (rate -
    100)", `RemoteTreatment.kt:61` "Eventual basal change in percent".
- AAPS reads (`RemoteTreatment.toTreatment`, `TreatmentMapper.kt:167–196`): `isAbsolute = absolute != null`,
  `rate = absolute ?: percent + 100`. `absolute` has priority; a percent alone is a change.
- AAPS 3.4 v1 is the same: `TB.toJson`
  (`plugins/sync/src/main/kotlin/app/aaps/plugins/sync/nsclient/extensions/TemporaryBasalExtension.kt:23–26`) sends
  `rate` (absolute) and `absolute`, or `percent: rate - 100`. `temporaryBasalFromJson` (lines 54–62) reads
  `absolute`, else `percent + 100`.
- Nighttroop:
  - Keeps `rate` (U/hr), and a `multiplier` of `1 + percent / 100` (0.5 for -50) for every percent from -100 up. It
    returns the same percent.
  - Keeps the rate AAPS sent, which followers read first, as `absolute`.
  - No rate: profile basal at the start × multiplier, as in Nightscout. No percent: calculated the same way.
  - Returns `rate`, `absolute`, and `percent` if it has one.

### Socket.IO

socket.io-client-java 2.1.2 (`EIO=4`), on two namespaces of the site URL
(`nsclientV3/ws/SocketNsConnection.kt:136–170`), when "Use websockets" is on.

- **`/storage`**:
  1. On connect, AAPS emits `subscribe` `{"accessToken": <token>, "collections": ["devicestatus", "entries",
     "profile", "treatments", "foods", "settings"]}` and reads `success` from the ack
     (`nsclientV3/ws/NsConnectHandler.kt:53–73`).
  2. AAPS listens for `create` and `update` (`colName`, `doc`, `doc.srvModified`; a frame without it is dropped) and
     `delete` (`colName`, `identifier`) (`nsclientV3/ws/NsFrameHandler.kt:88–223`).
  3. AAPS acts on treatments, entries, device statuses, profiles, foods (as `foods`) and settings.
- **`/alarm`**, if alarm or announcement notifications are on:
  1. AAPS emits `subscribe` `{"accessToken": <token>}`.
  2. AAPS listens for `announcement`, `alarm`, `urgent_alarm`, `clear_alarm`: `level`, `group`, `title`, `message`,
     `eventName` (`nsclientV3/NSAlarmObject.kt`).
  3. A snoozed alarm is acked with `emit("ack", level, group, silenceMillis)` (`jvmSharedMain/.../ws/SocketIoNsSocket.kt:51`).

### Quirks

- `pumpId` comes back as a string. AAPS reads it into its `Long` all the same.
- AAPS ignores `programmed`, which Nighttroop adds to a bolus for Loop (`ignoreUnknownKeys`, `core/nssdk/.../NsSdkJson.kt`).
- `created_at$gt` is the treatment time, as in Nightscout (#514). `srvCreated$gt` is when it arrived.

## NSClient v1 (AAPS 3.4 and older)

Socket.IO only, no REST: socket.io-client-java 2.1.2 on the default namespace, only over `https`
(`plugins/sync/src/main/kotlin/app/aaps/plugins/sync/nsclient/services/NSClientService.kt:277–300`).

1. On connect, AAPS emits `authorize` `{"client": "Android_AAPS", "history": 48, "status": true, "from": <newest ms
   it has>, "secret": <SHA1 of the API secret>}` (lines 361–375).
   - Nighttroop finds a token by its SHA1, as for `api-secret` (#512), so a device token works. `aaps_v1.sio` sends
     the token's SHA1.
2. AAPS reads `read`, `write`, `write_treatment` from the ack (`nsclient/acks/NSAuthAck.kt`), and writes only if
   permitted.
3. AAPS listens (lines 286–290, 466–580) for `announcement`, `alarm`, `urgent_alarm`, `clear_alarm`, and `dataUpdate`:
   `status`, `profiles` (the last is the newest), `treatments` (`action` `update` or `remove`), `devicestatus`,
   `food`, `mbgs`, `cals`, `sgvs`, `delta`.
4. AAPS writes with `dbAdd` `{collection, data}`, reading `_id` from the ack array's first item
   (`nsclient/acks/NSAddAck.kt`), and `dbUpdate` `{collection, _id, data}`, reading `result: "success"`
   (`nsclient/acks/NSUpdateAck.kt`).
5. AAPS acks alarms with `ack`, as in v3 (line 630).

## Status on Nighttroop

Unless a row says otherwise, Nighttroop keeps each field and returns it as sent.

| Interaction | Status | Test |
| --- | --- | --- |
| **Authentication** | | |
| JWT from `/api/v2/authorization/request/<token>`; 401 without one or for a bad token | Works | `auth_v3.http` |
| `/api/v3/status` fields | Works | `auth_v3.http` |
| `/api/v3/status` permissions of a read only token | Fixed in #511 (were always `crud`, so AAPS tried to upload) | `auth_v3.http` `read_status` |
| `/api/v2/authorization/debug/check/<permission>` | Fixed in #511 (was a 404, which fails the status once it shows real permissions) | `auth_v3.http` |
| `/api/v3/lastModified` | Works | `auth_v3.http`, `sync_v3.http` |
| **Entries and paging** | | |
| Create, duplicate, search by date, history with `ETag`, patch, delete, deletion in history | Works | `sync_v3.http` |
| `filtered` and `unfiltered` of 0, sent with every reading | Not kept (0 is no raw signal) and left out. AAPS does not read them | `flow_v3.http` |
| Calibrations (`isCalibration`, `sensorMgdlAtPairing`) | Works: kept with the mbg. A follower reads a calibration as one | `sync_v3.http` |
| Entries and treatments of an identity in mmol/L | Works: always mg/dL | `units/*.http` |
| First loads over 500 documents | Works: sorted by date, the oldest first | `first_load_paging.http` |
| History over 500 documents, and profile history's newest | Works: the oldest `srvModified` first | `history_paging.http` |
| **Treatments** | | |
| Bolus and SMB `type`, carbs, temp target, profile switch, BG check, delete in history | Works | `sync_v3.http`, `flow_v3.http` |
| A treatment with both insulin and carbs, such as a Meal Bolus from xDrip+ or the careportal (`RemoteTreatment.toTreatment` makes one record, insulin first) | Fixed: API v3 lists and the storage socket give it as two, its insulin and its carbs (identifier plus `carbs`). AAPS dropped the carbs, so COB was wrong. A change or delete of one part keeps the other | `meal_bolus_split.http`, `aaps_storage_meal_bolus.sio` |
| Bolus Wizard bolus and carbs: two Meal Boluses at the same `date` | Fixed: two treatments, each with its own identifier. The bolus keeps its insulin (Nightscout replaces it with the carbs). A delete of the carbs deletes only them (was merged, so it deleted the bolus) | `wizard_dedup_v3.http` |
| Bolus `icfg` and `isBasalInsulin` (without them, AAPS uses its settings' insulin and a bolus is not basal: `BolusExtension.kt` `NSBolus.toBolus`) | Works. Nighttroop summaries still count a basal insulin bolus as a bolus (TODO P3) | `sync_v3.http` |
| Profile switch and effective profile switch `profileJson` `iCfg` (AAPSClient reads the insulin from it: `pureProfileFromJson`) | Fixed: returned from the switch's `icfg`, or kept from the `profileJson` when there is no `icfg` (was left out) | `profile_switch_icfg.http` |
| Bolus Wizard `bolusCalculatorResult` (AAPS restores the record from it, or drops the record: `TreatmentMapper.kt:270`) | Works: returned exactly as sent. Left out if not JSON or over 4096 characters | `bolus_wizard.http` |
| Bolus Wizard with no BG (`glucose` 0) | Fixed: kept without a glucose reading (was 400, and AAPS does not send it again) | `bolus_wizard.http` |
| Profile Switch from the site or the careportal (`timeshift` in minutes) | Fixed: API v3 and the storage socket give it in ms, as AAPS reads it (a 60 minute shift was read as 60 ms) | `profile_switch_timeshift.http` |
| `units` as AAPS writes them | Works: a device added for AAPS writes `mg/dl`. Another device writes units as it first sent them, `mg/dl` once AAPS uploads | `bolus_wizard.http` |
| Temp basal at, over or under 100% (any `percent` from -100), and its update | Works: kept as a multiplier | `sync_v3.http`, `treatments/temp_basal_percent.http` |
| Temp basal `type` (for example `PUMP_SUSPEND`) | Works: a follower shows a suspended pump | `sync_v3.http` |
| Profile switch and effective profile switch `original*` | Works: a follower reads an effective profile switch as one, not a Note | `sync_v3.http` |
| Running mode `mode`, `autoForced`, `reasons`, `originalDuration` | Works: a follower reads its mode, not UNKNOWN. An empty `reasons` is left out | `sync_v3.http` |
| `durationInMilliseconds` of a temp target ended early, extended carbs, therapy events | Works if within a minute of `duration`. A temp target ended early ends to the millisecond | `duration_ms.http` |
| `eventType` `SettingsExport` | Kept as a Note (was a 400 on every sync) | `settings_export.http` |
| Careportal or T1DCarer profile switch (`profile`, no `profileJson`) | Fixed: no `profileJson`, so AAPS switches to its own profile of that name (had one with only `units`, so `toProfileSwitch` dropped it) | `careportal_profile_switch.http`, `aaps_storage.sio` |
| Treatment download (`RemoteTreatment.timestamp` is a Long) | Fixed: API v3 and the storage socket give no `timestamp` (the ISO date API v1 gives Loop made AAPS drop every treatment and its page) | `careportal_download.http`, `aaps_storage.sio` |
| Treatment deleted offline, uploaded as a create with `isValid` false (`DataSyncSelectorV3`) | Fixed: kept deleted, sent as a delete (was valid, so AAPSClient showed deleted carbs as COB) | `created_deleted.http`, `aaps_storage.sio` |
| Delete of a treatment that a follower already deleted | Fixed: 200, as Nightscout returns (was 404) | `delete_again.http`, `sync_v3.http` |
| Treatment or profile store without `enteredBy` (Meal Bolus, Carb Correction, Profile Switch, profile) | Fixed: `enteredBy` is its `app`, `AAPS` (was the device name, for example `pwd_phone`) | `entered_by.http`, `loop/profile.http` |
| Treatment ended early (`duration` 0, `durationInMilliseconds` 43491); `RemoteTreatment.duration` is a Long | Fixed: API v3 and the storage socket give whole minutes (a fraction, kept for Loop, made AAPS drop it and its page) | `duration_ms.http`, `aaps_storage.sio` |
| **Device status, food and profiles** | | |
| Device status: create, history with `srvModified` | Works | `sync_v3.http`, `flow_v3.http` |
| Device status `pump.clock`, `pump.status`, `pump.extended` | Works: the pump object comes back as sent | `sync_v3.http` |
| Food: create | Works | `sync_v3.http` |
| Food: read | Works: `type` `food`. `portion`, `unit` and `gi` not kept yet (TODO P4) | `sync_v3.http` |
| Profile store: create, newest, history | Works | `sync_v3.http`, `flow_v3.http`, `history_paging.http` |
| **Socket.IO and client control** | | |
| `/storage` subscribe, `create`, `update`, `delete` | Works | `aaps_storage.sio`, `storage.sio` |
| `/alarm` subscribe, alarms, `clear_alarm`, `ack` | Works | `alarm.sio`, `aaps_alarm.sio` |
| `/alarm` `announcement` | Fixed in #511 (was never sent) | `aaps_alarm.sio` |
| `/alarm` `announcement` added on the site | Fixed: its notification's notes are the message (was empty) | `aaps_alarm.sio` |
| Client control: `settings` documents | Kept as sent, shared by the site's identities. Keys up to 128 characters for `aaps_clientcontrol_<kind>_<client UUID>` (over 36 was refused, so no client could pair) | `running_config.http` |
| Client control: v2 roles and subjects | Not supported (TODO P4) | — |
| v1 `authorize` with the SHA1 of the secret | Works (#512) | `aaps_v1.sio` |
| v1 `authorize` payload, `dataUpdate`, `dbAdd`, `dbUpdate` | Works | `aaps_v1.sio`, `db_writes.sio` |
