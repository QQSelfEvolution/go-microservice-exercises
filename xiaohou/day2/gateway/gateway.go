package gateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Backend represents a backend service
type Backend struct {
	URL           *url.URL
	Name          string
	Healthy       bool
	FailCount     int
	SuccessCount  int64
	TotalRequests int64
	LastCheck     time.Time
	mu            sync.RWMutex
}

// Gateway handles request proxying and load balancing
type Gateway struct {
	backends    []*Backend
	currentIdx  uint32
	mu          sync.RWMutex
	totalReq    int64
	activeReq   int64
	muStats     sync.Mutex
}

// NewGateway creates a new Gateway instance
func NewGateway() *Gateway {
	backends := []Backend{
		{URL: &url.URL{Host: "localhost:3001", Scheme: "http"}, Name: "service-1", Healthy: true},
		{URL: &url.URL{Host: "localhost:3002", Scheme: "http"}, Name: "service-2", Healthy: true},
		{URL: &url.URL{Host: "localhost:3003", Scheme: "http"}, Name: "service-3", Healthy: true},
	}

	gw := &Gateway{
		backends: make([]*Backend, len(backends)),
	}

	for i := range backends {
		gw.backends[i] = &backends[i]
	}

	return gw
}

// GetBackends returns a copy of current backends
func (g *Gateway) GetBackends() []*Backend {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*Backend, len(g.backends))
	copy(result, g.backends)
	return result
}

// GetNextBackend returns the next healthy backend using round-robin
func (g *Gateway) GetNextBackend() (*Backend, error) {
	g.mu.RLock()
	backends := g.backends
	g.mu.RUnlock()

	// Find healthy backends
	var healthy []*Backend
	for _, b := range backends {
		b.mu.RLock()
		if b.Healthy {
			healthy = append(healthy, b)
		}
		b.mu.RUnlock()
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy backends available")
	}

	// Round-robin selection
	idx := atomic.AddUint32(&g.currentIdx, 1)
	selected := healthy[(idx-1)%uint32(len(healthy))]

	return selected, nil
}

// HandleProxy handles incoming requests and proxies them to backends
func (g *Gateway) HandleProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&g.totalReq, 1)
		atomic.AddInt64(&g.activeReq, 1)
		defer atomic.AddInt64(&g.activeReq, -1)

		// Get next backend
		backend, err := g.GetNextBackend()
		if err != nil {
			log.Printf("Error getting backend: %v", err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		// Update original URL
		r.URL.Host = backend.URL.Host
		r.URL.Scheme = backend.URL.Scheme
		r.Host = backend.URL.Host

		// Add headers
		r.Header.Set("X-Forwarded-For", r.RemoteAddr)
		r.Header.Set("X-Forwarded-Proto", "http")
		r.Header.Set("X-Gateway-Name", backend.Name)
		r.Header.Set("X-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

		// Track requests
		atomic.AddInt64(&backend.TotalRequests, 1)

		// Director function for logging
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			log.Printf("[PROXY] %s %s -> %s", req.Method, req.URL.Path, backend.URL.String())
		}

		// Error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			log.Printf("[ERROR] Proxy error for %s: %v", backend.Name, err)
			backend.MarkUnhealthy()
			http.Error(w, "Backend error", http.StatusBadGateway)
		}

		// Response writer wrapper
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		proxy.ServeHTTP(rw, r)

		if rw.statusCode >= 200 && rw.statusCode < 400 {
			backend.MarkHealthy()
		} else {
			backend.MarkUnhealthy()
		}
	}
}

// MarkHealthy marks a backend as healthy
func (b *Backend) MarkHealthy() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.FailCount = 0
	b.Healthy = true
	b.LastCheck = time.Now()
}

// MarkUnhealthy marks a backend as unhealthy
func (b *Backend) MarkUnhealthy() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.FailCount++
	if b.FailCount >= 3 {
		b.Healthy = false
		log.Printf("[HEALTH] Backend %s marked unhealthy (failures: %d)", b.Name, b.FailCount)
	}
	b.LastCheck = time.Now()
}

// StartHealthChecker starts periodic health checks for all backends
func (g *Gateway) StartHealthChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		g.checkAllBackends()
	}
}

// checkAllBackends checks health of all backends
func (g *Gateway) checkAllBackends() {
	g.mu.RLock()
	backends := g.backends
	g.mu.RUnlock()

	for _, backend := range backends {
		go g.checkBackend(backend)
	}
}

// checkBackend checks health of a single backend
func (g *Gateway) checkBackend(backend *Backend) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("%s/health", backend.URL.String())
	resp, err := client.Get(url)

	backend.mu.Lock()
	defer backend.mu.Unlock()

	if err != nil || resp.StatusCode != http.StatusOK {
		backend.FailCount++
		if backend.FailCount >= 3 && backend.Healthy {
			backend.Healthy = false
			log.Printf("[HEALTH] Backend %s is DOWN", backend.Name)
		}
	} else {
		resp.Body.Close()
		if !backend.Healthy {
			backend.Healthy = true
			backend.FailCount = 0
			log.Printf("[HEALTH] Backend %s is UP", backend.Name)
		}
	}
	backend.LastCheck = time.Now()
}

// LoggingMiddleware logs all incoming requests
func (g *Gateway) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		log.Printf("[REQUEST] %s %s - %d (%v)",
			r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}

// GetStats returns gateway statistics as JSON
func (g *Gateway) GetStats() string {
	g.mu.RLock()
	backends := g.backends
	g.mu.RUnlock()

	type BackendStats struct {
		Name          string  `json:"name"`
		URL           string  `json:"url"`
		Healthy       bool    `json:"healthy"`
		TotalRequests int64   `json:"total_requests"`
		LastCheck     string  `json:"last_check"`
	}

	type Stats struct {
		TotalRequests int64           `json:"total_requests"`
		ActiveRequests int64         `json:"active_requests"`
		BackendCount   int           `json:"backend_count"`
		HealthyCount   int           `json:"healthy_count"`
		Backends       []BackendStats `json:"backends"`
	}

	var healthyCount int
	backendStats := make([]BackendStats, len(backends))

	for i, b := range backends {
		b.mu.RLock()
		healthy := b.Healthy
		if healthy {
			healthyCount++
		}
		backendStats[i] = BackendStats{
			Name:          b.Name,
			URL:           b.URL.String(),
			Healthy:       healthy,
			TotalRequests: b.TotalRequests,
			LastCheck:     b.LastCheck.Format(time.RFC3339),
		}
		b.mu.RUnlock()
	}

	stats := Stats{
		TotalRequests:  atomic.LoadInt64(&g.totalReq),
		ActiveRequests: atomic.LoadInt64(&g.activeReq),
		BackendCount:   len(backends),
		HealthyCount:   healthyCount,
		Backends:       backendStats,
	}

	jsonData, _ := json.MarshalIndent(stats, "", "  ")
	return string(jsonData)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
