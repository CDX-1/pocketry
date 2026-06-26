package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = db.Q.CreateUser(r.Context(), db.CreateUserParams{
		Username:     req.Username,
		PasswordHash: hashed,
	})
	if err != nil {
		http.Error(w, "Username taken", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Registration successful"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := db.Q.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	match, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   strconv.FormatInt(user.ID, 10),
	})
}

func handleSaveVault(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	var req VaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err = db.Q.SaveVault(r.Context(), db.SaveVaultParams{
		UserID: userID,
		EncryptedBlob: req.EncryptedBlob,
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "Vault synced"})
}

func handleGetVault(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	blob, err := db.Q.GetVaultByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Vault record empty", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"encrypted_blob": blob})
}