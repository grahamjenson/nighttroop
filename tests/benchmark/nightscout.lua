-- wrk script that benchmarks the Nightscout API with many sites, each sending the reads and writes an uploader and
-- follower make, slowly enough to stay under the site's rate limits.
--
--   wrk -t4 -c64 -d60s --latency -s tests/benchmark/nightscout.lua http://localhost:8080
--
-- Each thread creates NS_SITES sites (default 50) with POST /site, backing off for 10s on a 429, and has each send one
-- of the requests in `mix` every NS_INTERVAL seconds (default 2), to its Nightscout host (<subdomain>.ns.<host>, sent
-- as the Host header so wrk can connect to <host>). A site's 30 requests a minute keep under its limits of 180 reads
-- and 60 writes, and the connections (-c) only need to keep up with the sites: 200 sites send 100 requests a second.
-- Each site sends from its own X-Forwarded-For address, as each user has their own phone, which a server trusts from
-- localhost; so only the global site create limit holds back how fast sites are made.
-- Set NS_SUBDOMAIN and NS_TOKEN (a phone token) to send to an existing site instead, such as on a deployed server.
-- Each site is a person whose CGM readings walk back from now 5 minutes at a time, with their treatments, device
-- statuses, profiles, food and activities in between.

local sitesPerThread = tonumber(os.getenv("NS_SITES")) or 50
local interval = tonumber(os.getenv("NS_INTERVAL")) or 2
local timezones = { "Pacific/Auckland", "Australia/Sydney", "Europe/London", "America/New_York", "Asia/Tokyo" }

local threads = {}
local id
local sites, creating, creatingSince, retryAt = {}, 0, 0, 0
local counts, written, createFailures = {}, {}, 0

function setup(thread)
  thread:set("id", #threads)
  table.insert(threads, thread)
end

local function addSite(subdomain, token)
  table.insert(sites, {
    host = subdomain .. ".ns." .. wrk.host, token = token,
    ip = string.format("10.%d.%d.%d", id % 256, math.floor(#sites / 250), #sites % 250 + 1),
    timezone = timezones[math.random(#timezones)], clock = os.time() * 1000, bg = math.random(80, 180), due = 0,
  })
  wrk.thread:set("sites", #sites)
end

function init(args)
  id = wrk.thread:get("id")
  math.randomseed(os.time() + id)
  if os.getenv("NS_SUBDOMAIN") then
    sitesPerThread = 0
    addSite(os.getenv("NS_SUBDOMAIN"), os.getenv("NS_TOKEN"))
  end
end

local function iso(ms)
  return os.date("!%Y-%m-%dT%H:%M:%SZ", math.floor(ms / 1000))
end

local function get(s, path)
  return wrk.format("GET", path .. (path:find("?", 1, true) and "&" or "?") .. "token=" .. s.token,
    { Host = s.host, ["X-Forwarded-For"] = s.ip })
end

local function post(s, path, body)
  return wrk.format("POST", path .. "?token=" .. s.token,
    { Host = s.host, ["X-Forwarded-For"] = s.ip, ["Content-Type"] = "application/json" }, body)
end

-- at is a time within the site's latest 5 minutes, so records sit between its readings
local function at(s)
  return s.clock + math.random(0, 299) * 1000
end

-- sgv sends the reading at the site's clock and walks back to the one 5 minutes before it, which sets its direction
local function sgv(s)
  local bg, date = s.bg, s.clock
  s.clock = s.clock - 300000
  s.bg = math.max(45, math.min(380, bg + math.random(-12, 12) + math.floor((120 - bg) / 20)))
  local delta = bg - s.bg
  local direction = delta > 10 and "SingleUp" or delta > 5 and "FortyFiveUp" or delta < -10 and "SingleDown"
    or delta < -5 and "FortyFiveDown" or "Flat"
  return post(s, "/api/v1/entries", string.format(
    '[{"type":"sgv","sgv":%d,"date":%d,"dateString":"%s","direction":"%s","noise":1,"device":"wrk"}]',
    bg, date, iso(date), direction))
end

local treatments = {
  function() local c = math.random(10, 90) return string.format('"eventType":"Meal Bolus","carbs":%d,"insulin":%.2f', c, c / 12) end,
  function() return string.format('"eventType":"Correction Bolus","insulin":%.2f', math.random(5, 30) / 10) end,
  function() return string.format('"eventType":"Carb Correction","carbs":%d', math.random(5, 20)) end,
  function() return string.format('"eventType":"Temp Basal","rate":%.2f,"absolute":%.2f,"duration":30', math.random(0, 20) / 10, math.random(0, 20) / 10) end,
  function() return '"eventType":"Temporary Target","targetTop":140,"targetBottom":140,"units":"mg/dl","duration":60,"reason":"Activity"' end,
  function() return '"eventType":"Profile Switch","profile":"Default","duration":0,"percentage":100,"timeshift":0' end,
  function() return string.format('"eventType":"Exercise","duration":%d,"notes":"walk"', math.random(15, 90)) end,
  function() return string.format('"eventType":"BG Check","glucose":%d,"glucoseType":"Finger","units":"mg/dl"', math.random(60, 250)) end,
  function() return '"eventType":"Site Change","notes":"new site"' end,
  function() return '"eventType":"Sensor Start","notes":"new sensor"' end,
  function() return '"eventType":"Insulin Change"' end,
  function() return '"eventType":"Pump Battery Change"' end,
  function() return '"eventType":"Note","notes":"benchmark note"' end,
}

-- v1 treatments are as xDrip and Loop send them, v3 as AAPS does
local function treatment(s, v3)
  local fields = treatments[math.random(#treatments)]()
  if v3 then
    return post(s, "/api/v3/treatments", string.format('{%s,"app":"wrk","date":%d,"utcOffset":0}', fields, at(s)))
  end
  return post(s, "/api/v1/treatments", string.format('[{%s,"created_at":"%s","enteredBy":"wrk"}]', fields, iso(at(s))))
end

local function devicestatus(s)
  local t = iso(at(s))
  return post(s, "/api/v1/devicestatus", string.format(
    '{"device":"loop://wrk","created_at":"%s","uploader":{"battery":%d},"pump":{"battery":{"percent":%d},' ..
    '"reservoir":%d,"clock":"%s","suspended":false},"loop":{"timestamp":"%s","iob":{"iob":%.2f,"timestamp":"%s"},' ..
    '"cob":{"cob":%d,"timestamp":"%s"},"predicted":{"startDate":"%s","values":[%d,%d,%d,%d]}}}',
    t, math.random(10, 100), math.random(10, 100), math.random(5, 200), t, t, math.random(0, 50) / 10, t,
    math.random(0, 60), t, t, s.bg, s.bg + 3, s.bg + 5, s.bg + 6))
end

local function profile(s)
  return post(s, "/api/v1/profile", string.format(
    '{"defaultProfile":"Default","startDate":"%s","units":"mg/dl","store":{"Default":{"dia":5,"units":"mg/dl",' ..
    '"timezone":"%s","carbratio":[{"time":"00:00","timeAsSeconds":0,"value":%d}],' ..
    '"sens":[{"time":"00:00","timeAsSeconds":0,"value":%d}],"basal":[{"time":"00:00","timeAsSeconds":0,"value":%.2f},' ..
    '{"time":"06:00","timeAsSeconds":21600,"value":%.2f}],"target_low":[{"time":"00:00","timeAsSeconds":0,"value":90}],' ..
    '"target_high":[{"time":"00:00","timeAsSeconds":0,"value":140}]}}}',
    iso(at(s)), s.timezone, math.random(8, 15), math.random(30, 80), math.random(5, 15) / 10, math.random(5, 15) / 10))
end

-- Each request is picked with a chance of its weight over the total; writes are named for the counts at the end
local mix = {
  { 12, sgv, "sgv entries" },
  { 1, function(s) local t = at(s) return post(s, "/api/v1/entries", string.format(
    '[{"type":"mbg","mbg":%d,"date":%d,"dateString":"%s","device":"wrk"}]', s.bg, t, iso(t))) end, "mbg entries" },
  { 3, function(s) return treatment(s, false) end, "v1 treatments" },
  { 3, function(s) return treatment(s, true) end, "v3 treatments" },
  { 4, devicestatus, "device statuses" },
  { 1, profile, "profiles" },
  { 1, function(s) return post(s, "/api/v1/food", string.format(
    '{"type":"food","category":"Snacks","name":"Apple %d","portion":1,"unit":"pcs","carbs":%d,"gi":2}',
    math.random(1000), math.random(10, 30))) end, "food" },
  { 1, function(s) return post(s, "/api/v1/activity", string.format(
    '{"activityType":"walking","created_at":"%s","duration":%d,"steps":%d,"heartRate":%d}',
    iso(at(s)), math.random(600, 3600), math.random(500, 8000), math.random(80, 150))) end, "activities" },
  { 4, function(s) return get(s, "/api/v1/entries.json?count=10") end },
  { 1, function(s) return get(s, "/api/v1/entries.json?count=288") end },
  { 2, function(s) return get(s, "/api/v1/treatments.json?count=50") end },
  { 2, function(s) return get(s, "/api/v1/devicestatus.json?count=1") end },
  { 1, function(s) return get(s, "/api/v1/profile.json") end },
  { 1, function(s) return get(s, "/api/v1/status.json") end },
  { 1, function(s) return get(s, "/api/v2/properties") end },
  { 2, function(s) return get(s, "/api/v3/entries?limit=100&sort$desc=date") end },
  { 1, function(s) return get(s, "/api/v3/treatments?limit=100&sort$desc=date") end },
  { 1, function(s) return get(s, "/api/v3/lastModified") end },
}

local total = 0
for _, m in ipairs(mix) do
  total = total + m[1]
end

-- delay picks what a connection sends next, as wrk asks for the delay before the request, and request sends it: a site
-- to create, a site whose turn has come, or a health check while there are none
local queue = {}

function delay()
  local now = os.time()
  -- A create that has had no answer in 10s, such as one the global rate limit refused, is given up on
  if creating > 0 and now > creatingSince + 10 then
    creating = 0
  end
  -- Sites are made two at a time, so a refused burst does not send a create from every connection
  if creating < 2 and #sites + creating < sitesPerThread and now >= retryAt then
    creating = creating + 1
    creatingSince = now
    table.insert(queue, "create")
    return 0
  end

  local s
  for _, site in ipairs(sites) do
    if not s or site.due < s.due then
      s = site
    end
  end
  if not s then
    table.insert(queue, "health")
    return 1000
  end
  local wait = math.max(0, s.due - now) * 1000 + math.random(0, 999)
  s.due = math.max(s.due, now) + interval
  table.insert(queue, s)
  return wait
end

function request()
  local s = table.remove(queue, 1) or "health" -- wrk asks the first thread for a request to check it, without a delay
  if s == "health" then
    return wrk.format("GET", "/health")
  elseif s == "create" then
    return wrk.format("POST", "/site", {
      Accept = "application/json", ["Content-Type"] = "application/x-www-form-urlencoded",
      ["X-Forwarded-For"] = string.format("10.%d.%d.%d", id % 256, math.random(0, 255), math.random(1, 254)),
    }, "timezone=" .. timezones[math.random(#timezones)] .. "&language=en-NZ")
  end

  local pick = math.random(total)
  for _, m in ipairs(mix) do
    pick = pick - m[1]
    if pick <= 0 then
      if m[3] then
        written[m[3]] = (written[m[3]] or 0) + 1
        wrk.thread:set("written", written)
      end
      return m[2](s)
    end
  end
end

function response(status, headers, body)
  if body:find('"site_subdomain"', 1, true) then
    creating = math.max(0, creating - 1)
    addSite(body:match('"site_subdomain":"([^"]+)"'), body:match('"pwd_nstoken":"([^"]+)"'))
  elseif body:find('"Too many requests"', 1, true) or body:find("Unable to Create Site", 1, true) then
    -- The global site create limit is reached, so wait for it to refill
    creating = math.max(0, creating - 1)
    retryAt = os.time() + 10
    createFailures = createFailures + 1
    wrk.thread:set("createFailures", createFailures)
  elseif not body:match('^{"status":"ok"}%s*$') then
    counts[status] = (counts[status] or 0) + 1
    wrk.thread:set("counts", counts)
  end
end

local function sum(key)
  local totals = {}
  for _, thread in ipairs(threads) do
    local value = thread:get(key)
    if type(value) == "table" then
      for k, n in pairs(value) do
        totals[k] = (totals[k] or 0) + n
      end
    elseif value then
      totals[1] = (totals[1] or 0) + value
    end
  end
  return totals
end

local function show(title, totals)
  print(title)
  local keys = {}
  for k in pairs(totals) do
    table.insert(keys, k)
  end
  table.sort(keys, function(a, b) return tostring(a) < tostring(b) end)
  for _, k in ipairs(keys) do
    print(string.format("  %s: %d", k, totals[k]))
  end
end

function done(summary, latency, requests)
  print(string.format("Sites: %d, creates refused: %d", sum("sites")[1] or 0, sum("createFailures")[1] or 0))
  show("Records sent:", sum("written"))
  show("Nightscout API responses by status:", sum("counts"))
end
