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

	mux.HandleFunc("POST /api/chirps", apiCfg.HandlerAddChirps)

	mux.HandleFunc("GET /api/chirps", handlers.HandlerGetChirps)

	mux.HandleFunc("GET /api/chirps/", handlers.HandlerGetChirpsID)

	mux.HandleFunc("POST /api/users", handlers.HandlerAddUser)

	mux.HandleFunc("POST /api/login", apiCfg.HandlerAuthenticateUser)

	mux.HandleFunc("PUT /api/users", apiCfg.HandlerPutUser)

	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefreshToken)

	mux.HandleFunc("POST /api/revoke", apiCfg.HandlerRevokeToken)

	mux.HandleFunc("DELETE /api/chirps/", apiCfg.HandlerDeleteChirps)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandlerPolkaWebhooks)
}
