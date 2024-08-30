package handlers

import (
	// "fmt"
	"net/http"
	"encoding/json"
	// "github.com/RichardHoa/go-server/internal/config"
)

type Chirp struct {
	Body string `json:"body"`
}

// HandlerReadiness handles the /healthz endpoint
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func HandlerValidateChirp(w http.ResponseWriter, r *http.Request) {
// Check that the request method is POST
if r.Method != http.MethodPost {
	http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	return
}

// Parse the JSON body
var chirp Chirp
decoder := json.NewDecoder(r.Body)
if err := decoder.Decode(&chirp); err != nil {
	http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
	return
}

// Validate the chirp body
if len(chirp.Body) == 0 {
	http.Error(w, `{"error": "Chirp is empty"}`, http.StatusBadRequest)
	return
}
if len(chirp.Body) > 140 {
	http.Error(w, `{"error": "Chirp is too long"}`, http.StatusBadRequest)
	return
}

// fmt.Printf("Chirp length: %d\n", len(chirp.Body))

// If validation passes, respond with a 200 status code
response := map[string]bool{"valid": true}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
if err := json.NewEncoder(w).Encode(response); err != nil {
	http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
}


}