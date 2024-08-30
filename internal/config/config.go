package config

import (
	"fmt"
	"net/http"
	"sync"
)

// ApiConfig holds stateful, in-memory data
type ApiConfig struct {
	FileserverHits int
	Mu             sync.Mutex // Mutex to ensure safe concurrent access to FileserverHits
}

func (apiCfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.Mu.Lock()
		apiCfg.FileserverHits++
		apiCfg.Mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// HandlerMetrics returns the number of requests
func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.Mu.Lock()
	hits := cfg.FileserverHits
	cfg.Mu.Unlock()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", hits)))
}

func (cfg *ApiConfig) HandlerMetricsHTML(w http.ResponseWriter, r *http.Request) {
	cfg.Mu.Lock()
	hits := cfg.FileserverHits
	cfg.Mu.Unlock()
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	html := `<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>
</html>`

	fmt.Fprintf(w, html, hits)
}

// HandlerReset resets the hit counter
func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.Mu.Lock()
	cfg.FileserverHits = 0
	cfg.Mu.Unlock()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
