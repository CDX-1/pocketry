package main

import (
	"log"
	"net/http"

	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/handlers"
)

func main() {
	if err := db.InitDB("pocketry.db"); err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	mux := handlers.RegisterRoutes()

	log.Println("Pocketry server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
