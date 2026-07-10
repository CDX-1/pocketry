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

const (
	pendingAuthTTL = 10 * time.Minute
	accessTokenTTL = 15 * time.Minute
)

type Handler struct {
	opaque auth.OpaqueServer
}

// request/response structs
type RegisterStartRequest struct {
	Username      string `json:"username"`
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
	Username      string `json:"username"`
	ClientMessage string `json:"client_message"`
}

type LoginStartResponse struct {
	LoginID       string `json:"login_id"`
	ServerMessage string `json:"server_message"`
}

type LoginFinishRequest struct {
	LoginID       string `json:"login_id"`
	ClientMessage string `json:"client_message"`
}

type LoginFinishResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type VaultRequest struct {
	EncryptedBlob    vault.Envelope `json:"encrypted_blob"`
	ExpectedRevision int64          `json:"expected_revision"`
}

type VaultResponse struct {
	EncryptedBlob vault.Envelope `json:"encrypted_blob"`
	Revision      int64          `json:"revision"`
	UpdatedAt     string         `json:"updated_at"`
}

func RegisterRoutes(opaque auth.OpaqueServer) *http.ServeMux {
	if opaque == nil {
		panic("handlers: opaque server is nil")
	}

	h := &Handler{
		opaque: opaque,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register/start", h.handleRegisterStart)
	mux.HandleFunc("POST /api/auth/register/finish", h.handleRegisterFinish)

	mux.HandleFunc("POST /api/auth/login/start", h.handleLoginStart)
	mux.HandleFunc("POST /api/auth/login/finish", h.handleLoginFinish)

	mux.HandleFunc("GET /api/me", h.handleMe)

	mux.HandleFunc("POST /api/vault", h.handleSaveVault)
	mux.HandleFunc("GET /api/vault", h.handleGetVault)

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

// -- /api/auth/register/start
func (h *Handler) handleRegisterStart(w http.ResponseWriter, r *http.Request) {
	var req RegisterStartRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	usernameNormalized := auth.NormalizeUsername(username)

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateOpaqueClientMessage(req.ClientMessage); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	serverMessage, serverState, err := h.opaque.RegisterStart(
		[]byte(usernameNormalized),
		clientMessage,
	)
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
		ID:                 registrationID,
		Username:           username,
		UsernameNormalized: usernameNormalized,
		ServerState:        serverState,
		ExpiresAt:          time.Now().Add(pendingAuthTTL),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save pending registration")
		return
	}

	writeJSON(w, http.StatusOK, RegisterStartResponse{
		RegistrationID: registrationID,
		ServerMessage:  base64.RawURLEncoding.EncodeToString(serverMessage),
	})
}

// -- /api/auth/register/finish
func (h *Handler) handleRegisterFinish(w http.ResponseWriter, r *http.Request) {
	var req RegisterFinishRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateFlowID(req.RegistrationID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateOpaqueClientMessage(req.ClientMessage); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	registrationRecord, err := h.opaque.RegisterFinish(pending.ServerState, clientMessage)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to finish registration")
		return
	}

	err = db.Q.CreateUser(r.Context(), db.CreateUserParams{
		Username:                 pending.Username,
		UsernameNormalized:       pending.UsernameNormalized,
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
		"status":   "registration successful",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// -- /api/auth/login/start
func (h *Handler) handleLoginStart(w http.ResponseWriter, r *http.Request) {
	var req LoginStartRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateOpaqueClientMessage(req.ClientMessage); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	usernameNormalized := auth.NormalizeUsername(username)

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

	serverMessage, serverState, err := h.opaque.LoginStart(user.OpaqueRegistrationRecord, clientMessage)
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
		ID:          loginID,
		UserID:      user.ID,
		ServerState: serverState,
		ExpiresAt:   time.Now().Add(pendingAuthTTL),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save pending login")
		return
	}

	writeJSON(w, http.StatusOK, LoginStartResponse{
		LoginID:       loginID,
		ServerMessage: base64.RawURLEncoding.EncodeToString(serverMessage),
	})
}

// -- /api/auth/login/finish
func (h *Handler) handleLoginFinish(w http.ResponseWriter, r *http.Request) {
	var req LoginFinishRequest

	if err := decodeJSON(w, r, &req, maxAuthBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateFlowID(req.LoginID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateOpaqueClientMessage(req.ClientMessage); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	if err := h.opaque.LoginFinish(pending.ServerState, clientMessage); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

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
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
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
		"id":       user.ID,
		"username": user.Username,
	})
}

func (h *Handler) handleSaveVault(w http.ResponseWriter, r *http.Request) {
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

	// Create a new vault if expected revision is 0
	if req.ExpectedRevision == 0 {
		err = db.Q.CreateVault(r.Context(), db.CreateVaultParams{
			UserID:        userID,
			EncryptedBlob: encryptedBlobJSON,
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
		EncryptedBlob: encryptedBlobJSON,
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

func (h *Handler) handleGetVault(w http.ResponseWriter, r *http.Request) {
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
	if err := json.Unmarshal(vaultRecord.EncryptedBlob, &env); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unmarshal vault")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"encrypted_blob": env,
		"revision":       vaultRecord.Revision,
		"updated_at":     vaultRecord.UpdatedAt.Format(time.RFC3339),
	})
}
