package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg             *prometheus.Registry
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	LinksCreated    prometheus.Counter
	Redirects       prometheus.Counter
	CacheHits       prometheus.Counter
	CacheMisses     prometheus.Counter
}

func New(namespace string) *Metrics {
	m := &Metrics{
		reg: prometheus.NewRegistry(),
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests.",
		}, []string{"method", "path", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),
		LinksCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "links_created_total",
			Help:      "Total created short links.",
		}),
		Redirects: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "redirects_total",
			Help:      "Total redirects.",
		}),
		CacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_hits_total",
			Help:      "Total cache hits.",
		}),
		CacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_misses_total",
			Help:      "Total cache misses.",
		}),
	}
	m.reg.MustRegister(m.RequestsTotal, m.RequestDuration, m.LinksCreated, m.Redirects, m.CacheHits, m.CacheMisses)
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}

func (m *Metrics) Observe(method, path string, status int, started time.Time) {
	m.RequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	m.RequestDuration.WithLabelValues(method, path).Observe(time.Since(started).Seconds())
}
