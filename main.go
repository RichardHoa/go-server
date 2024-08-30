package main

import (
	"log"
	"net/http"
	"github.com/RichardHoa/go-server/internal/config"
	"github.com/RichardHoa/go-server/internal/route"
)

func main() {
	const port = "8080"

	// Initialize apiConfig
	apiCfg := &config.ApiConfig{}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Configure routes
	route.ConfigureRoutes(mux, apiCfg)

	// Create and start the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", ".", port)
	log.Fatal(server.ListenAndServe())
}
