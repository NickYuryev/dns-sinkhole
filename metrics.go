package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const metricsNamespace = "sinkhole"

var (
	queryLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "query_duration_seconds",
		Help:      "DNS query handling latency in seconds.",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 14),
	}, []string{"status", "qtype"})

	blocklistLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "blocklist_lookup_duration_seconds",
		Help:      "Time spent checking a domain against the blocklist.",
		Buckets:   prometheus.ExponentialBuckets(1e-7, 2, 20),
	})

	cacheGetLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "cache_get_duration_seconds",
		Help:      "Time spent on cache Get, labeled by hit/miss.",
		Buckets:   prometheus.ExponentialBuckets(1e-7, 2, 20),
	}, []string{"result"})

	cacheSetLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "cache_set_duration_seconds",
		Help:      "Time spent writing to the cache.",
		Buckets:   prometheus.ExponentialBuckets(1e-7, 2, 20),
	})
)

func startMetricsServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)
	log.Printf("metrics listening on %s", addr)
	go func() {
		srv := &http.Server{Addr: addr, Handler: mux}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server: %v", err)
		}
	}()
}
