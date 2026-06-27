package handlers

import (
	"net/http"
	"time"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VaultRequest struct {
	EncryptedBlob string `json:"encrypted_blob"`
}

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", handleRegister)
	mux.HandleFunc("POST /api/login", handleLogin)
	mux.HandleFunc("POST /api/vault", handleSaveVault)
	mux.HandleFunc("GET /api/vault", handleGetVault)

	return mux
}

// checks if a session token is valid
func requireAuth(r *http.Request) (int64, error) {
	authHeader := r.Header.Get("Authorization")

	rawToken, err := auth.ExtractBearerToken(authHeader)
	if err != nil {
		return 0, err
	}

	tokenHash := auth.HashSessionToken(rawToken)

	userID, err := db.Q.GetSessionByTokenHash(r.Context(), tokenHash)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateAuthRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	err = db.Q.CreateUser(r.Context(), db.CreateUserParams{
		Username:     req.Username,
		PasswordHash: hashed,
	})
	if err != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status": "registration successful",
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateAuthRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := db.Q.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	match, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	rawToken, err := auth.GenerateSessionToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	tokenHash := auth.HashSessionToken(rawToken)

	err = db.Q.CreateSession(r.Context(), db.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"token":   rawToken,
	})
}

func handleSaveVault(w http.ResponseWriter, r *http.Request) {
	userID, err := requireAuth(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req VaultRequest

	if err := decodeJSON(w, r, &req, maxVaultBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateVaultRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = db.Q.SaveVault(r.Context(), db.SaveVaultParams{
		UserID:        userID,
		EncryptedBlob: req.EncryptedBlob,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "vault synced",
	})
}

func handleGetVault(w http.ResponseWriter, r *http.Request) {
	userID, err := requireAuth(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	blob, err := db.Q.GetVaultByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "vault not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"encrypted_blob": blob,
	})
}