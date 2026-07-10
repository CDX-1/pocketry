package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if err := db.InitDB("pocketry.db"); err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	accessTokenSecret := []byte(os.Getenv("POCKETRY_ACCESS_TOKEN_SECRET"))
	if err := auth.SetAccessTokenSecret(accessTokenSecret); err != nil {
		log.Fatal(err)
	}

	var opaqueKeyMaterial []byte
	if encoded := os.Getenv("POCKETRY_OPAQUE_KEY_MATERIAL"); encoded != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			log.Fatal(err)
		}

		opaqueKeyMaterial = decoded
	}

	opaqueServer, err := auth.NewBytemareOpaqueServer("pocketry-server", opaqueKeyMaterial)
	if err != nil {
		log.Fatal(err)
	}

	mux := handlers.RegisterRoutes(opaqueServer)

	log.Println("Pocketry server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}