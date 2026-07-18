package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
)

// fake opauqe server, avoids testing OPAQUE cryptography and tests purely router code

type fakeOpaqueServer struct {
	registerStartErr	error
	registerFinishErr	error
	loginStartErr		error
	loginFinishErr		error
}

func (f *fakeOpaqueServer) RegisterStart(
	credentialIdentifier []byte,
	clientMessage        []byte,
) ([]byte, []byte, error) {
	if f.registerStartErr != nil {
		return nil, nil, f.registerStartErr
	}

	serverMessage := append([]byte("register-server-message"), credentialIdentifier...)
	serverState := append([]byte(nil), clientMessage...)

	return serverMessage, serverState, nil
}

func (f *fakeOpaqueServer) RegisterFinish(
	serverState   []byte,
	clientMessage []byte,
) ([]byte, error) {
	if f.registerFinishErr != nil {
		return nil, f.registerFinishErr
	}

	if !bytes.Equal(serverState, clientMessage) {
		return nil, errors.New("registration client message mismatch")
	}

	return append([]byte(nil), clientMessage...), nil
}

func (f *fakeOpaqueServer) LoginStart(
	registrationRecord []byte,
	clientMessage      []byte,
) ([]byte, []byte, error) {
	if f.loginStartErr != nil {
		return nil, nil, f.loginStartErr
	}

	serverMessage := []byte("login-server-message")

	if bytes.Equal(registrationRecord, clientMessage) {
		return serverMessage, []byte("valid"), nil
	}

	return serverMessage, []byte("invalid"), nil
}

func (f *fakeOpaqueServer) LoginFinish(
	serverState   []byte,
	clientMessage []byte,
) error {
	if f.loginFinishErr != nil {
		return f.loginFinishErr
	}

	if !bytes.Equal(serverState, []byte("valid")) {
		return errors.New("invalid credentials")
	}

	if len(clientMessage) == 0 {
		return errors.New("empty login finish message")
	}

	return nil
}

// test helpers

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	err := auth.SetAccessTokenSecret(
		[]byte("0123456789abcdef0123456789abcdef"),
	)
	if err != nil {
		t.Fatalf("failed to configure access token secret: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open returned err: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	return RegisterRoutes(
		store.Q,
		&fakeOpaqueServer{},
	)
}

func makeJSONRequest(
	t *testing.T,
	method string,
	path string,
	body any,
	token string,
) *http.Request {
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
		req.Header.Set("Authorization", "Bearer " + token)
	}

	return req
}

func decodeResponse[T any](
	t *testing.T,
	rr *httptest.ResponseRecorder,
) T {
	t.Helper()

	var result T

	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf(
			"failed to decode response body: %v; body=%q",
			err,
			rr.Body.String(),
		)
	}

	return result
}

func encodeOpaqueMessage(message []byte) string {
	return base64.RawURLEncoding.EncodeToString(message)
}

func performRegistration(
	t *testing.T,
	server http.Handler,
	username string,
	password string,
) {
	t.Helper()

	registerMessage := []byte(password)

	startReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/start",
		RegisterStartRequest{
			Username: 	   username,
			ClientMessage: encodeOpaqueMessage(registerMessage),
		},
		"",
	)

	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusOK {
		t.Fatalf(
			"expected registration start status %d, got %d body=%s",
			http.StatusOK,
			startRR.Code,
			startRR.Body.String(),
		)
	}

	startBody := decodeResponse[RegisterStartResponse](t, startRR)

	if startBody.RegistrationID == "" {
		t.Fatal("expected registration start response to include registration_id")
	}

	if startBody.ServerMessage == "" {
		t.Fatal("expected registration start response to include server_message")
	}

	finishReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/finish",
		RegisterFinishRequest{
			RegistrationID: startBody.RegistrationID,
			ClientMessage:  encodeOpaqueMessage(registerMessage),
		},
		"",
	)

	finishRR := httptest.NewRecorder()
	server.ServeHTTP(finishRR, finishReq)

	if finishRR.Code != http.StatusCreated {
		t.Fatalf(
			"expected registration finish status %d, got %d body=%s",
			http.StatusCreated,
			finishRR.Code,
			finishRR.Body.String(),
		)
	}
}

func performLogin(
	t *testing.T,
	server http.Handler,
	username string,
	password string,
) string {
	t.Helper()

	loginMessage := []byte(password)
	
	startReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/start",
		LoginStartRequest{
			Username: 	   username,
			ClientMessage: encodeOpaqueMessage(loginMessage),
		},
		"",
	)

	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusOK {
		t.Fatalf(
			"expected login start status %d, got %d body=%s",
			http.StatusOK,
			startRR.Code,
			startRR.Body.String(),
		)
	}

	startBody := decodeResponse[LoginStartResponse](t, startRR)

	if startBody.LoginID == "" {
		t.Fatal("expected login start response to include login_id")
	}

	if startBody.ServerMessage == "" {
		t.Fatal("expected login start response to include server_message")
	}

	finishReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/finish",
		LoginFinishRequest{
			LoginID: 	   startBody.LoginID,
			ClientMessage: encodeOpaqueMessage([]byte("login-finish")),
		},
		"",
	)

	finishRR := httptest.NewRecorder()
	server.ServeHTTP(finishRR, finishReq)

	if finishRR.Code != http.StatusOK {
		t.Fatalf(
			"expected login finish status %d, got %d body=%s",
			http.StatusOK,
			finishRR.Code,
			finishRR.Body.String(),
		)
	}
	
	finishBody := decodeResponse[LoginFinishResponse](t, finishRR)

	if finishBody.AccessToken == "" {
		t.Fatal("expected login finish response to include access_token")
	}

	if finishBody.ExpiresIn != int64(accessTokenTTL.Seconds()) {
		t.Fatalf(
			"expected expires_in %d, got %d",
			int64(accessTokenTTL.Seconds()),
			finishBody.ExpiresIn,
		)
	}

	return finishBody.AccessToken
}

func registerAndLogin(
	t *testing.T,
	server http.Handler,
	username string,
	password string,
) string {
	t.Helper()

	performRegistration(t, server, username, password)
	return performLogin(t, server, username, password)
}

// full OPAQUE registration, login, authenticated user, vault save and vault retrieval flow
func TestRegisterLoginSaveAndGetVault(t *testing.T) {
	server := setupTestServer(t)

	token := registerAndLogin(
		t,
		server,
		"username",
		"password",
	)

	meReq := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/me",
		nil,
		token,
	)

	meRR := httptest.NewRecorder()
	server.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusOK {
		t.Fatalf(
			"expected me status %d, got %d body=%s",
			http.StatusOK,
			meRR.Code,
			meRR.Body.String(),
		)
	}

	meBody := decodeResponse[MeResponse](t, meRR)

	if meBody.ID <= 0 {
		t.Fatalf("expected positive user ID, got %d", meBody.ID)
	}

	if meBody.Username != "username" {
		t.Fatalf(
			"expected username %q, got %q",
			"username",
			meBody.Username,
		)
	}

	saveReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		VaultRequest{
			EncryptedBlob: 	  validTestEnvelope(),
			ExpectedRevision: 0,
		},
		token,
	)
	
	saveRR := httptest.NewRecorder()
	server.ServeHTTP(saveRR, saveReq)

	if saveRR.Code != http.StatusCreated {
		t.Fatalf(
			"expected save vault status %d, got %d body=%s",
			http.StatusCreated,
			saveRR.Code,
			saveRR.Body.String(),
		)
	}

	saveBody := decodeResponse[struct {
		Status   string `json:"status"`
		Revision int64  `json:"revision"`
	}](t, saveRR)

	if saveBody.Revision != 1 {
		t.Fatalf("expected created vault revision 1, got %d", saveBody.Revision)
	}

	getReq := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/vault",
		nil,
		token,
	)

	getRR := httptest.NewRecorder()
	server.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf(
			"expected get vault status %d, got %d body=%s",
			http.StatusOK,
			getRR.Code,
			getRR.Body.String(),
		)
	}

	getBody := decodeResponse[VaultResponse](t, getRR)

	if !reflect.DeepEqual(getBody.EncryptedBlob, validTestEnvelope()) {
		t.Fatalf(
			"unexpected encrypted blob:\ngot:  %+v\nwant: %+v",
			getBody.EncryptedBlob,
			validTestEnvelope(),
		)
	}

	if getBody.Revision != 1 {
		t.Fatalf("expected revision 1, got %d", getBody.Revision)
	}

	if getBody.UpdatedAt == "" {
		t.Fatal("expected updated_at to be non-empty")
	}
}

// ensures vault access requires an authenticated user
func TestVaultRequiresAuth(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/vault",
		nil,
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusUnauthorized,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures /api/me endpoint requires an authenticated user
func TestMeRequiresAuth(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/me",
		nil,
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusUnauthorized,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that logging in with an incorrect password fails
func TestLoginRejectsWrongPassword(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"username",
		"password",
	)

	startReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/start",
		LoginStartRequest{
			Username:      "username",
			ClientMessage: encodeOpaqueMessage([]byte("WrongPassword123!")),
		},
		"",
	)

	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusOK {
		t.Fatalf(
			"expected login start status %d, got %d body=%s",
			http.StatusOK,
			startRR.Code,
			startRR.Body.String(),
		)
	}

	startBody := decodeResponse[LoginStartResponse](t, startRR)

	finishReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/finish",
		LoginFinishRequest{
			LoginID:       startBody.LoginID,
			ClientMessage: encodeOpaqueMessage([]byte("login-finish")),
		},
		"",
	)

	finishRR := httptest.NewRecorder()
	server.ServeHTTP(finishRR, finishReq)

	if finishRR.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected login finish status %d, got %d body=%s",
			http.StatusUnauthorized,
			finishRR.Code,
			finishRR.Body.String(),
		)
	}
}

// ensures that logins reject usernames without associated accounts
func TestLoginRejectsUnknownUsername(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/start",
		LoginStartRequest{
			Username:      "unknownusername",
			ClientMessage: encodeOpaqueMessage([]byte("password")),
		},
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusUnauthorized,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures router rejects duplicate usernames
func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"username",
		"password",
	)

	req := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/start",
		RegisterStartRequest{
			Username:      "username",
			ClientMessage: encodeOpaqueMessage([]byte("password")),
		},
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf(
			"expected duplicate registration status %d, got %d body=%s",
			http.StatusConflict,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that the router rejects duplicate usernames even if normalized

func TestRegisterRejectsDuplicateNormalizedUsername(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"Username",
		"password",
	)

	req := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/start",
		RegisterStartRequest{
			Username:      "username",
			ClientMessage: encodeOpaqueMessage([]byte("password")),
		},
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf(
			"expected normalized duplicate status %d, got %d body=%s",
			http.StatusConflict,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures registration username is properly normalized
func TestLoginUsesNormalizedUsername(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"Username",
		"password",
	)

	token := performLogin(
		t,
		server,
		"username",
		"password",
	)

	if token == "" {
		t.Fatal("expected login with normalized username to return token")
	}
}

// ensures router rejects fake tokens

func TestVaultRejectsFakeToken(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/vault",
		nil,
		"fake-token",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusUnauthorized,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that the register endpoint rejects invalid JSON
func TestRegisterStartRejectsInvalidJSON(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register/start",
		strings.NewReader("{bad json"),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusBadRequest,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that the register endpoint rejects unknown JSON fields
func TestRegisterStartRejectsUnknownJSONField(t *testing.T) {
	server := setupTestServer(t)

	body := strings.NewReader(`{
		"username": "username",
		"client_message": "dGVzdCBtZXNzYWdlCg==",
		"role": "admin"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register/start",
		body,
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusBadRequest,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that register endpoint rejects invalid client message(s)
func TestRegisterStartRejectsInvalidClientMessage(t *testing.T) {
	server := setupTestServer(t)

	req := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/start",
		RegisterStartRequest{
			Username:      "username",
			ClientMessage: "invalid-b64!",
		},
		"",
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusBadRequest,
			rr.Code,
			rr.Body.String(),
		)
	}
}

// ensures that router rejects registration finish replays
func TestRegisterFinishCannotBeReplayed(t *testing.T) {
	server := setupTestServer(t)

	message := []byte("password")

	startReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/start",
		RegisterStartRequest{
			Username: 	   "username",
			ClientMessage: encodeOpaqueMessage(message),
		},
		"",
	)

	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusOK,
			startRR.Code,
			startRR.Body.String(),
		)
	}

	startBody := decodeResponse[RegisterStartResponse](t, startRR)

	finishBody := RegisterFinishRequest{
		RegistrationID: startBody.RegistrationID,
		ClientMessage:  encodeOpaqueMessage(message),
	}

	firstReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/finish",
		finishBody,
		"",
	)

	firstRR := httptest.NewRecorder()
	server.ServeHTTP(firstRR, firstReq)

	if firstRR.Code != http.StatusCreated {
		t.Fatalf(
			"expected first finish status %d, got %d body=%s",
			http.StatusCreated,
			firstRR.Code,
			firstRR.Body.String(),
		)
	}

	secondReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/register/finish",
		finishBody,
		"",
	)

	secondRR := httptest.NewRecorder()
	server.ServeHTTP(secondRR, secondReq)

	if secondRR.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected replayed finish status %d, got %d body=%s",
			http.StatusBadRequest,
			secondRR.Code,
			secondRR.Body.String(),
		)
	}
}

// ensures that router rejects login finish replays
func TestLoginFinishCannotBeReplayed(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"username",
		"password",
	)

	startReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/start",
		LoginStartRequest{
			Username:      "username",
			ClientMessage: encodeOpaqueMessage([]byte("password")),
		},
		"",
	)

	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusOK {
		t.Fatalf(
			"expected start status %d, got %d body=%s",
			http.StatusOK,
			startRR.Code,
			startRR.Body.String(),
		)
	}

	startBody := decodeResponse[LoginStartResponse](t, startRR)

	finishBody := LoginFinishRequest{
		LoginID:       startBody.LoginID,
		ClientMessage: encodeOpaqueMessage([]byte("login-finish")),
	}

	firstReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/finish",
		finishBody,
		"",
	)

	firstRR := httptest.NewRecorder()
	server.ServeHTTP(firstRR, firstReq)

	if firstRR.Code != http.StatusOK {
		t.Fatalf(
			"expected first finish status %d, got %d body=%s",
			http.StatusOK,
			firstRR.Code,
			firstRR.Body.String(),
		)
	}

	secondReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/auth/login/finish",
		finishBody,
		"",
	)

	secondRR := httptest.NewRecorder()
	server.ServeHTTP(secondRR, secondReq)

	if secondRR.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected replayed finish status %d, got %d body=%s",
			http.StatusUnauthorized,
			secondRR.Code,
			secondRR.Body.String(),
		)
	}
}

// ensures that multiple concurrent registration flows remain valid
func TestSeparateRegistrationFlowsRemainValid(t *testing.T) {
	server := setupTestServer(t)

	firstMessage := []byte("first-password")
	secondMessage := []byte("second-password")
	
	startRegistration := func(username string, message []byte) RegisterStartResponse {
		t.Helper()

		req := makeJSONRequest(
			t,
			http.MethodPost,
			"/api/auth/register/start",
			RegisterStartRequest{
				Username: 	   username,
				ClientMessage: encodeOpaqueMessage(message),
			},
			"",
		)

		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf(
				"expected registration start status %d, got %d body=%s",
				http.StatusOK,
				rr.Code,
				rr.Body.String(),
			)
		}

		return decodeResponse[RegisterStartResponse](t, rr)
	}

	first := startRegistration("firstuser", firstMessage)
	second := startRegistration("seconduser", secondMessage)

	
	finishRegistration := func(
		start RegisterStartResponse,
		message []byte,
	) {
		t.Helper()

		req := makeJSONRequest(
			t,
			http.MethodPost,
			"/api/auth/register/finish",
			RegisterFinishRequest{
				RegistrationID: start.RegistrationID,
				ClientMessage:  encodeOpaqueMessage(message),
			},
			"",
		)

		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf(
				"expected registration finish status %d, got %d body=%s",
				http.StatusCreated,
				rr.Code,
				rr.Body.String(),
			)
		}
	}

	// Starting the second flow must not delete the first pending flow.
	finishRegistration(first, firstMessage)
	finishRegistration(second, secondMessage)
}

// ensures that multiple concurrent login flows remain valid
func TestSeparateLoginFlowsRemainValid(t *testing.T) {
	server := setupTestServer(t)

	performRegistration(
		t,
		server,
		"username",
		"password",
	)

	startLogin := func() LoginStartResponse {
		t.Helper()

		req := makeJSONRequest(
			t,
			http.MethodPost,
			"/api/auth/login/start",
			LoginStartRequest{
				Username:      "username",
				ClientMessage: encodeOpaqueMessage([]byte("password")),
			},
			"",
		)

		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf(
				"expected login start status %d, got %d body=%s",
				http.StatusOK,
				rr.Code,
				rr.Body.String(),
			)
		}

		return decodeResponse[LoginStartResponse](t, rr)
	}

	first := startLogin()
	second := startLogin()

	finishLogin := func(start LoginStartResponse) {
		t.Helper()

		req := makeJSONRequest(
			t,
			http.MethodPost,
			"/api/auth/login/finish",
			LoginFinishRequest{
				LoginID:       start.LoginID,
				ClientMessage: encodeOpaqueMessage([]byte("login-finish")),
			},
			"",
		)

		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf(
				"expected login finish status %d, got %d body=%s",
				http.StatusOK,
				rr.Code,
				rr.Body.String(),
			)
		}
	}

	// Starting the second login must not delete the first pending login.
	finishLogin(first)
	finishLogin(second)
}

// ensures vault revision conflicts are handled smoothly
func TestVaultRevisionConflict(t *testing.T) {
	server := setupTestServer(t)

	token := registerAndLogin(
		t,
		server,
		"username",
		"password",
	)

	createReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		VaultRequest{
			EncryptedBlob: 	  validTestEnvelope(),
			ExpectedRevision: 0,
		},
		token,
	)

	createRR := httptest.NewRecorder()
	server.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf(
			"expected create vault status %d, got %d body=%s",
			http.StatusCreated,
			createRR.Code,
			createRR.Body.String(),
		)
	}

	updateReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		VaultRequest{
			EncryptedBlob: validTestEnvelope(),
			ExpectedRevision: 1,
		},
		token,
	)

	updateRR := httptest.NewRecorder()
	server.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf(
			"expected update vault status %d, got %d body=%s",
			http.StatusOK,
			updateRR.Code,
			updateRR.Body.String(),
		)
	}

	updateBody := decodeResponse[struct {
		Revision int64 `json:"revision"`
	}](t, updateRR)

	if updateBody.Revision != 2 {
		t.Fatalf(
			"expected updated vault revision 2, got %d",
			updateBody.Revision,
		)
	}

	conflictReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		VaultRequest{
			EncryptedBlob: 	  validTestEnvelope(),
			ExpectedRevision: 1,
		},
		token,
	)

	conflictRR := httptest.NewRecorder()
	server.ServeHTTP(conflictRR, conflictReq)

	if conflictRR.Code != http.StatusConflict {
		t.Fatalf(
			"expected conflict status %d, got %d body=%s",
			http.StatusConflict,
			conflictRR.Code,
			conflictRR.Body.String(),
		)
	}
}

// ensures that trying to create a vault after making one will fail
func TestCreatingSecondVaultReturnsConflict(t *testing.T) {
	server := setupTestServer(t)

	token := registerAndLogin(
		t,
		server,
		"username",
		"StrongPassword123!",
	)

	body := VaultRequest{
		EncryptedBlob:    validTestEnvelope(),
		ExpectedRevision: 0,
	}

	firstReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		body,
		token,
	)

	firstRR := httptest.NewRecorder()
	server.ServeHTTP(firstRR, firstReq)

	if firstRR.Code != http.StatusCreated {
		t.Fatalf(
			"expected first create status %d, got %d body=%s",
			http.StatusCreated,
			firstRR.Code,
			firstRR.Body.String(),
		)
	}

	secondReq := makeJSONRequest(
		t,
		http.MethodPost,
		"/api/vault",
		body,
		token,
	)

	secondRR := httptest.NewRecorder()
	server.ServeHTTP(secondRR, secondReq)

	if secondRR.Code != http.StatusConflict {
		t.Fatalf(
			"expected second create status %d, got %d body=%s",
			http.StatusConflict,
			secondRR.Code,
			secondRR.Body.String(),
		)
	}
}

// ensures router properly handles missing vault when vault is queried
func TestGetVaultReturnsNotFoundBeforeCreation(t *testing.T) {
	server := setupTestServer(t)

	token := registerAndLogin(
		t,
		server,
		"username",
		"password",
	)

	req := makeJSONRequest(
		t,
		http.MethodGet,
		"/api/vault",
		nil,
		token,
	)

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d body=%s",
			http.StatusNotFound,
			rr.Code,
			rr.Body.String(),
		)
	}
}
