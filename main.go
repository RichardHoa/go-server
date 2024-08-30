package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Define the apiConfig struct to hold stateful, in-memory data
type apiConfig struct {
	fileserverHits int
	mu             sync.Mutex // Mutex to ensure safe concurrent access to fileserverHits
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Initialize apiConfig
	apiCfg := &apiConfig{}

	// Create a new ServeMux
	mux := http.NewServeMux()
	
	// Create the file server handler
	fileServer := http.FileServer(http.Dir(filepathRoot))

	// Wrap the file server handler with middlewareMetricsInc
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Add the readiness endpoint at /healthz
	mux.HandleFunc("GET /healthz", handlerReadiness)

	// Register the metrics handler at /metrics
	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)

	// Register the reset handler at /reset
	mux.HandleFunc("/reset", apiCfg.handlerReset)

	// Create and start the server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

// Readiness handler for /healthz
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Middleware method to increment the fileserverHits counter
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		cfg.fileserverHits++
		cfg.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// Handler method for /metrics to show the number of requests
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	hits := cfg.fileserverHits
	cfg.mu.Unlock()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", hits)))
}

// Handler method for /reset to reset the hit counter
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	cfg.fileserverHits = 0
	cfg.mu.Unlock()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
