// Package usage counts how the server is used, in memory since it started: who is active each day, how much they
// send, and which routes are asked for, so its operator can see how many use it and how much, to plan its costs
package usage

import (
	"cmp"
	"math"
	"slices"
	"sync"
	"time"
)

// keepDays is how many days of activity are kept, today and the week before it
const keepDays = 8

// Tracker counts the server's use. A nil Tracker counts nothing, so tests without one need none.
type Tracker struct {
	mu      sync.Mutex
	started time.Time
	// days are each kept day's activity, by its UTC date
	days     map[string]*day
	ingested int64
	routes   map[string]*route
	cpu      cpuSample
}

type day struct {
	// requests and ingested are each active site's signed in requests and the bytes they sent
	requests   map[string]int64
	ingested   map[string]int64
	identities map[int64]struct{}
}

type route struct {
	count int64
	took  time.Duration
}

func New() *Tracker {
	return &Tracker{started: time.Now(), days: map[string]*day{}, routes: map[string]*route{}}
}

// Seen counts a site's signed in request by one of its identities, which sent bytes in its body
func (t *Tracker) Seen(siteID string, identityID int64, bytes int64, now time.Time) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	date := now.UTC().Format(time.DateOnly)
	d, ok := t.days[date]
	if !ok {
		d = &day{requests: map[string]int64{}, ingested: map[string]int64{}, identities: map[int64]struct{}{}}
		t.days[date] = d
		t.forgetOldDays(now)
	}
	d.requests[siteID]++
	d.ingested[siteID] += max(bytes, 0)
	d.identities[identityID] = struct{}{}
	t.ingested += max(bytes, 0)
}

func (t *Tracker) forgetOldDays(now time.Time) {
	oldest := now.UTC().AddDate(0, 0, -keepDays+1).Format(time.DateOnly)
	for date := range t.days {
		if date < oldest {
			delete(t.days, date)
		}
	}
}

// Route counts a request to a route, named by its router and pattern, which took took
func (t *Tracker) Route(name string, took time.Duration) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.routes[name]
	if !ok {
		r = &route{}
		t.routes[name] = r
	}
	r.count++
	r.took += took
}

// Activity is how many sites and identities were active, the requests they made and the bytes they sent
type Activity struct {
	Sites      int   `json:"sites"`
	Identities int   `json:"identities"`
	Requests   int64 `json:"requests"`
	Ingested   int64 `json:"ingested"`
}

// Stats is the mean and standard deviation of N values
type Stats struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
	N    int     `json:"n"`
}

// RouteCount is how often a route was asked for, and how long it took on average, in milliseconds
type RouteCount struct {
	Route  string  `json:"route"`
	Count  int64   `json:"count"`
	MeanMs float64 `json:"mean_ms"`
}

// Snapshot is the server's use since Started, in milliseconds
type Snapshot struct {
	Started int64 `json:"started"`
	// Today is since midnight UTC, and Week the last 7 days including today
	Today Activity `json:"today"`
	Week  Activity `json:"week"`
	// Ingested is every byte sent to sites since the server started
	Ingested int64 `json:"ingested"`
	// RequestsPerSiteDay and IngestedPerSiteDay are over each site's active days in the week
	RequestsPerSiteDay Stats        `json:"requests_per_site_day"`
	IngestedPerSiteDay Stats        `json:"ingested_per_site_day"`
	Routes             []RouteCount `json:"routes"`
}

// Snapshot is the use so far, with the limit most asked for routes
func (t *Tracker) Snapshot(now time.Time, limit int) Snapshot {
	if t == nil {
		return Snapshot{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	snapshot := Snapshot{Started: t.started.UnixMilli(), Ingested: t.ingested}
	weekSites, weekIdentities := map[string]struct{}{}, map[int64]struct{}{}
	requests, ingested := []float64{}, []float64{}
	for i := range 7 {
		d, ok := t.days[now.UTC().AddDate(0, 0, -i).Format(time.DateOnly)]
		if !ok {
			continue
		}
		activity := d.activity()
		if i == 0 {
			snapshot.Today = activity
		}
		snapshot.Week.Requests += activity.Requests
		snapshot.Week.Ingested += activity.Ingested
		for site, count := range d.requests {
			weekSites[site] = struct{}{}
			requests, ingested = append(requests, float64(count)), append(ingested, float64(d.ingested[site]))
		}
		for identity := range d.identities {
			weekIdentities[identity] = struct{}{}
		}
	}
	snapshot.Week.Sites, snapshot.Week.Identities = len(weekSites), len(weekIdentities)
	snapshot.RequestsPerSiteDay, snapshot.IngestedPerSiteDay = stats(requests), stats(ingested)
	for name, r := range t.routes {
		snapshot.Routes = append(snapshot.Routes, RouteCount{Route: name, Count: r.count, MeanMs: float64(r.took.Microseconds()) / float64(r.count) / 1000})
	}
	slices.SortFunc(snapshot.Routes, func(a, b RouteCount) int { return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Route, b.Route)) })
	snapshot.Routes = snapshot.Routes[:min(len(snapshot.Routes), limit)]
	return snapshot
}

func (d *day) activity() Activity {
	activity := Activity{Sites: len(d.requests), Identities: len(d.identities)}
	for site, count := range d.requests {
		activity.Requests += count
		activity.Ingested += d.ingested[site]
	}
	return activity
}

// stats is the mean and population standard deviation of the values
func stats(values []float64) Stats {
	if len(values) == 0 {
		return Stats{}
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean, squares := sum/float64(len(values)), 0.0
	for _, v := range values {
		squares += (v - mean) * (v - mean)
	}
	return Stats{Mean: mean, Std: math.Sqrt(squares / float64(len(values))), N: len(values)}
}
