package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"strconv"
)

type Chirp struct {
	ID   int `json:"id"`
	Body string `json:"body"`
}


// Getter for ID
func (c *Chirp) GetID() int {
	return c.ID
}

// Setter for ID
func (c *Chirp) SetID(id int) {
	c.ID = id
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
	var chirp Chirp
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	chirp.SetID(currentID) // Use the setter method
	currentID++
	mutex.Unlock()

	// fmt.Printf("Chirp: %+v\n", chirp)
	// Access ID with chirp.GetID() when needed

	if err := addChirpToDatabase(chirp); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save chirp: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Include ID in the response explicitly
	response := map[string]interface{}{
		"id":   chirp.GetID(),
		"body": chirp.Body,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
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
// Parse the JSON file content into a map
var jsonData struct {
	Chirps map[string]struct {
		Body string `json:"body"`
	} `json:"chirps"`
}

if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
	http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON format: %v"}`, err), http.StatusInternalServerError)
	return
}

// Convert the map to a slice of Chirp structs
chirpsArray := []Chirp{}
for idStr, chirpData := range jsonData.Chirps {
	// fmt.Printf("ID: %s, Body: %s\n", idStr, chirpData.Body)
	// Convert string id to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid chirp ID: %v"}`, err), http.StatusInternalServerError)
		return
	}
	chirp := Chirp{
		Body: chirpData.Body,
	}
	chirp.SetID(id)
	// fmt.Printf("Chrip: %v\n", chirp)

	// Append the chirp to the array with the correct ID and body
	chirpsArray = append(chirpsArray,chirp)
	// fmt.Printf("Chirps: %v\n", chirpsArray)
}

// Set the response headers and write the JSON array of chirps
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
if err := json.NewEncoder(w).Encode(chirpsArray); err != nil {
	http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
}
}
