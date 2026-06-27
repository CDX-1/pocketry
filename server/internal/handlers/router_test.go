package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CDX-1/pocketry/internal/db"
)

// test helpers

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := db.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	t.Cleanup(func() {
		_ = db.DB.Close()
	})

	return RegisterRoutes()
}

func makeJSONRequest(t *testing.T, method string, path string, body any, token string) *http.Request {
	t.Helper()

	var buf bytes.Buffer

	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func decodeResponse[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var result T

	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return result
}

// checks the register, login, save vault, retrieve vault flow
func TestRegisterLoginSaveAndGetVault(t *testing.T) {
	server := setupTestServer(t)

	registerReq := makeJSONRequest(t, http.MethodPost, "/api/register", AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}, "")

	registerRR := httptest.NewRecorder()
	server.ServeHTTP(registerRR, registerReq)

	if registerRR.Code != http.StatusCreated {
		t.Fatalf("expected register status %d, got %d body=%s", http.StatusCreated, registerRR.Code, registerRR.Body.String())
	}

	loginReq := makeJSONRequest(t, http.MethodPost, "/api/login", AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}, "")

	loginRR := httptest.NewRecorder()
	server.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d body=%s", http.StatusOK, loginRR.Code, loginRR.Body.String())
	}

	loginBody := decodeResponse[map[string]string](t, loginRR)
	token := loginBody["token"]

	if token == "" {
		t.Fatal("expected login response to include token")
	}

	saveReq := makeJSONRequest(t, http.MethodPost, "/api/vault", VaultRequest{
		EncryptedBlob:    "placeholder-vault-blob",
		ExpectedRevision: 0,
	}, token)

	saveRR := httptest.NewRecorder()
	server.ServeHTTP(saveRR, saveReq)

	if saveRR.Code != http.StatusCreated {
		t.Fatalf("expected save vault status: %d, got %d body=%s", http.StatusOK, saveRR.Code, saveRR.Body.String())
	}

	getReq := makeJSONRequest(t, http.MethodGet, "/api/vault", nil, token)

	getRR := httptest.NewRecorder()
	server.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected get vault status %d, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	getBody := decodeResponse[map[string]any](t, getRR)

	if getBody["encrypted_blob"] != "placeholder-vault-blob" {
		t.Fatalf("unexpected vault blob: %q", getBody["encrypted_blob"])
	}

	if getBody["revision"].(float64) != 1 {
		t.Fatalf("expected revision 1, got %v", getBody["revision"])
	}
}

// checks that the /vault endpoint requires authentication
func TestVaultRequiresAuth(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(t, http.MethodGet, "/api/vault", nil, "")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// checks that the login endpoint rejects wrong passwords
func TestLoginRejectsWrongPassword(t *testing.T) {
	server := setupTestServer(t)

	registerReq := makeJSONRequest(t, http.MethodPost, "/api/register", AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}, "")

	registerRR := httptest.NewRecorder()
	server.ServeHTTP(registerRR, registerReq)

	if registerRR.Code != http.StatusCreated {
		t.Fatalf("expected register status %d, got %d body=%s", http.StatusCreated, registerRR.Code, registerRR.Body.String())
	}

	loginReq := makeJSONRequest(t, http.MethodPost, "/api/login", AuthRequest{
		Username: "username",
		Password: "WrongPassword123!",
	}, "")

	loginRR := httptest.NewRecorder()
	server.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected login status %d, got %d body=%s", http.StatusUnauthorized, loginRR.Code, loginRR.Body.String())
	}
}

// checks that the register endpoint rejects duplicate usernames
func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	server := setupTestServer(t)

	body := AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}

	firstReq := makeJSONRequest(t, http.MethodPost, "/api/register", body, "")
	firstRR := httptest.NewRecorder()
	server.ServeHTTP(firstRR, firstReq)

	if firstRR.Code != http.StatusCreated {
		t.Fatalf("expected first register status %d, got %d body=%s", http.StatusCreated, firstRR.Code, firstRR.Body.String())
	}

	secondReq := makeJSONRequest(t, http.MethodPost, "/api/register", body, "")
	secondRR := httptest.NewRecorder()
	server.ServeHTTP(secondRR, secondReq)

	if secondRR.Code != http.StatusConflict {
		t.Fatalf("expected duplicate register status %d, got %d body=%s", http.StatusConflict, secondRR.Code, secondRR.Body.String())
	}
}

// checks that the /vault endpoint rejects fake tokens
func TestVaultRejectsFakeToken(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(t, http.MethodGet, "/api/vault", nil, "fake-token")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

// checks that the /register endpoint rejects invalid JSON
func TestRegisterRejectsInvalidJSON(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

// checks that the /register endpoint rejects unknown JSON fields
func TestRegisterRejectsUnknownJSONField(t *testing.T) {
	server := setupTestServer(t)

	body := strings.NewReader(`{
		"username": "username",
		"password": "StrongPassword123!",
		"role": "admin"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/register", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

// checks whether a concurrent update to vault is rejected
func TestVaultRevisionConflict(t *testing.T) {
	server := setupTestServer(t)

	registerReq := makeJSONRequest(t, http.MethodPost, "/api/register", AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}, "")

	registerRR := httptest.NewRecorder()
	server.ServeHTTP(registerRR, registerReq)

	if registerRR.Code != http.StatusCreated {
		t.Fatalf("expected register status %d, got %d body=%s", http.StatusCreated, registerRR.Code, registerRR.Body.String())
	}

	loginReq := makeJSONRequest(t, http.MethodPost, "/api/login", AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}, "")

	loginRR := httptest.NewRecorder()
	server.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d body=%s", http.StatusOK, loginRR.Code, loginRR.Body.String())
	}

	loginBody := decodeResponse[map[string]string](t, loginRR)
	token := loginBody["token"]

	createReq := makeJSONRequest(t, http.MethodPost, "/api/vault", VaultRequest{
		EncryptedBlob:    "revision-1-data",
		ExpectedRevision: 0,
	}, token)

	createRR := httptest.NewRecorder()
	server.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create vault status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	updateReq := makeJSONRequest(t, http.MethodPost, "/api/vault", VaultRequest{
		EncryptedBlob:    "revision-2-data",
		ExpectedRevision: 1,
	}, token)

	updateRR := httptest.NewRecorder()
	server.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected update vault status %d, got %d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	conflictReq := makeJSONRequest(t, http.MethodPost, "/api/vault", VaultRequest{
		EncryptedBlob:    "stale-device-data",
		ExpectedRevision: 1,
	}, token)

	conflictRR := httptest.NewRecorder()
	server.ServeHTTP(conflictRR, conflictReq)

	if conflictRR.Code != http.StatusConflict {
		t.Fatalf("expected conflict status %d, got %d body=%s", http.StatusConflict, conflictRR.Code, conflictRR.Body.String())
	}
}