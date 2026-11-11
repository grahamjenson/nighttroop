# Benchmark

[wrk](https://github.com/wg/wrk) scripts to benchmark a local or deployed server.

`nightscout.lua` has each wrk thread create many sites and has each site, like a person's phone and followers, send
one request every couple of seconds to its Nightscout API, over v1, v2 and v3 with the site's token. Most are writes,
so the sites fill with data the site's pages can show:

- CGM readings 5 minutes apart walking back from now, and finger sticks
- treatments of many kinds, over v1 as xDrip and Loop send them and v3 as AAPS does: meal, correction and carb
  boluses, temp basals, temporary targets, profile switches, exercise, BG checks, site, sensor, insulin and battery
  changes, and notes
- Loop device statuses, profiles, food and activities

It prints, at the end, the sites it made, the records it sent of each kind, and the API's responses by status, so rate
limits (429) and errors (5xx) show up.

## Local

Start a blank server (`air`, which uses `.env.dev`), then:

```bash
wrk -t4 -c64 -d60s --latency -s tests/benchmark/nightscout.lua http://localhost:8080
```

That makes 200 sites (`-t` threads with `NS_SITES` sites each, default 50), each sending a request every 2 seconds
(`NS_INTERVAL`), so 100 requests a second over the `-c` connections. A site's 30 requests a minute stay under its
limits of 180 reads and 60 writes, and each site sends from its own `X-Forwarded-For` address, which the server trusts
from localhost, so the per IP limits are not hit either. Raise `NS_SITES` or `-t` for more load rather than lowering
`NS_INTERVAL`.

Without `.env.dev`'s raised limits a server makes 20 sites at once and then 7 a minute; the script waits 10 seconds
each time it is refused and keeps making sites as the limit allows.

## Deployed

A deployed server lets an IP create a site only every 10 minutes, so create a site in the browser and benchmark it
with its phone token:

```bash
NS_SUBDOMAIN=<subdomain> NS_TOKEN=<phone token> wrk -t1 -c4 -d60s --latency -s tests/benchmark/nightscout.lua https://<host>
```

Each thread sends to the site at its own pace, so keep `-t1` to stay under the site's limits. wrk connects to `<host>`
and sends `<subdomain>.ns.<host>` as the Host header.
