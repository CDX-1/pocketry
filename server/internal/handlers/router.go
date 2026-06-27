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
	EncryptedBlob    string `json:"encrypted_blob"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type VaultResponse struct {
	EncryptedBlob string `json:"encrypted_blob"`
	Revision      int64  `json:"revision"`
	UpdatedAt	  string `json:"updated_at"`
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

	// create new vault if expected revision is 0
	if req.ExpectedRevision == 0 {
		err = db.Q.CreateVault(r.Context(), db.CreateVaultParams{
			UserID: 	userID,
			EncryptedBlob: req.EncryptedBlob,
		})

		if err != nil {
			writeError(w, http.StatusConflict, "vault already exists")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"status": "vault created",
			"revision": 1,
		})
		return
	}

	rowsAffected, err := db.Q.UpdateVaultIfRevisionMatches(r.Context(), db.UpdateVaultIfRevisionMatchesParams{
		EncryptedBlob: req.EncryptedBlob,
		UserID: 	   userID,
		Revision: 	   req.ExpectedRevision,
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault")
		return
	}

	if rowsAffected == 0 {
		writeError(w, http.StatusConflict, "vault has been updated by another client; please try again")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "vault synced",
		"revision": req.ExpectedRevision + 1,
	})
}

func handleGetVault(w http.ResponseWriter, r *http.Request) {
	userID, err := requireAuth(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vault, err := db.Q.GetVaultByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "vault not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"encrypted_blob": vault.EncryptedBlob,
		"revision":       vault.Revision,
		"updated_at":      vault.UpdatedAt,
	})
}