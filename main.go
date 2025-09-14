package main

import (
	"app/pkg/setup"
	"app/pkg/sharding"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shard_router_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"shard"},
	)
	responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shard_router_responses_total",
			Help: "Total number of HTTP responses",
		},
		[]string{"shard", "status"},
	)
)

func main() {

	setup.Init()

	proxyHandler := func(w http.ResponseWriter, r *http.Request) {

		shardKey := sharding.GetShardingKey(r)
		shardURL := sharding.GetShardHost(shardKey)
		targetURL, err := url.Parse(shardURL + r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusBadRequest)
			return
		}
		requestsCounter.WithLabelValues(shardURL).Inc()
		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
		proxyReq.Header = r.Header

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		responseCounter.WithLabelValues(shardURL, strconv.Itoa(resp.StatusCode)).Inc()

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}

	healthCheckHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// Prometheus
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestsCounter,
		responseCounter,
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/healthz", healthCheckHandler)
	mux.HandleFunc("/", proxyHandler)

	port := os.Getenv("ROUTER_PORT")
	log.Printf("HTTP Proxy running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
