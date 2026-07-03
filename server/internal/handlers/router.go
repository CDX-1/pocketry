package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/vault"
)

// request/response structs
type RegisterStartRequest struct {
	Username	  string `json:"username"`
	ClientMessage string `json:"client_message"`
}

type RegisterStartResponse struct {
	RegistrationID string `json:"registration_id"`
	ServerMessage  string `json:"server_message"`
}

type RegisterFinishRequest struct {
	RegistrationID string `json:"registration_id"`
	ClientMessage  string `json:"client_message"`
}

type LoginStartRequest struct {
	Username	  string `json:"username"`
	ClientMessage string `json:"client_message"`
}

type LoginStartResponse struct {
	LoginID 	  string `json:"login_id"`
	ServerMessage string `json:"server_message"`
}

type LoginFinishRequest struct {
	LoginID	  	  string `json:"login_id"`
	ClientMessage string `json:"client_message"`
}

type LoginFinishResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

var opaqueServer auth.OpaqueServer

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register/start", handleRegisterStart)
	mux.HandleFunc("POST /api/auth/register/finish", handleRegisterFinish)

	mux.HandleFunc("POST /api/auth/login/start", handleLoginStart)
	mux.HandleFunc("POST /api/auth/login/finish", handleLoginFinish)

	mux.HandleFunc("GET /api/me", handleMe)

	mux.HandleFunc("POST /api/vault", handleSaveVault)
	mux.HandleFunc("GET /api/vault", handleGetVault)

	return mux
}

// checks if an access token is valid
func requireAuth(r *http.Request) (int64, error) {
	authHeader := r.Header.Get("Authorization")

	rawToken, err := auth.ExtractBearerToken(authHeader)
	if err != nil {
		return 0, err
	}

	claims, err := auth.VerifyAccessToken(rawToken)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// -- /api/register/start
func handleRegisterStart(w http.ResponseWriter, r *http.Request) {
	var req RegisterStartRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	usernameNormalized := auth.NormalizeUsername(username)

	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	exists, err := db.Q.UsernameExists(r.Context(), usernameNormalized)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check username")
		return
	}

	if exists {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	clientMessage, err := base64.RawURLEncoding.DecodeString(req.ClientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid client message")
		return
	}

	serverMessage, serverState, err := opaqueServer.RegisterStart(clientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to start registration")
		return
	}

	registrationID, err := auth.NewFlowID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create registration")
		return
	}

	_ = db.Q.DeleteExpiredPendingRegistrations(r.Context())
	_ = db.Q.DeletePendingRegistrationsByUsernameNormalized(r.Context(), usernameNormalized)

	err = db.Q.CreatePendingRegistration(r.Context(), db.CreatePendingRegistrationParams{
		ID:					registrationID,
		Username:			username,
		UsernameNormalized: usernameNormalized,
		ServerState: 		serverState,
		ExpiresAt: 			time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save pending registration")
	}

	writeJSON(w, http.StatusOK, RegisterStartResponse{
		RegistrationID: registrationID,
		ServerMessage:  base64.RawURLEncoding.EncodeToString(serverMessage),
	})
}

// -- /api/register/finish
func handleRegisterFinish(w http.ResponseWriter, r *http.Request) {
	var req RegisterFinishRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pending, err := db.Q.GetPendingRegistration(r.Context(), req.RegistrationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "registration not found or expired")
		return
	}

	if time.Now().After(pending.ExpiresAt) {
		_ = db.Q.DeletePendingRegistration(r.Context(), req.RegistrationID)
		writeError(w, http.StatusBadRequest, "registration expired")
		return
	}

	clientMessage, err := base64.RawURLEncoding.DecodeString(req.ClientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid client message")
		return
	}

	registrationRecord, err := opaqueServer.RegisterFinish(pending.ServerState, clientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to finish registration")
		return
	}

	err = db.Q.CreateUser(r.Context(), db.CreateUserParams{
		Username:				  pending.Username,
		UsernameNormalized: 	  pending.UsernameNormalized,
		OpaqueRegistrationRecord: registrationRecord,
	})
	if err != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	_ = db.Q.DeletePendingRegistration(r.Context(), req.RegistrationID)

	user, err := db.Q.GetUserByUsernameNormalized(r.Context(), pending.UsernameNormalized)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status": 	"registration successful",
		"user_id":	user.ID,
		"username": user.Username,
	})
}

// -- /api/login/start
func handleLoginStart(w http.ResponseWriter, r *http.Request) {
	var req LoginStartRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	usernameNormalized := auth.NormalizeUsername(req.Username)

	user, err := db.Q.GetUserByUsernameNormalized(r.Context(), usernameNormalized)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	clientMessage, err := base64.RawURLEncoding.DecodeString(req.ClientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid client message")
		return
	}

	serverMessage, serverState, err := opaqueServer.LoginStart(user.OpaqueRegistrationRecord, clientMessage)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	loginID, err := auth.NewFlowID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create login")
		return
	}

	_ = db.Q.DeleteExpiredPendingLogins(r.Context())
	_ = db.Q.DeletePendingLoginsByUserID(r.Context(), user.ID)

	err = db.Q.CreatePendingLogin(r.Context(), db.CreatePendingLoginParams{
		ID: 		 loginID,
		UserID: 	 user.ID,
		ServerState: serverState,
		ExpiresAt: 	 time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save pending login")
		return
	}

	writeJSON(w, http.StatusOK, LoginStartResponse{
		LoginID: 	   loginID,
		ServerMessage: base64.RawURLEncoding.EncodeToString(serverMessage),
	})
}

// -- /api/login/finish
func handleLoginFinish(w http.ResponseWriter, r *http.Request) {
	var req LoginFinishRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pending, err := db.Q.GetPendingLogin(r.Context(), req.LoginID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if time.Now().After(pending.ExpiresAt) {
		_ = db.Q.DeletePendingLogin(r.Context(), req.LoginID)
		writeError(w, http.StatusUnauthorized, "login expired")
		return
	}

	clientMessage, err := base64.RawURLEncoding.DecodeString(req.ClientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid client message")
		return
	}

	if err := opaqueServer.LoginFinish(pending.ServerState, clientMessage); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessTokenTTL := 15 * time.Minute

	accessToken, err := auth.IssueAccessToken(pending.UserID, accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue access token")
		return
	}

	_ = db.Q.DeletePendingLogin(r.Context(), req.LoginID)

	writeJSON(w, http.StatusOK, LoginFinishResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTokenTTL.Seconds()),
	})
}

// -- /api/me
func handleMe(w http.ResponseWriter, r *http.Request) {
	userID, err := requireAuth(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := db.Q.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":	    user.ID,
		"username": user.Username,
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

	encryptedBlobJSON, err := json.Marshal(req.EncryptedBlob)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid encrypted_blob")
		return
	}

	// create new vault if expected revision is 0
	if req.ExpectedRevision == 0 {
		err = db.Q.CreateVault(r.Context(), db.CreateVaultParams{
			UserID:        userID,
			EncryptedBlob: string(encryptedBlobJSON),
		})

		if err != nil {
			writeError(w, http.StatusConflict, "vault already exists")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"status":   "vault created",
			"revision": 1,
		})
		return
	}

	rowsAffected, err := db.Q.UpdateVaultIfRevisionMatches(r.Context(), db.UpdateVaultIfRevisionMatchesParams{
		EncryptedBlob: string(encryptedBlobJSON),
		UserID:        userID,
		Revision:      req.ExpectedRevision,
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
		"status":   "vault synced",
		"revision": req.ExpectedRevision + 1,
	})
}

func handleGetVault(w http.ResponseWriter, r *http.Request) {
	userID, err := requireAuth(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vaultRecord, err := db.Q.GetVaultByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "vault not found")
		return
	}

	var env vault.Envelope
	if err := json.Unmarshal([]byte(vaultRecord.EncryptedBlob), &env); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unmarshal vault")
		return
	}

	var updatedAtStr string
    if vaultRecord.UpdatedAt.Valid {
        updatedAtStr = vaultRecord.UpdatedAt.Time.Format(time.RFC3339)
    }

	writeJSON(w, http.StatusOK, map[string]any{
		"encrypted_blob": env,
		"revision":       vaultRecord.Revision,
		"updated_at":     updatedAtStr,
	})
}
