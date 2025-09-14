package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockShardRouter para testes do main
type MockShardRouter struct {
	shardingKey   string
	expectedShard string
	initCalled    bool
	shardsAdded   []string
}

func (m *MockShardRouter) InitHashRing(size int) {
	m.initCalled = true
}

func (m *MockShardRouter) AddShard(shardHost string) {
	m.shardsAdded = append(m.shardsAdded, shardHost)
}

func (m *MockShardRouter) GetShardingKey(r *http.Request) string {
	return r.Header.Get(m.shardingKey)
}

func (m *MockShardRouter) GetShardHost(key string) string {
	return m.expectedShard
}

// MockMetricsRecorder para testes
type MockMetricsRecorder struct {
	requests  map[string]int
	responses map[string]map[int]int
}

func NewMockMetricsRecorder() *MockMetricsRecorder {
	return &MockMetricsRecorder{
		requests:  make(map[string]int),
		responses: make(map[string]map[int]int),
	}
}

func (m *MockMetricsRecorder) RecordRequest(shard string) {
	m.requests[shard]++
}

func (m *MockMetricsRecorder) RecordResponse(shard string, statusCode int) {
	if m.responses[shard] == nil {
		m.responses[shard] = make(map[int]int)
	}
	m.responses[shard][statusCode]++
}

func TestNewPrometheusMetricsRecorder(t *testing.T) {
	recorder := NewPrometheusMetricsRecorder()

	if recorder == nil {
		t.Fatal("Expected non-nil metrics recorder")
	}
}

func TestPrometheusMetricsRecorder_RecordRequest(t *testing.T) {
	recorder := NewPrometheusMetricsRecorder()

	// Não vai causar panic
	recorder.RecordRequest("http://shard01:80")
}

func TestPrometheusMetricsRecorder_RecordResponse(t *testing.T) {
	recorder := NewPrometheusMetricsRecorder()

	// Não vai causar panic
	recorder.RecordResponse("http://shard01:80", 200)
}

func TestNewProxyHandler(t *testing.T) {
	mockRouter := &MockShardRouter{}
	mockRecorder := NewMockMetricsRecorder()

	handler := NewProxyHandler(mockRouter, mockRecorder)

	if handler == nil {
		t.Fatal("Expected non-nil proxy handler")
	}

	if handler.router != mockRouter {
		t.Error("Expected router to be set correctly")
	}

	if handler.metricsRecorder != mockRecorder {
		t.Error("Expected metrics recorder to be set correctly")
	}
}

func TestProxyHandler_ServeHTTP(t *testing.T) {
	// Setup mock backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backendServer.Close()

	// Setup mock router
	mockRouter := &MockShardRouter{
		shardingKey:   "user_id",
		expectedShard: backendServer.URL,
	}

	mockRecorder := NewMockMetricsRecorder()
	handler := NewProxyHandler(mockRouter, mockRecorder)

	// Create test request
	req := httptest.NewRequest("GET", "/test-path", nil)
	req.Header.Set("user_id", "test-user")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "backend response" {
		t.Errorf("Expected 'backend response', got '%s'", rr.Body.String())
	}

	// Verify metrics were recorded
	if mockRecorder.requests[backendServer.URL] != 1 {
		t.Errorf("Expected 1 request recorded, got %d", mockRecorder.requests[backendServer.URL])
	}

	if mockRecorder.responses[backendServer.URL][200] != 1 {
		t.Errorf("Expected 1 response recorded, got %d", mockRecorder.responses[backendServer.URL][200])
	}
}

func TestProxyHandler_ServeHTTP_InvalidURL(t *testing.T) {
	// Setup mock router with invalid URL that will cause url.Parse to fail
	mockRouter := &MockShardRouter{
		shardingKey:   "user_id",
		expectedShard: "ht!tp://invalid-url", // Invalid URL with special character
	}

	mockRecorder := NewMockMetricsRecorder()
	handler := NewProxyHandler(mockRouter, mockRecorder)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("user_id", "test-user")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rr, req)

	// Verify error response - URL parsing fails, so we get BadRequest
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Invalid target URL") {
		t.Errorf("Expected error message about invalid URL, got '%s'", rr.Body.String())
	}
}

func TestProxyHandler_ServeHTTP_BackendError(t *testing.T) {
	// Setup mock router with non-existent backend
	mockRouter := &MockShardRouter{
		shardingKey:   "user_id",
		expectedShard: "http://non-existent-backend:12345",
	}

	mockRecorder := NewMockMetricsRecorder()
	handler := NewProxyHandler(mockRouter, mockRecorder)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("user_id", "test-user")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rr, req)

	// Verify error response
	if rr.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d", rr.Code)
	}

	// Verify request was still recorded
	if mockRecorder.requests["http://non-existent-backend:12345"] != 1 {
		t.Errorf("Expected 1 request recorded even on error, got %d",
			mockRecorder.requests["http://non-existent-backend:12345"])
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	HealthCheckHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestProxyHandler_HeaderPropagation(t *testing.T) {
	// Setup mock backend server that echoes headers
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo custom header back
		if customHeader := r.Header.Get("X-Custom-Header"); customHeader != "" {
			w.Header().Set("X-Echo-Header", customHeader)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backendServer.Close()

	mockRouter := &MockShardRouter{
		shardingKey:   "user_id",
		expectedShard: backendServer.URL,
	}

	mockRecorder := NewMockMetricsRecorder()
	handler := NewProxyHandler(mockRouter, mockRecorder)

	// Create test request with custom header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("user_id", "test-user")
	req.Header.Set("X-Custom-Header", "custom-value")

	rr := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify header was propagated and echoed back
	if echoHeader := rr.Header().Get("X-Echo-Header"); echoHeader != "custom-value" {
		t.Errorf("Expected echoed header 'custom-value', got '%s'", echoHeader)
	}
}

func TestProxyHandler_HTTPMethods(t *testing.T) {
	// Test different HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Setup mock backend server
			backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("Expected method %s, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer backendServer.Close()

			mockRouter := &MockShardRouter{
				shardingKey:   "user_id",
				expectedShard: backendServer.URL,
			}

			mockRecorder := NewMockMetricsRecorder()
			handler := NewProxyHandler(mockRouter, mockRecorder)

			// Create test request
			var req *http.Request
			if method == "POST" || method == "PUT" || method == "PATCH" {
				req = httptest.NewRequest(method, "/test", strings.NewReader("test body"))
			} else {
				req = httptest.NewRequest(method, "/test", nil)
			}
			req.Header.Set("user_id", "test-user")

			rr := httptest.NewRecorder()

			// Execute
			handler.ServeHTTP(rr, req)

			// Verify response
			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200 for method %s, got %d", method, rr.Code)
			}
		})
	}
}
