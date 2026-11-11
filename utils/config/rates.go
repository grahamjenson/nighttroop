package config

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/netip"
	"slices"
	"strconv"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"
)

// This contains all rate limiters in the system
// This might be more efficient if joined together, but right now this is fine
type RateLimiters struct {
	// rates are each rate limit's rate, from the env unless a control admin set it
	rates *rateSettings

	userLoginRateLimiters *lru.Cache[string, *rate.Limiter]
	ipLoginRateLimiters   *lru.Cache[string, *rate.Limiter]

	siteUnauthRateLimiters *lru.Cache[string, *rate.Limiter]

	identityGETRateLimiters *lru.Cache[string, *rate.Limiter]
	siteGETRateLimiters     *lru.Cache[string, *rate.Limiter]

	identityPOSTRateLimiters *lru.Cache[string, *rate.Limiter]
	sitePOSTRateLimiters     *lru.Cache[string, *rate.Limiter]

	siteUpdateConnectorsRateLimiters *lru.Cache[string, *rate.Limiter]
	siteIdentityCreateRateLimiters   *lru.Cache[string, *rate.Limiter]
	siteDownloadRateLimiters         *lru.Cache[string, *rate.Limiter]
	siteReportPrintRateLimiters      *lru.Cache[string, *rate.Limiter]
	sitePushRateLimiters             *lru.Cache[string, *rate.Limiter]

	SiteCreateLimiter   *rate.Limiter
	siteCreateIpLimiter *lru.Cache[string, *rate.Limiter]

	GlobalRateLimiter    *rate.Limiter // Limiter for all requestes everywhere
	globalIPRateLimiters *lru.Cache[string, *rate.Limiter]
	// ControlRateLimiter is the control site's own limit at the global rate, for those not signed in to it, such as
	// its login, so a site using up the global limit cannot stop a control admin signing in to fix it
	ControlRateLimiter *rate.Limiter
}

func newRateLimiters(config ServerConfig) *RateLimiters {
	maxSites := config.MaxSites
	maxIdentities := config.MaxIdentitiesPerSite
	wiggle := 2
	// We are using the hashicorp LRU to make sure that we dont run out of memory if all users with all identitys are hitting us always
	settings := rateDefaults(config)
	defaultRate := func(name string) Rate {
		return settings[slices.IndexFunc(settings, func(s RateSetting) bool { return s.Name == name })].Rate
	}
	return &RateLimiters{
		rates: &rateSettings{settings: settings},

		userLoginRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		ipLoginRateLimiters:   lruMust(maxSites * maxIdentities * wiggle * 2),

		siteUnauthRateLimiters: lruMust(maxSites * maxIdentities * wiggle * 2),

		identityGETRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		siteGETRateLimiters:     lruMust(maxSites * wiggle),

		identityPOSTRateLimiters: lruMust(maxSites * maxIdentities * wiggle),
		sitePOSTRateLimiters:     lruMust(maxSites * wiggle),

		siteUpdateConnectorsRateLimiters: lruMust(maxSites * wiggle),
		siteIdentityCreateRateLimiters:   lruMust(maxSites * wiggle),
		siteDownloadRateLimiters:         lruMust(maxSites * wiggle),
		siteReportPrintRateLimiters:      lruMust(maxSites * wiggle),
		sitePushRateLimiters:             lruMust(maxSites * wiggle),

		SiteCreateLimiter:   defaultRate("site_create").limiter(),
		siteCreateIpLimiter: lruMust(maxSites * wiggle),

		GlobalRateLimiter:    defaultRate("global").limiter(),
		globalIPRateLimiters: lruMust(maxSites * maxIdentities * wiggle * 2), // If a DDOS attack wiped out this cache with unique addresses the global rate limiter would still catch it
		ControlRateLimiter:   defaultRate("global").limiter(),
	}

}

func lruMust(maxEntries int) *lru.Cache[string, *rate.Limiter] {
	cache, err := lru.New[string, *rate.Limiter](maxEntries)
	if err != nil {
		panic(err)
	}

	return cache
}

func (r *RateLimiters) getOrCreateLimiter(limiters *lru.Cache[string, *rate.Limiter], key string, name string) *rate.Limiter {
	// Keys come from user input (e.g. a login name), so hash them to keep every entry a fixed size
	sum := sha256.Sum256([]byte(key))
	key = hex.EncodeToString(sum[:])
	limiter, exists := limiters.Get(key)
	if !exists {
		limiter = r.rate(name).limiter()
		limiters.Add(key, limiter)
	}
	return limiter
}

// ipKey is what an IP's limiters are keyed on. An IPv6 address is keyed on its /64, as a host is usually given a whole
// /64 to pick addresses from, so stepping through it cannot reset its limits.
func ipKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	if addr = addr.Unmap(); addr.Is4() {
		return addr.String()
	}
	prefix, _ := addr.Prefix(64)
	return prefix.String()
}

func (r *RateLimiters) GetGlobalIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.globalIPRateLimiters, ipKey(ipAddress), "global_ip")
}

func (r *RateLimiters) GetSiteCreateIPRateLimiter(ipAddress string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteCreateIpLimiter, ipKey(ipAddress), "site_create_ip")
}

// //
// Site Rate Limiters
// //

func (r *RateLimiters) IdentityGETLimiter(identityID int64) *rate.Limiter {
	return r.getOrCreateLimiter(r.identityGETRateLimiters, strconv.FormatInt(identityID, 10), "site_get")
}

func (r *RateLimiters) IdentityPOSTLimiter(identityID int64) *rate.Limiter {

	return r.getOrCreateLimiter(r.identityPOSTRateLimiters, strconv.FormatInt(identityID, 10), "site_post")
}

func (r *RateLimiters) SiteGETLimiter(subdomain string) *rate.Limiter {

	return r.getOrCreateLimiter(r.siteGETRateLimiters, subdomain, "site_get")
}

func (r *RateLimiters) SitePOSTLimiter(subdomain string) *rate.Limiter {

	return r.getOrCreateLimiter(r.sitePOSTRateLimiters, subdomain, "site_post")
}

func (r *RateLimiters) SiteUpdateConnectorsLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteUpdateConnectorsRateLimiters, subdomain, "site_update_connectors")
}

func (r *RateLimiters) SiteIdentityCreateLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteIdentityCreateRateLimiters, subdomain, "site_identity_create")
}

// SiteReportPrintLimiter is a site's budget of printable reports, each building four months of its reports
func (r *RateLimiters) SiteReportPrintLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteReportPrintRateLimiters, subdomain, "site_report_print")
}

func (r *RateLimiters) SiteDownloadLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteDownloadRateLimiters, subdomain, "site_download")
}

// SitePushLimiter is a site's budget of pushes, each a request to Apple or Firebase, so a site cannot use the server to
// send them without end
func (r *RateLimiters) SitePushLimiter(subdomain string) *rate.Limiter {
	return r.getOrCreateLimiter(r.sitePushRateLimiters, subdomain, "site_push")
}

// SiteUnauthLimiter is one IP's budget on a site before it authenticates: 401s and Socket.IO handshakes spend it, not the site's own limit
func (r *RateLimiters) SiteUnauthLimiter(subdomain string, ip string) *rate.Limiter {
	return r.getOrCreateLimiter(r.siteUnauthRateLimiters, subdomain+":"+ipKey(ip), "site_unauth")
}

func (r *RateLimiters) SiteLoginAllow(subdomain string, ip string) (bool, int) {
	userLimiter := r.getOrCreateLimiter(r.userLoginRateLimiters, subdomain, "login")
	ipLimiter := r.getOrCreateLimiter(r.ipLoginRateLimiters, ipKey(ip), "login")

	// Calculate tokens remaining
	userTokens := userLimiter.Tokens()
	ipTokens := ipLimiter.Tokens()

	// The effective remaining attempts is the minimum of user and IP limits
	remaining := int(math.Min(userTokens, ipTokens))

	return userLimiter.Allow() && ipLimiter.Allow(), remaining
}

// RateLimiterStats is how close one kind of rate limiter is to refusing. One is kept for each Per: the server, an IP, a
// site, an identity, a login name or an IP on a site; each refills at PerMinute up to Burst.
type RateLimiterStats struct {
	Name      string  `json:"name"`
	Per       string  `json:"per"`
	PerMinute float64 `json:"per_minute"`
	Burst     int     `json:"burst"`
	// Tracked is how many are kept, Limited how many would refuse a request now, and FullestPercent how much of its
	// burst the most used one has spent
	Tracked        int     `json:"tracked"`
	Limited        int     `json:"limited"`
	FullestPercent float64 `json:"fullest_percent"`
}

// Rate is a rate limit: a limiter refills PerMinute a minute, up to Burst
type Rate struct {
	PerMinute float64 `json:"per_minute"`
	Burst     int     `json:"burst"`
}

func (r Rate) limiter() *rate.Limiter { return rate.NewLimiter(rate.Limit(r.PerMinute/60.0), r.Burst) }

// Valid is a rate that lets requests through, under limits far above any server's
func (r Rate) Valid() bool {
	return r.PerMinute > 0 && r.PerMinute <= 1e9 && r.Burst >= 1 && r.Burst <= 1e9
}

// rateSettings are read under lock as one can change while requests are limited. It is kept by pointer, so a copy of
// the RateLimiters, as a test makes to swap a limiter, has the same.
type rateSettings struct {
	lock     sync.RWMutex
	settings []RateSetting
}

// RateSetting is a rate limit a control admin can change. Its Default is from the env variables named Env, with
// _LIMIT and _BURST, and Rate is what it is now: the Default unless Set by a control admin.
type RateSetting struct {
	Name    string `json:"name"`
	Env     string `json:"env"`
	Default Rate   `json:"default"`
	Rate    Rate   `json:"rate"`
	Set     bool   `json:"set"`
}

func rateDefaults(c ServerConfig) []RateSetting {
	settings := []RateSetting{
		{Name: "global", Env: "GLOBAL_RATE", Default: Rate{float64(c.GlobalRateLimit), c.GlobalRateBurst}},
		{Name: "global_ip", Env: "GLOBAL_IP_PER_MINUTE", Default: Rate{float64(c.GlobalIPPerMinuteLimit), c.GlobalIPPerMinuteBurst}},
		{Name: "site_create", Env: "SITE_CREATE_GLOBAL_PER_MINUTE", Default: Rate{float64(c.SiteCreateGlobalPerMinuteLimit), c.SiteCreateGlobalPerMinuteBurst}},
		{Name: "site_create_ip", Env: "SITE_CREATE_IP_PER_MINUTE", Default: Rate{c.SiteCreateIPPerMinuteLimit, c.SiteCreateIPPerMinuteBurst}},
		{Name: "login", Env: "SITE_LOGIN_PER_MINUTE", Default: Rate{float64(c.SiteLoginPerMinuteLimit), c.SiteLoginPerMinuteBurst}},
		{Name: "site_unauth", Env: "SITE_UNAUTH_PER_MINUTE", Default: Rate{float64(c.SiteUnauthPerMinuteLimit), c.SiteUnauthPerMinuteBurst}},
		{Name: "site_get", Env: "SITE_GET_PER_MINUTE", Default: Rate{float64(c.SiteGETPerMinuteLimit), c.SiteGETPerMinuteBurst}},
		{Name: "site_post", Env: "SITE_POST_PER_MINUTE", Default: Rate{float64(c.SitePOSTPerMinuteLimit), c.SitePOSTPerMinuteBurst}},
		{Name: "site_update_connectors", Env: "SITE_UPDATE_CONNECTORS_PER_MINUTE", Default: Rate{float64(c.SiteUpdateConnectorsPerMinuteLimit), c.SiteUpdateConnectorsPerMinuteBurst}},
		{Name: "site_identity_create", Env: "SITE_IDENTITY_CREATE_PER_MINUTE", Default: Rate{float64(c.SiteIdentityCreatePerMinuteLimit), c.SiteIdentityCreatePerMinuteBurst}},
		{Name: "site_download", Env: "SITE_DOWNLOAD_PER_MINUTE", Default: Rate{float64(c.SiteDownloadPerMinuteLimit), c.SiteDownloadPerMinuteBurst}},
		{Name: "site_push", Env: "SITE_PUSH_PER_MINUTE", Default: Rate{float64(c.SitePushPerMinuteLimit), c.SitePushPerMinuteBurst}},
		{Name: "site_report_print", Env: "SITE_REPORT_PRINT_PER_MINUTE", Default: Rate{float64(c.SiteReportPrintPerMinuteLimit), c.SiteReportPrintPerMinuteBurst}},
	}
	for i := range settings {
		settings[i].Rate = settings[i].Default
	}
	return settings
}

// rate is the named rate limit's rate now
func (r *RateLimiters) rate(name string) Rate {
	r.rates.lock.RLock()
	defer r.rates.lock.RUnlock()
	return r.rates.settings[slices.IndexFunc(r.rates.settings, func(s RateSetting) bool { return s.Name == name })].Rate
}

// Settings are each rate limit, its env default and its rate now
func (r *RateLimiters) Settings() []RateSetting {
	r.rates.lock.RLock()
	defer r.rates.lock.RUnlock()
	return slices.Clone(r.rates.settings)
}

// SetRates uses the rates a control admin set, by name, in place of the env's, for every limiter kept and made: one
// that is not valid, or is the env's, is left out, so the env's is used. It is the rates set, to keep.
func (r *RateLimiters) SetRates(rates map[string]Rate) map[string]Rate {
	r.rates.lock.Lock()
	set := map[string]Rate{}
	for i, setting := range r.rates.settings {
		given, ok := rates[setting.Name]
		r.rates.settings[i].Set = ok && given.Valid() && given != setting.Default
		r.rates.settings[i].Rate = setting.Default
		if r.rates.settings[i].Set {
			r.rates.settings[i].Rate, set[setting.Name] = given, given
		}
	}
	r.rates.lock.Unlock()
	for _, kind := range r.kinds() {
		current := r.rate(kind.setting)
		for _, limiter := range kind.limiters {
			limiter.SetLimit(rate.Limit(current.PerMinute / 60.0))
			limiter.SetBurst(current.Burst)
		}
	}
	return set
}

// limiterKind is one kind of rate limiter: what it is named on the rate limits page, what one is kept for, the rate
// limit setting it is limited by, and those kept
type limiterKind struct {
	name, per, setting string
	limiters           []*rate.Limiter
}

func (r *RateLimiters) kinds() []limiterKind {
	return []limiterKind{
		{"global", "server", "global", []*rate.Limiter{r.GlobalRateLimiter}},
		{"control", "server", "global", []*rate.Limiter{r.ControlRateLimiter}},
		{"global_ip", "ip", "global_ip", r.globalIPRateLimiters.Values()},
		{"site_create", "server", "site_create", []*rate.Limiter{r.SiteCreateLimiter}},
		{"site_create_ip", "ip", "site_create_ip", r.siteCreateIpLimiter.Values()},
		{"login_name", "login", "login", r.userLoginRateLimiters.Values()},
		{"login_ip", "ip", "login", r.ipLoginRateLimiters.Values()},
		{"site_unauth", "site_ip", "site_unauth", r.siteUnauthRateLimiters.Values()},
		{"site_get", "site", "site_get", r.siteGETRateLimiters.Values()},
		{"site_post", "site", "site_post", r.sitePOSTRateLimiters.Values()},
		{"identity_get", "identity", "site_get", r.identityGETRateLimiters.Values()},
		{"identity_post", "identity", "site_post", r.identityPOSTRateLimiters.Values()},
		{"site_update_connectors", "site", "site_update_connectors", r.siteUpdateConnectorsRateLimiters.Values()},
		{"site_identity_create", "site", "site_identity_create", r.siteIdentityCreateRateLimiters.Values()},
		{"site_download", "site", "site_download", r.siteDownloadRateLimiters.Values()},
		{"site_push", "site", "site_push", r.sitePushRateLimiters.Values()},
		{"site_report_print", "site", "site_report_print", r.siteReportPrintRateLimiters.Values()},
	}
}

// Stats are how close each kind of rate limiter is to refusing at now, read without spending any
func (r *RateLimiters) Stats(now time.Time) []RateLimiterStats {
	kinds := r.kinds()
	stats := make([]RateLimiterStats, 0, len(kinds))
	for _, kind := range kinds {
		current := r.rate(kind.setting)
		s := RateLimiterStats{Name: kind.name, Per: kind.per, PerMinute: current.PerMinute, Burst: current.Burst, Tracked: len(kind.limiters)}
		for _, limiter := range kind.limiters {
			tokens := limiter.TokensAt(now)
			if tokens < 1 {
				s.Limited++
			}
			if current.Burst > 0 {
				s.FullestPercent = max(s.FullestPercent, min(100, 100*(1-tokens/float64(current.Burst))))
			}
		}
		stats = append(stats, s)
	}
	return stats
}
