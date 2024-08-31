package route

import (
	"net/http"
	"path/filepath"
	"github.com/RichardHoa/go-server/internal/config"
	"github.com/RichardHoa/go-server/internal/handlers"
)

func ConfigureRoutes(mux *http.ServeMux, apiCfg *config.ApiConfig) {
	fileServer := http.FileServer(http.Dir(filepath.Join(".")))

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	mux.HandleFunc("GET /api/healthz", handlers.HandlerReadiness)

	mux.HandleFunc("GET /api/metrics", apiCfg.HandlerMetrics)

	mux.HandleFunc("GET /api/reset", apiCfg.HandlerReset)

	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetricsHTML)	

	mux.HandleFunc("POST /api/chirps", handlers.HandlerAddChirps)

	mux.HandleFunc("GET /api/chirps", handlers.HandlerGetChirps)
}
