// Package metrics is the server's metrics, in Prometheus's format, served on METRICS_PORT for Fly.io's Prometheus or
// any other. None is labelled with a site, an identity or an IP address: thousands of sites would each be a series,
// and their IDs would leave the server.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()
	metric   = promauto.With(registry)

	requests = metric.NewCounterVec(prometheus.CounterOpts{Name: "nighttroop_http_requests_total",
		Help: "Requests answered, by the router they were for and their status"}, []string{"router", "code"})
	requestSeconds = metric.NewHistogramVec(prometheus.HistogramOpts{Name: "nighttroop_http_request_duration_seconds",
		Help: "How long requests took, by router, but for long lived ones (Socket.IO, server-sent events)"}, []string{"router"})
	refused = metric.NewCounterVec(prometheus.CounterOpts{Name: "nighttroop_requests_refused_total",
		Help: "Requests refused before they were signed in, as for a rate limit, by why and status"}, []string{"reason", "code"})
	rateLimited = metric.NewCounter(prometheus.CounterOpts{Name: "nighttroop_rate_limited_total",
		Help: "Requests or writes a rate limit refused"})
	writesRefused = metric.NewCounterVec(prometheus.CounterOpts{Name: "nighttroop_site_writes_refused_total",
		Help: "Site writes refused, as the site is far over its storage or the server's disk is nearly full"}, []string{"why"})
	jobRuns = metric.NewCounterVec(prometheus.CounterOpts{Name: "nighttroop_job_runs_total",
		Help: "Background job runs, by job and whether they failed"}, []string{"job", "result"})
	jobSeconds = metric.NewHistogramVec(prometheus.HistogramOpts{Name: "nighttroop_job_duration_seconds",
		Help: "How long background job runs took, by job", Buckets: prometheus.ExponentialBuckets(0.01, 4, 10)}, []string{"job"})
	sitesOpen = metric.NewGauge(prometheus.GaugeOpts{Name: "nighttroop_sites_open",
		Help: "Sites whose DB is open, at the last resource check"})
	sitesUnavailable = metric.NewGaugeVec(prometheus.GaugeOpts{Name: "nighttroop_sites_unavailable",
		Help: "Sites whose DB is not open, by why, at the last resource check"}, []string{"reason"})
	diskFree = metric.NewGauge(prometheus.GaugeOpts{Name: "nighttroop_disk_free_bytes",
		Help: "Free space on the DBs' disk, at the last resource check"})
	siteDBs = metric.NewGauge(prometheus.GaugeOpts{Name: "nighttroop_site_dbs_bytes",
		Help: "The open site DBs' size in all, at the last measure of the sites"})
)

func init() {
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// Handler serves the metrics
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// Request counts a request for router answered with code after took, and times it unless it is long lived
func Request(router string, code int, took time.Duration, longLived bool) {
	requests.WithLabelValues(router, strconv.Itoa(code)).Inc()
	if !longLived {
		requestSeconds.WithLabelValues(router).Observe(took.Seconds())
	}
}

// Refused counts a request refused for reason, a fixed message, with code
func Refused(reason string, code int) { refused.WithLabelValues(reason, strconv.Itoa(code)).Inc() }

// RateLimited counts a request or write a rate limit refused
func RateLimited() { rateLimited.Inc() }

// WriteRefused counts a site write refused for why
func WriteRefused(why string) { writesRefused.WithLabelValues(why).Inc() }

// JobRun counts a run of job that started at start and failed with err, or not, and times it
func JobRun(job string, start time.Time, err error) {
	result := "ok"
	if err != nil {
		result = "failed"
	}
	jobRuns.WithLabelValues(job, result).Inc()
	jobSeconds.WithLabelValues(job).Observe(time.Since(start).Seconds())
}

// Sites keeps how many sites are open, and how many are unavailable by why
func Sites(open int, unavailable map[string]string) {
	sitesOpen.Set(float64(open))
	sitesUnavailable.Reset()
	for _, reason := range unavailable {
		sitesUnavailable.WithLabelValues(reason).Inc()
	}
}

// DiskFree keeps the DBs' disk's free bytes
func DiskFree(bytes float64) { diskFree.Set(bytes) }

// SiteDBs keeps the open site DBs' size in all, in bytes
func SiteDBs(bytes float64) { siteDBs.Set(bytes) }
