package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Attach handlers
	http.HandleFunc("/", indexHandlers)

	// Fetch port from ENV
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Print("Listing on http://localhost:" + port)

	provider := os.Getenv("PROVIDER")
	fmt.Print("Provider: " + provider)

	// Start listening
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Could not start server: ", err)
	}
}

func indexHandlers(w http.ResponseWriter, r *http.Request) {
	provider := os.Getenv("PROVIDER")
	fmt.Fprintf(w, "PROVIDER %s", provider)
}
