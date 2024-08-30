package handlers

import (
	// "fmt"
	"encoding/json"
	"net/http"
	// "regexp"
	"strings"
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

	// Replace sensitive words in the chirp body
	cleanedBody := ReplaceSensitiveWords(chirp.Body)

	// If validation passes, respond with a 200 status code
	response := map[string]string{"cleaned_body": cleanedBody}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}

}


func ReplaceSensitiveWords(text string) string {
	wordsToReplace := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	// Split the text into words
	words := strings.Fields(text)
	for i, word := range words {
		// Check for punctuation at the end of the word
		if endsWithPunctuation(word) {
			continue
		}
		if _, found := wordsToReplace[strings.ToLower(word)]; found {
			// Replace the word if it should be replaced
			words[i] = strings.ReplaceAll(word, word, "****")
		}
	}

	// Join the words back into a single string
	return strings.Join(words, " ")
}


func endsWithPunctuation(word string) bool {
	return len(word) > 0 && strings.ContainsAny(word[len(word)-1:], ".,!?")
}