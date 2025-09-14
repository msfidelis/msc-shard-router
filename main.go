package main

import (
	"app/pkg/interfaces"
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

// ProxyServer encapsula as dependências e configurações do servidor
type ProxyServer struct {
	router          interfaces.ShardRouter
	metricsRecorder interfaces.MetricsRecorder
	port            string
}

// PrometheusMetricsRecorder implementa a interface MetricsRecorder
type PrometheusMetricsRecorder struct {
	requestsCounter prometheus.CounterVec
	responseCounter prometheus.CounterVec
}

// Garantir que PrometheusMetricsRecorder implementa a interface
var _ interfaces.MetricsRecorder = (*PrometheusMetricsRecorder)(nil)

func (pm *PrometheusMetricsRecorder) RecordRequest(shard string) {
	pm.requestsCounter.WithLabelValues(shard).Inc()
}

func (pm *PrometheusMetricsRecorder) RecordResponse(shard string, statusCode int) {
	pm.responseCounter.WithLabelValues(shard, strconv.Itoa(statusCode)).Inc()
}

// NewPrometheusMetricsRecorder cria uma nova instância do recorder de métricas
func NewPrometheusMetricsRecorder() *PrometheusMetricsRecorder {
	requestsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shard_router_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"shard"},
	)
	responseCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shard_router_responses_total",
			Help: "Total number of HTTP responses",
		},
		[]string{"shard", "status"},
	)

	return &PrometheusMetricsRecorder{
		requestsCounter: *requestsCounter,
		responseCounter: *responseCounter,
	}
}

// NewProxyServer cria uma nova instância do servidor proxy
func NewProxyServer(port string) *ProxyServer {
	shardingKey := os.Getenv("SHARDING_KEY")
	if shardingKey == "" {
		log.Fatal("SHARDING_KEY environment variable is required")
	}

	router := sharding.NewShardRouter(shardingKey)
	metricsRecorder := NewPrometheusMetricsRecorder()

	return &ProxyServer{
		router:          router,
		metricsRecorder: metricsRecorder,
		port:            port,
	}
}

// ProxyHandler implementa a interface ProxyHandler
type ProxyHandler struct {
	router          interfaces.ShardRouter
	metricsRecorder interfaces.MetricsRecorder
}

// Garantir que ProxyHandler implementa a interface
var _ interfaces.ProxyHandler = (*ProxyHandler)(nil)

// ServeHTTP implementa o handler HTTP para o proxy
func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shardKey := ph.router.GetShardingKey(r)
	shardURL := ph.router.GetShardHost(shardKey)

	targetURL, err := url.Parse(shardURL + r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	ph.metricsRecorder.RecordRequest(shardURL)

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

	ph.metricsRecorder.RecordResponse(shardURL, resp.StatusCode)

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// NewProxyHandler cria um novo handler de proxy
func NewProxyHandler(router interfaces.ShardRouter, metricsRecorder interfaces.MetricsRecorder) *ProxyHandler {
	return &ProxyHandler{
		router:          router,
		metricsRecorder: metricsRecorder,
	}
}

// HealthCheckHandler implementa o health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// SetupRouter configura e inicializa o roteador de shards
func (ps *ProxyServer) SetupRouter() error {
	err := setup.InitWithRouter(ps.router)
	if err != nil {
		return err
	}
	return nil
}

// Start inicia o servidor HTTP
func (ps *ProxyServer) Start() error {
	// Setup do roteador
	err := ps.SetupRouter()
	if err != nil {
		return err
	}

	// Prometheus
	reg := prometheus.NewRegistry()

	// Type assertion para acessar os counters do Prometheus
	prometheusRecorder := ps.metricsRecorder.(*PrometheusMetricsRecorder)

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		&prometheusRecorder.requestsCounter,
		&prometheusRecorder.responseCounter,
	)

	// Setup dos handlers
	proxyHandler := NewProxyHandler(ps.router, ps.metricsRecorder)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/healthz", HealthCheckHandler)
	mux.Handle("/", proxyHandler)

	log.Printf("HTTP Proxy running on port %s", ps.port)
	return http.ListenAndServe(":"+ps.port, mux)
}

func main() {
	port := os.Getenv("ROUTER_PORT")
	if port == "" {
		port = "8080"
	}

	server := NewProxyServer(port)
	log.Fatal(server.Start())
}
