package main

import (
	"log"
	"net/http"
	"github.com/RichardHoa/go-server/internal/config"
	"github.com/RichardHoa/go-server/internal/route"
	"github.com/joho/godotenv"
	"os"

)

func main() {

	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	const port = "8080"

	// Initialize apiConfig
	apiCfg := &config.ApiConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		PolkaAPIKey: os.Getenv("POLKA_API_KEY"),
	}

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
