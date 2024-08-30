package main

import (
	"log"
	"net/http"
	"path/filepath"
	"github.com/RichardHoa/go-server/internal/config"
	"github.com/RichardHoa/go-server/internal/handlers"
	// "github.com/RichardHoa/go-server/internal/middleware"
)

func main() {
	const port = "8080"

	// Initialize apiConfig
	apiCfg := &config.ApiConfig{}

	// Create a new ServeMux
	mux := http.NewServeMux()
	
	// Create the file server handler
	fileServer := http.FileServer(http.Dir(filepath.Join(".")))

	// Wrap the file server handler with middlewareMetricsInc
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Add the readiness endpoint at /healthz
	mux.HandleFunc("/healthz", handlers.HandlerReadiness)

	// Register the metrics handler at /metrics
	mux.HandleFunc("/metrics", apiCfg.HandlerMetrics)

	// Register the reset handler at /reset
	mux.HandleFunc("/reset", apiCfg.HandlerReset)

	// Create and start the server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", ".", port)
	log.Fatal(srv.ListenAndServe())
}
