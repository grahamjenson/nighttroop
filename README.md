<p align="center">
  <img src="server/assets/static/logo.svg" alt="Nighttroop logo" width="120">
</p>

<h1 align="center">Nighttroop</h1>

<p align="center">
  <a href="https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml"><img src="https://github.com/grahamjenson/nighttroop/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI"></a>
</p>

Nighttroop is a multi-user rewrite of Nightscout. It keeps the Nightscout API, so Nightscout apps can use it.

## Goals

| Goal | What it means |
| ---- | ------------- |
| Simple | One binary to deploy. Few external dependencies, and all of them optional. |
| Cheap for each user | Each user uses very few server resources. |
| Multi-user | One server supports at least 10,000 users. |

## Safety

- **Warning:** Do not make treatment decisions from Nighttroop. It is not a medical device.
  - It does not give medical advice, and it does not replace medical care.
  - Its data and alarms can be late, missing or wrong.
  - See the [Terms of Service](server/assets/docs/TERMS_OF_SERVICE.md).
- **Caution:** The Dexcom Share and LibreLinkUp syncs use unofficial APIs that have no documentation.
  - The code is in `clients/dexcom_client` and `clients/libre_client`.
  - Other open-source followers use the same APIs.
  - These APIs can stop working at any time. Their use can be against the terms of those services.

## Try it

- Use the staging server: <https://staging.nighttroop.nz>.
- It runs the newest `main`.
- **Caution:** Do not keep data on it that you need. The server deletes its data from time to time.

## Run your own server

1. Read [docs/deploy.md](docs/deploy.md).
1. Pick a host guide in [docs/deploy](docs/deploy): [Voyager](docs/deploy/voyager/voyager.md),
   [Hostinger](docs/deploy/hostinger/hostinger.md) or [Fly.io staging](docs/deploy/staging_flyio/README.md).
1. Make the keysets for encryption, hashing and JWTs with `go run cmd/keysets/create_keysets.go`.
1. Put the keysets in the `KEY_SETS` env var on the server.
1. Start the server, then open `control.<host>/setup`.
1. Enter the one-time setup code from the server logs.
1. Keep the control admin's username, password and TOTP secret. The page shows them one time only.
1. Enter a code from the TOTP secret. The server then saves the control admin.

See [docs/config.md](docs/config.md) for all the env vars.

## Documents

| Topic | Document |
| ----- | -------- |
| All features | [REQUIREMENTS.md](REQUIREMENTS.md) |
| Configuration | [docs/config.md](docs/config.md) |
| Apps that use Nighttroop | [docs/clients](docs/clients): [AAPS](docs/clients/aaps.md), [Loop](docs/clients/loop.md), [Trio](docs/clients/trio.md), [iAPS](docs/clients/iaps.md), [LoopFollow](docs/clients/loopfollow.md), [xDrip+](docs/clients/xdrip.md), [xDrip4iOS](docs/clients/xdrip4ios.md), [Juggluco](docs/clients/juggluco.md), [Nightguard](docs/clients/nightguard.md) |
| Development | [DEVELOPMENT.md](DEVELOPMENT.md), [AGENTS.md](AGENTS.md), [AI_POLICY.md](AI_POLICY.md) |
| Database migrations | [docs/migrations.md](docs/migrations.md) |
| Data limits and truncation | [docs/truncation.md](docs/truncation.md) |
| Loop settings | [docs/loop_settings.md](docs/loop_settings.md) |
| Scripts | [docs/scripts.md](docs/scripts.md) |
| Work to do | [TODOs.md](TODOs.md) |

## Technology

Nighttroop uses the GoTHSS stack:

- **Go** for the server.
- **Templ** for the HTML templates.
- **htmx** for the pages.
- **SQLite** for the databases.
- **SQLc** for the database queries.
- **Tailwind** for the styles.

Generators make much of the code: SQLc, goverter, Templ and Tailwind. See [DEVELOPMENT.md](DEVELOPMENT.md).

## Security

To report a vulnerability, see [SECURITY.md](SECURITY.md).

## Licence

Nighttroop is free software under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).

- You can use, change and share it under the terms of that licence.
- If you run a changed Nighttroop for other people over a network, offer them its source. The AGPL asks for this.

### Third-party code and data

This repository, or its server image, contains these items. Each has its own licence.

| What | Where | Licence |
| ---- | ----- | ------- |
| [htmx](https://htmx.org) and its SSE extension | `server/assets/static/htmx.js`, `htmx_sse.js` | [0BSD](https://github.com/bigskysoftware/htmx/blob/master/LICENSE) |
| [Floating UI](https://floating-ui.com) core and dom 1.7.0 | `server/assets/static/floating_ui_core.js`, `floating_ui_dom.js` | [MIT](https://github.com/floating-ui/floating-ui/blob/master/LICENSE) |
| [QRCode.js](https://github.com/davidshimjs/qrcodejs) | `server/assets/static/qrcode.js` | [MIT](https://github.com/davidshimjs/qrcodejs/blob/master/LICENSE) |
| [Apache ECharts](https://echarts.apache.org) 5.6.0 | `server/assets/static/echarts.min.js` | [Apache-2.0](https://github.com/apache/echarts/blob/master/LICENSE), with its [NOTICE](https://github.com/apache/echarts/blob/master/NOTICE) |
| [Redoc](https://github.com/Redocly/redoc) 2.5.4 | `server/assets/static/redoc.standalone.js` | [MIT](https://github.com/Redocly/redoc/blob/main/LICENSE); `redoc.standalone.js.LICENSE.txt` lists the libraries it bundles |
| [templUI](https://templui.io) 1.11.2 components | `server/components/ui`, `utils/templui`, and their scripts in `server/assets/static` | [MIT](https://github.com/templui/templui/blob/main/LICENSE) |
| [Lucide](https://lucide.dev) icons | `server/components/ui/icon/icon_data.go` | [ISC](https://github.com/lucide-icons/lucide/blob/main/LICENSE) |
| [DB-IP](https://db-ip.com) IP to Country Lite | `utils/geoip/dbip-country-lite.mmdb` | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |
| [DB-IP](https://db-ip.com) IP to City Lite | The `Dockerfile` downloads it into the server image | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |

[IP Geolocation by DB-IP](https://db-ip.com):

- Nighttroop uses the free DB-IP Lite databases to find the location of a sign-in.
- The built-in country database is the DB-IP file, with no changes.
- Pages that show a location link to DB-IP, as its licence asks.

`go.mod` lists the Go modules that Nighttroop uses. Each module has its own licence.
