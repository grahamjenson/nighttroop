# Benchmark

[wrk](https://github.com/wg/wrk) scripts to benchmark a local or deployed server.

`nightscout.lua` has each wrk thread create a site and then send its Nightscout API a weighted mix of the reads and
writes uploaders and followers make, over v1 (with the token parameter), v2 and v3 (with a JWT). It prints the
API's responses by status at the end, so rate limits (429) and errors (5xx) show up.

## Local

Run the server (`air`), which with `.env.dev` has no rate limits, then:

```bash
wrk -t4 -c64 -d30s --latency -s tests/benchmark/nightscout.lua http://localhost:8080
```

`-t` is the number of sites and `-c` the number of connections shared between them.

## Deployed

A deployed server lets an IP create a site only every 10 minutes and a site make 180 reads and 60 writes a minute, so
create a site in the browser and benchmark it with its phone token, or raise the limits in the server's environment
(see `.env.dev`):

```bash
NS_SUBDOMAIN=<subdomain> NS_TOKEN=<phone token> wrk -t4 -c64 -d30s --latency -s tests/benchmark/nightscout.lua https://<host>
```

`NS_READ_ONLY=1` leaves out the writes. wrk connects to `<host>` and sends `<subdomain>.ns.<host>` as the Host header.
The JWT lasts an hour, so keep runs shorter than that.
