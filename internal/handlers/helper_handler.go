package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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
		// fmt.Printf("Word: %s\n", word)
		if endsWithPunctuation(word) {
			// fmt.Println("Word has been skip")
			continue
		}
		if _, found := wordsToReplace[strings.ToLower(word)]; found {
			// fmt.Println("Word need to be replaced")
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

func addDataToDatabase(data DataPoint, category string) error {
	// Define the file path
	filePath := "database.json"
	// fmt.Printf("Chirp inside function: %+v\n", chirp)

	// Initialize an empty map to hold the chirps
	database := map[string]map[string]interface{}{
		category: {},
	}

	if database[category] == nil {
		database[category] = make(map[string]interface{})
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read the existing data
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read file: %v", err)
		}
		if err := json.Unmarshal(fileBytes, &database); err != nil {
			return fmt.Errorf("could not unmarshal JSON: %v", err)
		}
	}

	// Add the new chirp
	database[category][fmt.Sprint(data.GetID())] = data

	// Write the updated data back to the file
	fileBytes, err := json.MarshalIndent(database, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal JSON: %v", err)
	}
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}

	return nil
}
