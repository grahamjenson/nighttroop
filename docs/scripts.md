# Test scripts

The API, Socket.IO and browser tests are scripts, read by one parser in one package (`tests/dsl/flow`):

- `.http` files (`tests/api`) send API requests,
- `.sio` files (`tests/socket`) also open Socket.IO connections and live update streams,
- `.bdd` files (`tests/bdd`) drive the pages in Chromium,
- `.flow` files mix all three, e.g. `tests/bdd/site/live_mix.flow`.

The package imports nothing from nighttroop, so it can be used on its own: a `flow.Runner` is given its HTTP
client, a Socket.IO client for CONNECT (`NewSocket`, nighttroop's `socketio.NewClient` here) and a page for the
browser steps.

## Writing one

A line `HTTP` or `BROWSE` switches the lines after it:

- **HTTP steps** are written as in a `.sio` file. Each starts with `NAME`, then one of:
  - an HTTP request (`GET <url>`, `POST <url>`…, with `H` headers and a `B` JSON body);
  - `CONNECT <url>` (with `NS` and a `B` auth), `EMIT <event>` (with `B` or `ARGS`, and `ACK`), or `DISCONNECT`;
  - `STREAM <url>`, which opens a page's live updates (`…/events`) with its `H` headers;
  - `WAIT <event>` and `COLLECT <event>`, for an event from the socket or a stream, such as `created`;
  - or `RUN_BACKGROUND_JOBS`.
- **Browser steps** are written as in a `.bdd` file: `GOTO`, `CLICK`, `INPUT`, `TYPE` and the rest.
- **Asserts and `SET`** after a step check or keep what it got.
- **Where a file starts:** a `.bdd` file starts with browser steps and any other with HTTP ones. A `.steps` file goes on as the file that `INCLUDE`s it was, and a switch in an included file lasts until it ends.
- **Values** are filled in from `{{ }}`:
  - `{{ .name.field }}` is from an earlier step, and `{{ dig "name.0.field" }}` too;
  - `{{ site_url "abcdefgh" }}` is a site's address, and `{{ server_url }}` the server's;
  - `{{ totp .create_site.totp_secret }}` is the code an authenticator app shows now;
  - `{{ env "PASSWORD" }}` is an environment variable.

## Running one against staging

`cmd/nighttroop-script` runs scripts against a server running elsewhere, as staging, to make sites with real data and check them:

```
go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root scripts/staging -test.v new_site.flow
```

- **URLs:** a URL on `localhost`, or a host under it such as `abcdefgh.ns.localhost`, goes to that name under `--host`, over `--scheme` (https by default), so a script runs as written in a test and against staging.
- **Values and secrets:** `--var KEY=VALUE` sets `{{ .KEY }}`. Give passwords and TOTP secrets in the environment, for `{{ env }}`, so they are never written in a script.
- **Background jobs:** a script with `RUN_BACKGROUND_JOBS` is refused, as only a test's own server can run its jobs at once. On staging, data is written through the API instead, as a phone would.
- **Browser steps** need Playwright's Chromium (`go run github.com/mxschmitt/playwright-go/cmd/playwright install chromium`). `--headed` shows the browser, `--timeout` is how long each step waits, and screenshots, `SCREENSHOT`'s and a failed script's, go to `--out`.
- **Running tests:** each script runs as a test, so `-test.v` prints every step and its response, and `-test.run` picks which scripts run.

`scripts/staging` has two:

- **`new_site.flow`** makes a site with an hour of readings, a meal bolus and a phone's status, and prints the site's credentials in its `create_site` response.
- **`sign_in.flow`** signs in to a site's page with them and screenshots its summary:

```
SITE=abcdefgh PASSWORD=... TOTP_SECRET=... go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root scripts/staging sign_in.flow
```
