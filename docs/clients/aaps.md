# AAPS (AndroidAPS) and Nighttroop

How AAPS talks to Nightscout, checked against Nighttroop. Read from
[nightscout/AndroidAPS](https://github.com/nightscout/AndroidAPS) at tag `4.0.0` (commit `f77abd9`, the same as
`master` on 2026-11-02) for NSClient v3, and at tag `3.4.2.6` for NSClient v1, which AAPS 4.0 removed. Paths are
from the repository's root. `core/nssdk/.../` is short for `core/nssdk/src/commonMain/kotlin/app/aaps/core/nssdk/`,
and `nsclientV3/` for `plugins/sync/src/commonMain/kotlin/app/aaps/plugins/sync/nsclientV3/`.

The tests are `tests/api/nightscout/aaps/*.http` and `tests/socket/nightscout/aaps_*.sio`.

## Setup AAPS needs on Nighttroop

- **Glucose is mg/dL.** AAPS reads an entry's `sgv` as mg/dL and ignores its `units`
  (`nsclientV3/extensions/GlucoseValueExtension.kt` `NSSgvV3.toGV`). It knows only the units `mg/dl` and `mmol`
  (`core/nssdk/.../localmodel/entry/NsUnits.kt`, and `GlucoseUnit.fromText` in `core/data`); anything else, such
  as Nighttroop's `mmol/L`, is read as mg/dL. The Nightscout API always gives glucose, temp targets, BG checks and
  profiles in mg/dL, as Nightscout does, whatever the identity's or site's units. Uploads are fine either way: AAPS
  labels them `mg/dl`.
- The URL is the site's Nightscout host. AAPS adds `https://` when it has no scheme and `/api/` after it
  (`core/nssdk/.../networking/NsUrl.kt:21`).
- The token is a Nighttroop identity token. A master (the phone with the pump) needs one that may write; a device
  added for AAPS is always the PWD Phone (#527). A follower (AAPSClient) can use a read only one, added as Other.

## NSClient v3 (AAPS 4.0, and 3.x when "Use API v3" is chosen)

### Authentication

Every request goes through one interceptor (`core/nssdk/.../networking/NsAuth.kt`):

- it sends `Date: <now in milliseconds>` and `Authorization: Bearer <jwt>`, the JWT empty until it has one
  (lines 46–47). Nighttroop ignores `Date`;
- on a **401 or 403** (line 127) it first checks the body for Nightscout's date tolerance message (line 55), then
  gets a JWT from `GET /api/v2/authorization/request/<token>`, built from the host only, without a sub path
  (line 98), and with no `Authorization` header. It sends the request once more with the new JWT (lines 63–65);
- the refresh's answer must have `token`, `iat` and `exp` (`core/nssdk/.../remotemodel/RemoteAuthResponse.kt`). A
  401 or 403 from the refresh is "invalid access token", which AAPS shows and does not retry.

The socket's `subscribe` sends the token itself, not the JWT.

### What it sends and reads

The endpoints are all in `core/nssdk/.../networking/NightscoutApi.kt`; the handling of each answer is in
`core/nssdk/.../NSAndroidClientImpl.kt`. Query keys keep their `$` (`date$gt`, not `date%24gt`). A v3 read is
`{"result": …}`; a create, patch and delete is read as `{identifier, isDeduplication, deduplicatedIdentifier,
lastModified}` at the top level (`remotemodel/RemoteStatusResponse.kt`, `RemoteCreateUpdateResponse`).

JSON is decoded leniently (`core/nssdk/.../NsSdkJson.kt`): unknown keys are ignored, a quoted number is read into a
number, an unknown `eventType` is skipped. But a whole number field (`Int`/`Long`) given a fraction fails the whole
page, and AAPS then asks for the same page again for ever.

| When | Request | What AAPS reads |
| --- | --- | --- |
| Each sync, first | `GET /api/v3/status` | `version`, `apiVersion`, `srvDate`, `storage.{storage,version}` and `apiPermissions` with all six collections, each a string of `c`, `r`, `u`, `d` (`RemoteStatusResponse`). A master uploads only when all six are `crud`, a client when it may read all and create and update `settings` (`localmodel/ApiPermissions.kt`, `NSClientV3Plugin.kt:534`) |
| Status not full | `GET /api/v2/authorization/debug/check/api:settings:create` and `…:update` | 200 is yes, 401 or 403 no, anything else fails the status (`NSAndroidClientImpl.kt:119`) |
| Each sync | `GET /api/v3/lastModified` | `collections.{devicestatus,entries,profile,treatments,foods,settings}` as numbers; it reads `foods`, not Nightscout's `food`, so that one is always 0 (`remotemodel/LastModified.kt`) |
| First load of entries | `GET /api/v3/entries?sort=date&date$gt=<ms>&limit=500` from 100 days ago (`NSClientV3Plugin.kt:185,232`), then from the newest `date` it got. As it sends no other filter, each page is a full 500, the oldest first | `type` must be `sgv`, `sgv`, `date`, `identifier`, `direction`, `utcOffset`, `isValid`, `srvModified` (`mapper/SvgMapper.kt`) |
| Later loads of entries | `GET /api/v3/entries/history/<ms>?limit=500`, from the last `ETag` | the same, `isValid: false` for a deletion, and the `ETag` header `W/"<ms>"`, read by position (`NsHttpResponse.kt:58`), as the next `<ms>`. It stops when the ETag does not move (`workers/LoadBgRunner.kt`) |
| A reading or calibration | `POST /api/v3/entries` | 200 or 201 and `identifier`. A 400 with "Bad or missing utcOffset field" is sent again with `utcOffset: 0` (`NSAndroidClientImpl.kt:226`) |
| Reading changed / invalidated | `PATCH /api/v3/entries/<id>` without `date` and `utcOffset` / `DELETE /api/v3/entries/<id>` | 2xx, or 404, which it takes as done (line 259) |
| First load of treatments | `GET /api/v3/treatments?sort=created_at&created_at$gt=<ISO>&limit=500` | see the treatments below |
| Later loads | `GET /api/v3/treatments/history/<ms>?limit=500` | as entries, with the `ETag` |
| A treatment | `POST /api/v3/treatments` | as entries. A 400 is logged and **skipped for good**; a status other than 200, 201, 400 or 404 stops the upload, to try again (`NSClientV3Plugin.kt:1079–1091`) |
| Treatment changed / invalidated | `PATCH` / `DELETE /api/v3/treatments/<id>` | as entries |
| A device status (each loop) | `POST /api/v3/devicestatus` | 200 or 201 |
| Master: the newest | `GET /api/v3/devicestatus/history/<now-7min>?limit=500` | `srvModified`, `date`, `device`, `uploaderBattery` (Int), `isCharging`, `pump.{clock,reservoir,battery.percent (Int),status.status,extended}`, `openaps.{suggested,enacted,iob}` (`remotemodel/RemoteDeviceStatus.kt`) |
| Client: all of the last 48 hours | the same, paged from the newest `srvModified` it got (`workers/LoadDeviceStatusRunner.kt`) | without `pump.clock` the pump's data is skipped (`nsclientV3/data/NSDeviceStatusHandler.kt:225`) |
| Foods | `GET /api/v3/food?limit=1000`, and `POST /api/v3/food`. Update and delete are never sent (`NightscoutApi.kt:216`) | only documents of `type` `food`; `carbs`, `gi`, `energy`, `protein`, `fat` as Int (`remotemodel/RemoteFood.kt`, `mapper/FoodMapper.kt`) |
| First load of profiles | `GET /api/v3/profile?sort$desc=date&limit=1` | the store as JSON: `defaultProfile`, `startDate`, `store`, `srvModified` (`workers/LoadProfileStoreRunner.kt`) |
| Later loads | `GET /api/v3/profile/history/<ms>?limit=10` | **the last** in the list, as the newest |
| Profile changed | `POST /api/v3/profile` with `app: AAPS` added | 200 or 201 |
| Client control (4.0) | `GET/POST/PATCH/PUT/DELETE /api/v3/settings[/<id>]` for `aaps`, `aaps-state` and `aaps_clientcontrol_*` documents; `/api/v2/authorization/roles` and `/subjects` with an admin token | used for pairing clients and remote commands |

### Bodies AAPS sends

Built in `core/nssdk/.../mapper/TreatmentMapper.kt` (`toRemoteTreatment`, line 397) from
`nsclientV3/extensions/*Extension.kt`. Nulls are left out, `false` and `0` are sent, `app` is set to `AAPS` on a
create, and `identifier` is sent on an update. Every treatment has `date` (ms), `utcOffset` (minutes), `isValid`,
`isReadOnly`, `eventType`, and the pump's `pumpId`, `pumpType`, `pumpSerial` and `endId` when it has them.

- **Bolus** (`Meal Bolus`, `Correction Bolus`): `insulin`, `type` (`NORMAL`, `SMB`, `PRIMING`), `isBasalInsulin`.
- **Carbs** (`Carb Correction`, `Meal Bolus`): `carbs`, `duration` (min), `durationInMilliseconds`.
- **Temp Basal**: `type` (`NORMAL`, `EMULATED_PUMP_SUSPEND`, `PUMP_SUSPEND`, `SUPERBOLUS`, `FAKE_EXTENDED`); see
  [Temp basal percent](#temp-basal-percent).
- **Temporary Target**: `targetTop`, `targetBottom` in mg/dL with `units: "mg/dl"`, `reason`, `duration`,
  `durationInMilliseconds`.
- **Profile Switch**: `profile`, `profileJson` (a JSON string), `originalProfileName`, `originalDuration`,
  `duration`, `durationInMilliseconds`, `timeshift` (h), `percentage` (Int).
- **Effective profile switch**: a `Note` with `profileJson`, `originalProfileName`, `originalCustomizedName`,
  `originalTimeshift`, `originalPercentage`, `originalDuration`, `originalEnd`. AAPS reads it back as one only with
  all of them (`TreatmentMapper.kt:198`).
- **Running mode**: `OpenAPS Offline` with `duration`, `durationInMilliseconds`, `reason`, `mode`, `autoForced`,
  `reasons`, `originalDuration`.
- **Therapy events** (`Site Change`, `Sensor Change`, `BG Check`, `Note`, `Exercise`, `Announcement`, …):
  `duration`, `durationInMilliseconds`, `glucose`, `glucoseType`, `units`, `enteredBy`, `location`, `arrow`, `notes`.
- **Bolus Wizard**: `bolusCalculatorResult` (a JSON string), `glucose`.
- **Extended bolus**: `Combo Bolus` with `enteredinsulin`, `relative` (U/hr), `splitNow: 0`, `splitExt: 100`,
  `isEmulatingTempBasal`, or a `Temp Basal` with `extendedEmulated` holding it.
- **Entries**: `type: "sgv"`, `sgv` (mg/dL), `date`, `device`, `direction`, `noise`, `filtered`, `unfiltered`,
  `units: "mg/dl"`; a calibration is `type: "mbg"` with `mbg`, `sensorMgdlAtPairing` and `isCalibration: true`.
- **Device status**: `date`, `device` (`openaps://<phone>`), `uploaderBattery`, `isCharging`, `uploader: {}`,
  `pump.{clock,reservoir,battery,status,extended}`, `openaps.{suggested,enacted,iob}`, `configuration`.

### Temp basal percent

Nighttroop reads a temp basal's `percent` as Nightscout's **change** from the basal, which is what AAPS means:

- Sending, `TB.toNSTemporaryBasal` (`nsclientV3/extensions/TemporaryBasalExtension.kt:27–44`): `rate` is always the
  absolute rate in U/hr (`convertedToAbsolute`, line 34); a temp basal set in U/hr also has `absolute` (line 36); one
  set as a percent of the basal, as Dana, Combo and Insight pumps do, has `percent = rate - 100` (line 37) and no
  `absolute`. So 50% of the basal is `percent: -50`, 120% is `percent: 20`, and a zero temp is `percent: -100`.
  `isAbsolute` is not sent: it is not a field of `RemoteTreatment`. The model says so too:
  `NSTemporaryBasal.kt:32` "when sending to NS (rate - 100)", `RemoteTreatment.kt:61` "Eventual basal change in
  percent".
- Reading, `RemoteTreatment.toTreatment` (`TreatmentMapper.kt:167–196`): `isAbsolute = absolute != null` and
  `rate = absolute ?: percent + 100`, so `absolute` wins, and a percent alone is taken as a change.
- AAPS 3.4's NSClient v1 does the same: `TB.toJson` (`plugins/sync/src/main/kotlin/app/aaps/plugins/sync/nsclient/
  extensions/TemporaryBasalExtension.kt:23–26`) sends `rate` (absolute) and `absolute`, or `percent: rate - 100`, and
  `temporaryBasalFromJson` (lines 54–62) reads `absolute`, else `percent + 100`.

On Nighttroop: a temp basal keeps its `rate` in U/hr and a `multiplier` of the basal, so every percent from -100 (a
zero temp) up is kept as `1 + percent / 100` (0.5 for -50) and returned as the same percent. The rate AAPS sent is
kept as its rate, which AAPS followers read first as `absolute`; a percent sent without one gets the profile's basal
at its start times the multiplier, as Nightscout does, and a rate sent without a percent gets its percent the same
way. Every temp basal is returned with `rate`, `absolute` and, when it has one, `percent`.

### Socket.IO

AAPS uses socket.io-client-java 2.1.2 (an `EIO=4` client) on two namespaces of the site's URL
(`nsclientV3/ws/SocketNsConnection.kt:136–170`), when "Use websockets" is on.

- **`/storage`**: on connect it emits `subscribe` with `{"accessToken": <token>, "collections": ["devicestatus",
  "entries", "profile", "treatments", "foods", "settings"]}` and reads `success` from the ack
  (`nsclientV3/ws/NsConnectHandler.kt:53–73`). It listens for `create` and `update` (`colName`, `doc`, and
  `doc.srvModified`, without which the frame is dropped) and `delete` (`colName`, `identifier`), and acts on
  treatments, entries, device statuses, profiles, foods (as `foods`) and settings (`nsclientV3/ws/NsFrameHandler.kt:88–223`).
- **`/alarm`**, when notifications from alarms or announcements are on: it emits `subscribe` with
  `{"accessToken": <token>}`, then listens for `announcement`, `alarm`, `urgent_alarm` and `clear_alarm`, reading
  `level`, `group`, `title`, `message` and `eventName` (`nsclientV3/NSAlarmObject.kt`). An alarm snoozed in AAPS is
  acked with `emit("ack", level, group, silenceMillis)` (`jvmSharedMain/.../ws/SocketIoNsSocket.kt:51`).

### Quirks

- A 400 from a create is skipped for good, so anything Nighttroop refuses is lost to it.
- A PATCH or DELETE that answers 404 is taken as done; any other status stops the upload queue until it works.
- `pumpId` comes back from Nighttroop as a string; AAPS reads it into its `Long` anyway.
- `programmed`, which Nighttroop adds to a bolus for Loop, is not a field of `RemoteTreatment`, so AAPS ignores it
  (`ignoreUnknownKeys`, `core/nssdk/.../NsSdkJson.kt`).
- A create's `identifier` must be at the top level of the answer, not only in `result`.
- Nighttroop's `created_at$gt` is the treatment's time, as Nightscout's (#514); `srvCreated$gt` is when it reached
  the server.
- The first load goes back 100 days, in pages of 500.

## NSClient v1 (AAPS 3.4 and older)

Socket.IO only, no REST: socket.io-client-java 2.1.2 on the site's default namespace, and only over `https`
(`plugins/sync/src/main/kotlin/app/aaps/plugins/sync/nsclient/services/NSClientService.kt:277–300`).

- On connect it emits `authorize` with `{"client": "Android_AAPS", "history": 48, "status": true, "from":
  <newest ms it has>, "secret": <SHA1 of the API secret>}` (lines 361–375) and reads `read`, `write` and
  `write_treatment` from the ack (`nsclient/acks/NSAuthAck.kt`). It writes only when it may.
- It listens for `dataUpdate` (`status`, `profiles` (the last as the newest), `treatments` with `action` `update`
  or `remove`, `devicestatus`, `food`, `mbgs`, `cals`, `sgvs`, `delta`), and `announcement`, `alarm`,
  `urgent_alarm` and `clear_alarm` (lines 286–290, 466–580).
- It writes with `dbAdd` `{collection, data}`, reading `_id` from the first of the ack's array
  (`nsclient/acks/NSAddAck.kt`), and `dbUpdate` `{collection, _id, data}`, reading `result: "success"`
  (`nsclient/acks/NSUpdateAck.kt`). It acks alarms with `ack` as v3 does (line 630).

AAPS 3.4 always sends the SHA1 of what is in its API secret field. Nighttroop finds a token by its SHA1, as for
`api-secret` (#512), so AAPS 3.4 authorizes with a device's token as its API secret; `aaps_v1.sio` sends the token's
SHA1, as AAPS does.

## Status on Nighttroop

| Interaction | Status | Test |
| --- | --- | --- |
| JWT from `/api/v2/authorization/request/<token>`, 401 without one, 401 for a bad token | works | `auth_v3.http` |
| `/api/v3/status` fields | works | `auth_v3.http` |
| `/api/v3/status` permissions of a read only token | fixed in #511 (was always `crud`, so AAPS tried to upload with a read only token) | `auth_v3.http` `read_status` |
| `/api/v2/authorization/debug/check/<permission>` | fixed in #511 (was a 404, which would fail AAPS's status once it showed real permissions) | `auth_v3.http` |
| `/api/v3/lastModified` | works | `auth_v3.http`, `sync_v3.http` |
| Entries: create, duplicate, search by date, history with `ETag`, patch, delete, deletion in history | works | `sync_v3.http` |
| Entries' `filtered` and `unfiltered` of 0, sent with every reading | not kept, as 0 is no raw signal, and left out, which AAPS does not read back | `flow_v3.http` |
| Entries and treatments in an identity in mmol/L | works: the Nightscout API always gives them in mg/dL | `units/*.http` |
| First loads over 500 documents | works: a search sorted by date is the oldest first | `first_load_paging.http` |
| History over 500 documents, and profile history's newest | works: history is the oldest `srvModified` first | `history_paging.http` |
| Treatments: bolus and SMB `type`, carbs, temp target, profile switch, BG check, delete in history | works | `sync_v3.http`, `flow_v3.http` |
| Bolus Wizard's bolus and carbs, two Meal Boluses at the same `date` | fixed: kept as two treatments, each with its own identifier, so the bolus keeps its insulin, where Nightscout replaces it with the carbs, and deleting the carbs in AAPS deletes only them (they were merged into the bolus, so deleting them deleted the bolus) | `wizard_dedup_v3.http` |
| Bolus `icfg` and `isBasalInsulin`, without which AAPS takes the insulin in its settings and a bolus as not basal (`BolusExtension.kt` `NSBolus.toBolus`) | works: kept and returned as sent; Nighttroop's summaries still count a basal insulin bolus as a bolus, TODO P3 | `sync_v3.http` |
| Bolus Wizard `bolusCalculatorResult`, which AAPS restores the record from and without which it drops it (`TreatmentMapper.kt:270`) | works: kept and returned exactly as sent; one not JSON or over 4096 characters is left out | `bolus_wizard.http` |
| `units` written as AAPS writes them | works for a device added for AAPS, whose Nightscout API writes mg/dL as `mg/dl`; another device writes `mg/dL`, which AAPS also reads as mg/dL | `bolus_wizard.http` |
| Temp basal at, over or under 100% (any `percent` from -100), its update | works: `percent`, the change from the basal, is kept as a multiplier of it and returned as sent | `sync_v3.http`, `treatments/temp_basal_percent.http` |
| Profile switch and effective profile switch `original*` | works: kept with the switch and returned as sent, so a follower reads an effective profile switch as one, not a Note | `sync_v3.http` |
| Temp basal `type` (e.g. `PUMP_SUSPEND`) | works: kept and returned as sent, so a follower shows a suspended pump as one | `sync_v3.http` |
| Running mode `mode`, `autoForced`, `reasons`, `originalDuration` | works: kept with the treatment and returned as sent, so a follower reads its mode, not UNKNOWN; an empty `reasons` is not kept and is left out | `sync_v3.http` |
| `durationInMilliseconds` of a temp target ended early, extended carbs and therapy events | works: kept with the treatment and returned as sent while within a minute of its `duration`, so a temp target ended early ends to the millisecond | `duration_ms.http` |
| Calibrations (`isCalibration`, `sensorMgdlAtPairing`) | works: kept with the mbg and returned as sent, so a follower reads a calibration as one | `sync_v3.http` |
| Device statuses: create, history with `srvModified` | works | `sync_v3.http`, `flow_v3.http` |
| Device status `pump.clock`, `pump.status`, `pump.extended` | works: the pump object is read back as sent | `sync_v3.http` |
| Food: create | works | `sync_v3.http` |
| Food: read | works: each food has `type` `food`; its `portion`, `unit` and `gi` are not kept yet, TODO P4 | `sync_v3.http` |
| Profile store: create, newest, history | works | `sync_v3.http`, `flow_v3.http`, `history_paging.http` |
| `/storage` subscribe, `create`, `update`, `delete` | works | `aaps_storage.sio`, `storage.sio` |
| `/alarm` subscribe, alarms, `clear_alarm`, `ack` | works | `alarm.sio`, `aaps_alarm.sio` |
| `/alarm` `announcement` | fixed in #511 (was never sent) | `aaps_alarm.sio` |
| Client control: `settings` documents | kept as sent, shared by the site's identities | `running_config.http` |
| Client control: v2 roles and subjects | not supported; TODO P4 | — |
| v1 `authorize` with the SHA1 of the secret | works (#512) | `aaps_v1.sio` |
| v1 `authorize` payload, `dataUpdate`, `dbAdd`, `dbUpdate` | works | `aaps_v1.sio`, `db_writes.sio` |
| `POST treatments` with `eventType` `SettingsExport` | kept as a Note (was a 400 on every sync) | `settings_export.http` |
| Downloading treatments (`RemoteTreatment.timestamp` is a Long) | fixed: API v3 and the storage socket give no `timestamp`. Every treatment had the ISO date that API v1 gives Loop, so AAPS dropped every one it downloaded, and its whole page | `careportal_download.http`, `aaps_storage.sio` |
