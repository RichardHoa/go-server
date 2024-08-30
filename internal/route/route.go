package route

import (
	"net/http"
	"path/filepath"
	"github.com/RichardHoa/go-server/internal/config"
	"github.com/RichardHoa/go-server/internal/handlers"
)

// configureRoutes sets up the routes and middleware for the server.
func ConfigureRoutes(mux *http.ServeMux, apiCfg *config.ApiConfig) {
	// Create the file server handler
	fileServer := http.FileServer(http.Dir(filepath.Join(".")))

	// Wrap the file server handler with MiddlewareMetricsInc
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Add the readiness endpoint at /healthz
	mux.HandleFunc("GET /api/healthz", handlers.HandlerReadiness)

	// Register the metrics handler at /metrics
	mux.HandleFunc("GET /api/metrics", apiCfg.HandlerMetrics)

	// Register the reset handler at /reset
	mux.HandleFunc("GET /api/reset", apiCfg.HandlerReset)
}
