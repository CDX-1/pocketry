package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/handlers"
	"github.com/CDX-1/pocketry/internal/middleware"
	"github.com/joho/godotenv"
)

func parseAllowedOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)

		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}

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

	encodedOpaqueKeyMaterial := os.Getenv("POCKETRY_OPAQUE_KEY_MATERIAL")
	if encodedOpaqueKeyMaterial == "" {
		log.Fatal("POCKETRY_OPAQUE_KEY_MATERIAL is required")
	}

	opaqueKeyMaterial, err := base64.RawURLEncoding.DecodeString(encodedOpaqueKeyMaterial)
	if err != nil {
		log.Fatalf("failed to decode POCKETRY_OPAQUE_KEY_MATERIAL: %v", err)
	}

	opaqueServer, err := auth.NewBytemareOpaqueServer("pocketry-server", opaqueKeyMaterial)
	if err != nil {
		log.Fatal(err)
	}

	mux := handlers.RegisterRoutes(opaqueServer)
	allowedOrigins := parseAllowedOrigins(os.Getenv("POCKETRY_ALLOWED_ORIGINS"))
	allowFirefoxExtensions := strings.EqualFold(
		strings.TrimSpace(os.Getenv("POCKETRY_ALLOW_FIREFOX_EXTENSIONS")),
		"true",
	)

	if len(allowedOrigins) == 0 && !allowFirefoxExtensions {
		log.Fatal("POCKETRY_ALLOWED_ORIGINS is required unless Firefox extension origins are enabled")
	}

	cors := middleware.NewCORS(allowedOrigins, allowFirefoxExtensions)
	handler := cors.Handler(mux)

	log.Println("Pocketry server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
