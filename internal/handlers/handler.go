package handlers

import (
	"fmt"
	"net/http"
	"github.com/RichardHoa/go-server/internal/config"
)

// HandlerReadiness handles the /healthz endpoint
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// HandlerMetrics returns the number of requests
func HandlerMetrics(apiCfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiCfg.Mu.Lock()
		hits := apiCfg.FileserverHits
		apiCfg.Mu.Unlock()
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %d", hits)))
	}
}

// HandlerReset resets the hit counter
func HandlerReset(apiCfg *config.ApiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiCfg.Mu.Lock()
		apiCfg.FileserverHits = 0
		apiCfg.Mu.Unlock()
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hits reset to 0"))
	}
}
