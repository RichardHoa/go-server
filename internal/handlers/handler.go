package handlers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	usersID = 1
	mutex   sync.Mutex
)

// HandlerReadiness handles the /healthz endpoint
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
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
		Chirps map[string]Chirp `json:"chirps"`
	}

	if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON format: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Convert the map to a slice of Chirp structs
	chirpsArray := []Chirp{}
	for _, chirpData := range jsonData.Chirps {

		chirp := Chirp{
			ID:       chirpData.ID,
			Body:     chirpData.Body,
			AuthorID: chirpData.AuthorID,
		}

		// Append the chirp to the array with the correct ID and body
		chirpsArray = append(chirpsArray, chirp)
	}

	// Filter by author_id if provided
	authorIDStr := r.URL.Query().Get("author_id")
	if authorIDStr != "" {
		authorID, err := strconv.Atoi(authorIDStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid author_id format"}`, http.StatusBadRequest)
			return
		}
		var filteredChirps []Chirp
		for _, chirp := range chirpsArray {
			if chirp.AuthorID == authorID {
				filteredChirps = append(filteredChirps, chirp)
			}
		}
		chirpsArray = filteredChirps
	}
	// Determine the sorting order from the query parameter
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "asc" // Default to ascending order
	}

	// Sort chirps by ID
	sort.Slice(chirpsArray, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirpsArray[i].ID > chirpsArray[j].ID
		}
		return chirpsArray[i].ID < chirpsArray[j].ID
	})

	// Set the response headers and write the JSON array of chirps
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(chirpsArray); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func HandlerGetChirpsID(w http.ResponseWriter, r *http.Request) {
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
		Chirps map[string]Chirp `json:"chirps"`
	}

	if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON format: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract chirpID from the URL
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/chirps/") {
		http.Error(w, `{"error": "Invalid URL format"}`, http.StatusBadRequest)
		return
	}

	// Extract chirpID by splitting the path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error": "Invalid URL format"}`, http.StatusBadRequest)
		return
	}

	chirpID := parts[2]

	// Look up the chirp in the map
	chirp, exists := jsonData.Chirps[chirpID]
	if !exists {
		http.Error(w, `{"error": "Chirp not found"}`, http.StatusNotFound)
		return
	}

	// Set the content type and encode the chirp into the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chirp); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Error encoding response: %v"}`, err), http.StatusInternalServerError)
	}
}

func HandlerAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var user User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Hash the user's password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Update the user's password with the hashed password
	user.Password = string(hashedPassword)

	mutex.Lock()
	user.ID = usersID // Use the setter method
	usersID++
	mutex.Unlock()

	if err := AddDataToDatabase(w, user, "users"); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save chirp: %v"}`, err), http.StatusInternalServerError)
		mutex.Lock()
		usersID--
		mutex.Unlock()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Include ID in the response explicitly
	response := map[string]interface{}{
		"id":            user.GetID(),
		"email":         user.Email,
		"is_chirpy_red": user.IsChirpyRed,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
