package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RichardHoa/go-server/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ApiConfig struct {
	FileserverHits int
	JWTSecret      string
	PolkaAPIKey    string
	Mu             sync.Mutex // Mutex to ensure safe concurrent access to FileserverHits
}

var (
	chirpsID = 1

	mutex sync.Mutex
)

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

func (cfg *ApiConfig) HandlerAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cfg.Mu.Lock()
	JWTSecret := cfg.JWTSecret
	cfg.Mu.Unlock()

	var user handlers.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Define the file path to the database
	filePath := "database.json"

	// Define a struct for the entire database

	// Initialize the database struct
	var database handlers.Database

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data into the database struct
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Ensure chirps is initialized if it's null
	if database.Chirps == nil {
		database.Chirps = make(map[string]handlers.Chirp)
	}

	// Extract the "users" map
	users := database.Users

	// Iterate over users to find a match by email
	for _, storedUser := range users {
		if storedUser.GetUniqueIdentifier() == user.GetUniqueIdentifier() {
			// Compare the hashed password with the provided password
			if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
				http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
				return
			}

			// JWT token generation
			secretKey := []byte(JWTSecret)
			expiresInSeconds := 1 * 60 * 60 // Default to 1 hour

			// Override with client-provided expiration time if valid
			if user.ExpiresInSeconds > 0 && user.ExpiresInSeconds <= 24*60*60 {
				expiresInSeconds = user.ExpiresInSeconds
			}

			// Create the JWT claims
			claims := jwt.RegisteredClaims{
				Issuer:    "chirpy",
				Subject:   fmt.Sprintf("%v", storedUser.GetID()),
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(expiresInSeconds))),
			}

			// Create the token with HS256 signing method
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			// Sign the token with the secret key
			tokenString, err := token.SignedString(secretKey)
			if err != nil {
				http.Error(w, `{"error": "Failed to sign token"}`, http.StatusInternalServerError)
				return
			}

			// Generate refresh token
			refreshTokenBytes := make([]byte, 32)
			_, err = rand.Read(refreshTokenBytes)
			if err != nil {
				http.Error(w, `{"error": "Failed to generate refresh token"}`, http.StatusInternalServerError)
				return
			}
			refreshToken := hex.EncodeToString(refreshTokenBytes)

			// Store refresh token and its expiration date in the database
			storedUser.RefreshToken = refreshToken
			storedUser.RefreshTokenExpiresAt = time.Now().UTC().Add(60 * 24 * time.Hour) // 60 days

			// Update the user in the database
			database.Users[fmt.Sprintf("%v", storedUser.GetID())] = storedUser

			// Write updated data back to the database, preserving chirps
			fileBytes, err = json.MarshalIndent(database, "", "  ")
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "Failed to serialize database: %v"}`, err), http.StatusInternalServerError)
				return
			}
			if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "Failed to write database: %v"}`, err), http.StatusInternalServerError)
				return
			}

			// Respond with the token and refresh token
			response := map[string]interface{}{
				"id":            storedUser.GetID(),
				"email":         storedUser.GetUniqueIdentifier(),
				"refresh_token": refreshToken,
				"token":         tokenString,
				"is_chirpy_red": storedUser.IsChirpyRed,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// If no user is found with the given email
	http.Error(w, `{"error": "User email does not exist"}`, http.StatusNotFound)
}

func (cfg *ApiConfig) HandlerPutUser(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cfg.Mu.Lock()
	JWTSecret := cfg.JWTSecret
	cfg.Mu.Unlock()

	// Extract the JWT from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error": "Invalid or missing Authorization header"}`, http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Define the claims struct
	claims := &jwt.RegisteredClaims{}

	// Parse the JWT token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil // Replace with your actual secret key
	})
	if err != nil || !token.Valid {
		http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	// Extract user ID from the token's Subject field
	userID := claims.Subject

	// Define the file path to the database
	filePath := "database.json"

	// Initialize a map to hold the data from the database
	database := map[string]map[string]handlers.User{}

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract the "users" map
	users := database["users"]

	// Find the user by ID
	user, exists := users[userID]
	if !exists {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Define a struct to parse the incoming JSON body
	type updatedUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var updatedUserObject updatedUser
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedUserObject); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedUserObject.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Update the user's email and password in the database
	user.Email = updatedUserObject.Email
	user.Password = string(hashedPassword)

	// Save the updated user back to the database
	users[userID] = user
	database["users"] = users

	// Convert the database map back to JSON
	updatedDatabaseBytes, err := json.Marshal(database)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to serialize database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Write the updated database back to the file
	if err := os.WriteFile(filePath, updatedDatabaseBytes, 0644); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to write to database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Return the updated user in the response
	response := map[string]interface{}{
		"id":    user.GetID(),
		"email": user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (cfg *ApiConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is POST
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cfg.Mu.Lock()
	JWTSecret := cfg.JWTSecret
	cfg.Mu.Unlock()

	// Extract the refresh token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error": "Invalid or missing Authorization header"}`, http.StatusUnauthorized)
		return
	}
	refreshTokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Define the file path to the database
	filePath := "database.json"

	// Initialize a map to hold the data from the database
	database := map[string]map[string]handlers.User{}

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract the "users" map
	users := database["users"]

	// Find the user with the given refresh token
	var storedUser handlers.User
	found := false
	for _, user := range users {
		if user.RefreshToken == refreshTokenString {
			// Check if the refresh token has expired
			if time.Now().UTC().After(user.RefreshTokenExpiresAt) {
				http.Error(w, `{"error": "Refresh token expired"}`, http.StatusUnauthorized)
				return
			}

			storedUser = user
			found = true
			break
		}
	}

	if !found {
		http.Error(w, `{"error": "Invalid or non-existent refresh token"}`, http.StatusUnauthorized)
		return
	}

	// JWT token generation
	secretKey := []byte(JWTSecret)
	expiresInSeconds := 1 * 60 * 60 // 1 hour

	// Create the JWT claims
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   fmt.Sprintf("%v", storedUser.GetID()),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(expiresInSeconds))),
	}

	// Create the token with HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, `{"error": "Failed to sign token"}`, http.StatusInternalServerError)
		return
	}

	// Respond with the new access token
	response := map[string]interface{}{
		"token": tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (cfg *ApiConfig) HandlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is POST
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract the refresh token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error": "Invalid or missing Authorization header"}`, http.StatusUnauthorized)
		return
	}
	refreshTokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Define the file path to the database
	filePath := "database.json"

	// Initialize a map to hold the data from the database
	database := map[string]map[string]handlers.User{}

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract the "users" map
	users := database["users"]

	// Find and update the user with the given refresh token
	var userUpdated bool
	for id, user := range users {
		if user.RefreshToken == refreshTokenString {
			// Check if the refresh token has expired
			if time.Now().UTC().After(user.RefreshTokenExpiresAt) {
				http.Error(w, `{"error": "Refresh token expired"}`, http.StatusUnauthorized)
				return
			}

			// Remove the refresh token and reset the expiration
			user.RefreshToken = ""
			user.RefreshTokenExpiresAt = time.Time{}

			// Update the user in the map
			users[id] = user
			userUpdated = true
			break
		}
	}

	if !userUpdated {
		http.Error(w, `{"error": "Invalid or non-existent refresh token"}`, http.StatusUnauthorized)
		return
	}

	// Update the database file with the removed refresh token
	database["users"] = users
	fileBytes, err = json.Marshal(database)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to serialize database: %v"}`, err), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to write database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) HandlerAddChirps(w http.ResponseWriter, r *http.Request) {
	var chirp handlers.Chirp
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Extract the access token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error": "Invalid or missing Authorization header"}`, http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	cfg.Mu.Lock()
	JWTSecret := cfg.JWTSecret
	cfg.Mu.Unlock()

	// Define the claims struct
	claims := &jwt.RegisteredClaims{}

	// Parse and validate the JWT token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, `{"error": "Invalid or expired access token"}`, http.StatusUnauthorized)
		return
	}

	// Extract the user ID (authorID) from the JWT claims
	authorID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		http.Error(w, `{"error": "Invalid token subject"}`, http.StatusUnauthorized)
		return
	}

	// Set the chirp's AuthorID
	chirp.AuthorID = authorID

	mutex.Lock()
	chirp.SetID(chirpsID) // Use the setter method
	chirpsID++
	mutex.Unlock()

	if err := handlers.AddDataToDatabase(w, chirp, "chirps"); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save chirp: %v"}`, err), http.StatusInternalServerError)
		mutex.Lock()
		chirpsID--
		mutex.Unlock()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Include ID in the response explicitly
	response := map[string]interface{}{
		"id":        chirp.GetID(),
		"body":      chirp.Body,
		"author_id": chirp.AuthorID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func (cfg *ApiConfig) HandlerDeleteChirps(w http.ResponseWriter, r *http.Request) {
	// Extract the chirp ID from the URL path
	chirpIDStr := strings.TrimPrefix(r.URL.Path, "/api/chirps/")

	// Extract the access token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, `{"error": "Invalid or missing Authorization header"}`, http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	cfg.Mu.Lock()
	JWTSecret := cfg.JWTSecret
	cfg.Mu.Unlock()

	// Define the claims struct
	claims := &jwt.RegisteredClaims{}

	// Parse and validate the JWT token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, `{"error": "Invalid or expired access token"}`, http.StatusUnauthorized)
		return
	}

	// Extract the user ID (authorID) from the JWT claims
	authorID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		http.Error(w, `{"error": "Invalid token subject"}`, http.StatusUnauthorized)
		return
	}

	// Define the file path to the database
	filePath := "database.json"

	// Initialize a map to hold the data from the database
	var database handlers.Database

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract the "chirps" map
	chirps := database.Chirps

	// Find the chirp by ID
	chirp, exists := chirps[chirpIDStr]
	if !exists {
		http.Error(w, `{"error": "Chirp not found"}`, http.StatusNotFound)
		return
	}

	// Check if the chirp's author ID matches the token's author ID
	if chirp.AuthorID != authorID {
		http.Error(w, `{"error": "Forbidden: You do not have permission to delete this chirp"}`, http.StatusForbidden)
		return
	}

	delete(chirps, chirpIDStr)

	// Write the updated database back to the file
	updatedDatabase, err := json.MarshalIndent(database, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to marshal database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(filePath, updatedDatabase, 0644); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to write database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) HandlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	expectedPrefix := "ApiKey "

	if !strings.HasPrefix(authHeader, expectedPrefix) || authHeader[len(expectedPrefix):] != cfg.PolkaAPIKey {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	type WebhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	var webhookReq WebhookRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&webhookReq); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Check if the event is "user.upgraded", otherwise return 204
	if webhookReq.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Define the file path to the database
	filePath := "database.json"

	// Initialize a map to hold the data from the database
	var database handlers.Database

	// Read the database file
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON data
	if err := json.Unmarshal(fileBytes, &database); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Extract the "users" map
	users := database.Users

	// Find the user by user_id from the webhook data
	userID := fmt.Sprintf("%d", webhookReq.Data.UserID)
	user, exists := users[userID]
	if !exists {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Update the user's IsChirpyRed field
	user.IsChirpyRed = true

	// Update the database with the modified user
	users[userID] = user
	database.Users = users

	// Write the updated data back to the database
	fileBytes, err = json.Marshal(database)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to serialize database: %v"}`, err), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to write database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with a 204 status code
	w.WriteHeader(http.StatusNoContent)

}
