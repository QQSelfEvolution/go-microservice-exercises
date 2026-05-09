package main

import (
	"log"
	"net/http"
	"time"

	"github.com/QQSelfEvolution/go-microservice-exercises/xiaohou/day2/gateway"
)

func main() {
	// Initialize gateway
	gw := gateway.NewGateway()

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"api-gateway"}`))
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		stats := gw.GetStats()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(stats))
	})

	// API proxy (catch all /api/* routes)
	mux.HandleFunc("/api/", gw.HandleProxy())

	// Start gateway
	addr := ":8080"
	log.Printf("API Gateway starting on %s", addr)
	log.Printf("Configured backends: %d", len(gw.GetBackends()))

	// Start health checker in background
	go gw.StartHealthChecker(10 * time.Second)

	server := &http.Server{
		Addr:         addr,
		Handler:      gw.LoggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
