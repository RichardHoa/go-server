package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"fmt"
	"os"
)

type Chirp struct {
	ID int
	Body string `json:"body"`
}

var (
	currentID = 1
	mutex     sync.Mutex
)

// HandlerReadiness handles the /healthz endpoint
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func HandlerAddChirps(w http.ResponseWriter, r *http.Request) {
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

	// Replace sensitive words in the chirp body
	chirp.Body = ReplaceSensitiveWords(chirp.Body)

	// Synchronize access to the currentID and database
	mutex.Lock()
	// Set chirp ID and increment global ID counter
	chirp.ID = currentID
	currentID++
	mutex.Unlock()


	// Add chirp to database
	if err := addChirpToDatabase(chirp); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save chirp: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Send the chirp back in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(chirp); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}


func HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Define the file path
	filePath := "database.json"

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, `{"error": "No chirps found"}`, http.StatusNotFound)
		return
	}

	// Read the file contents
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read chirps: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Check if the file content is valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON format: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the JSON data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
}

