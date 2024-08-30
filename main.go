package main

import (
	"log"
	"net/http"
)

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the readiness endpoint at /healthz
	mux.HandleFunc("/healthz", readinessHandler)

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", fileServer))

	// Create a new http.Server struct
	server := &http.Server{
		Addr:    ":8080", 
		Handler: mux,     
	}

	log.Println("Starting server on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// readinessHandler responds with a 200 OK status and a plain text "OK" message
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Write the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write the status code
	w.WriteHeader(http.StatusOK)

	// Write the body text
	w.Write([]byte("OK"))
}
