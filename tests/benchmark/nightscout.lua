-- wrk script that benchmarks the Nightscout API with a mix of the reads and writes uploaders and followers make.
--
--   wrk -t4 -c64 -d30s --latency -s tests/benchmark/nightscout.lua http://localhost:8080
--
-- Each thread creates its own site with POST /site, gets a JWT for the site's phone token, then sends the requests in
-- `mix` to the site's Nightscout host (<subdomain>.ns.<host>, sent as the Host header so wrk can connect to <host>).
-- Set NS_SUBDOMAIN and NS_TOKEN to use an existing site instead, as servers limit how often an IP creates sites.
-- Set NS_READ_ONLY=1 to leave out the writes.
-- wrk's own stats include creating the site and the health checks waiting for it; the status counts printed at the end
-- are only the Nightscout API's.

local subdomain = os.getenv("NS_SUBDOMAIN")
local token = os.getenv("NS_TOKEN")
local readOnly = os.getenv("NS_READ_ONLY") == "1"

local threads = {}
local id, jwt, nsHost, phase, start, verifying
local writes = 0
local counts = {}

function setup(thread)
  thread:set("id", #threads)
  table.insert(threads, thread)
end

function init(args)
  id = wrk.thread:get("id")
  start = os.time() * 1000
  math.randomseed(start + id)
  phase = subdomain and "auth" or "create"
  -- wrk calls the first thread's request once to check it before connecting, so that one is not sent
  verifying = id == 0
  if subdomain then
    nsHost = subdomain .. ".ns." .. wrk.host
  end
end

-- date steps back from the start for each write, so writes on a site, from up to 100 threads, are not duplicates
local function date()
  writes = writes + 1
  return start - writes * 100 - id
end

local function get(path)
  return wrk.format("GET", path, { Host = nsHost, Authorization = "Bearer " .. jwt })
end

local function post(path, body)
  return wrk.format("POST", path, {
    Host = nsHost, Authorization = "Bearer " .. jwt, ["Content-Type"] = "application/json",
  }, body)
end

-- v1 authenticates with the token parameter, as xDrip and Loop do, and v2 and v3 with the JWT, as AAPS does
local function v1(path)
  return wrk.format("GET", path .. "&token=" .. token, { Host = nsHost })
end

-- Each request is picked with a chance of its weight over the total
local mix = {
  { 6, function() return v1("/api/v1/entries.json?count=10") end },
  { 2, function() return v1("/api/v1/entries.json?count=288") end },
  { 2, function() return v1("/api/v1/treatments.json?count=50") end },
  { 2, function() return v1("/api/v1/devicestatus.json?count=1") end },
  { 2, function() return v1("/api/v1/status.json?") end },
  { 2, function() return get("/api/v2/properties") end },
  { 3, function() return get("/api/v3/entries?limit=100&sort$desc=date") end },
  { 2, function() return get("/api/v3/treatments?limit=100&sort$desc=date") end },
  { 2, function() return get("/api/v3/lastModified") end },
  { 3, function()
    return post("/api/v1/entries", string.format(
      '[{"type":"sgv","sgv":%d,"date":%d,"direction":"Flat","device":"wrk"}]', math.random(60, 250), date()))
  end, write = true },
  { 1, function()
    return post("/api/v3/treatments", string.format(
      '{"app":"wrk","date":%d,"eventType":"Meal Bolus","carbs":%d,"insulin":2.5}', date(), math.random(5, 80)))
  end, write = true },
  { 1, function()
    return post("/api/v1/devicestatus", string.format(
      '{"device":"wrk://benchmark","date":%d,"uploader":{"battery":%d}}', date(), math.random(10, 100)))
  end, write = true },
}

local total = 0
for i = #mix, 1, -1 do
  if readOnly and mix[i].write then
    table.remove(mix, i)
  else
    total = total + mix[i][1]
  end
end

function request()
  if verifying then
    verifying = false
    return wrk.format("GET", "/health")
  elseif phase == "create" then
    phase = "creating"
    return wrk.format("POST", "/site", { Accept = "application/json" })
  elseif phase == "auth" then
    phase = "authorizing"
    return wrk.format("GET", "/api/v2/authorization/request/" .. token, { Host = nsHost })
  elseif phase ~= "ready" then
    -- Other connections wait for the site with health checks, see delay
    return wrk.format("GET", "/health")
  end

  local pick = math.random(total)
  for _, m in ipairs(mix) do
    pick = pick - m[1]
    if pick <= 0 then
      return m[2]()
    end
  end
end

-- Connections waiting for the site check in every 100ms, so their health checks do not swamp the results
function delay()
  return phase == "ready" and 0 or 100
end

local function fail(step, status, body)
  print(string.format("thread %d: %s failed with %d: %s", id, step, status, body))
  wrk.thread:stop()
end

function response(status, headers, body)
  if phase == "ready" then
    counts[status] = (counts[status] or 0) + 1
    wrk.thread:set("counts", counts)
  elseif phase == "creating" and status ~= 200 then
    return fail("creating a site", status, body)
  elseif phase == "creating" and body:find('"site_subdomain"') then
    subdomain = body:match('"site_subdomain":"([^"]+)"')
    token = body:match('"pwd_nstoken":"([^"]+)"')
    nsHost = subdomain .. ".ns." .. wrk.host
    phase = "auth"
  elseif phase == "authorizing" and not body:find('"status"') then
    if status ~= 200 then
      return fail("getting a JWT", status, body)
    end
    jwt = body:match('"token":"([^"]+)"')
    phase = "ready"
  end
end

function done(summary, latency, requests)
  local totals = {}
  for _, thread in ipairs(threads) do
    for status, n in pairs(thread:get("counts") or {}) do
      totals[status] = (totals[status] or 0) + n
    end
  end
  local statuses = {}
  for status in pairs(totals) do
    table.insert(statuses, status)
  end
  table.sort(statuses)
  print("Nightscout API responses by status:")
  for _, status in ipairs(statuses) do
    print(string.format("  %d: %d", status, totals[status]))
  end
end
