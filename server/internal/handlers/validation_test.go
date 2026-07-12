package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fully tests username validation (per the regex)
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		wantErr     bool
		wantMessage string
	}{
		{
			name:     "valid username",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "minimum length",
			username: "abc",
			wantErr:  false,
		},
		{
			name:     "maximum length",
			username: strings.Repeat("a", 24),
			wantErr:  false,
		},
		{
			name:     "uppercase and numbers",
			username: "TestUser123",
			wantErr:  false,
		},
		{
			name:     "leading and trailing whitespace",
			username: "  testuser  ",
			wantErr:  false,
		},
		{
			name:        "empty username",
			username:    "",
			wantErr:     true,
			wantMessage: "username must be between 3 and 24 characters",
		},
		{
			name:        "whitespace only",
			username:    "   ",
			wantErr:     true,
			wantMessage: "username must be between 3 and 24 characters",
		},
		{
			name:        "too short",
			username:    "ab",
			wantErr:     true,
			wantMessage: "username must be between 3 and 24 characters",
		},
		{
			name:        "too long",
			username:    strings.Repeat("a", 25),
			wantErr:     true,
			wantMessage: "username must be between 3 and 24 characters",
		},
		{
			name:        "contains underscore",
			username:    "test_user",
			wantErr:     true,
			wantMessage: "username must be alphanumeric",
		},
		{
			name:        "contains hyphen",
			username:    "test-user",
			wantErr:     true,
			wantMessage: "username must be alphanumeric",
		},
		{
			name:        "contains space",
			username:    "test user",
			wantErr:     true,
			wantMessage: "username must be alphanumeric",
		},
		{
			name:        "contains punctuation",
			username:    "testuser!",
			wantErr:     true,
			wantMessage: "username must be alphanumeric",
		},
		{
			name:        "contains non ASCII letters",
			username:    "téstuser",
			wantErr:     true,
			wantMessage: "username must be alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantMessage {
				t.Fatalf(
					"expected error %q, got %q",
					tt.wantMessage,
					err.Error(),
				)
			}
		})
	}
}

// fully tests Opaque client message validation
func TestValidateOpaqueClientMessage(t *testing.T) {
	tests := []struct {
		name          string
		clientMessage string
		wantErr       bool
	}{
		{
			name:          "valid message",
			clientMessage: "dGVzdC1tZXNzYWdl",
			wantErr:       false,
		},
		{
			name:          "single character",
			clientMessage: "a",
			wantErr:       false,
		},
		{
			name:          "message with surrounding whitespace",
			clientMessage: "  dGVzdA  ",
			wantErr:       false,
		},
		{
			name:          "empty message",
			clientMessage: "",
			wantErr:       true,
		},
		{
			name:          "whitespace only",
			clientMessage: "   ",
			wantErr:       true,
		},
		{
			name:          "tabs and newlines only",
			clientMessage: "\t\n\r",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOpaqueClientMessage(tt.clientMessage)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if err.Error() != "client_message is required" {
					t.Fatalf(
						"expected error %q, got %q",
						"client_message is required",
						err.Error(),
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// this tests flow ID validation
func TestValidateFlowID(t *testing.T) {
	tests := []struct {
		name    string
		flowID  string
		wantErr bool
	}{
		{
			name:    "valid flow ID",
			flowID:  "registration-flow-id",
			wantErr: false,
		},
		{
			name:    "UUID-like flow ID",
			flowID:  "cbad5406-6105-4e3b-8cf9-626c924a346f",
			wantErr: false,
		},
		{
			name:    "flow ID with surrounding whitespace",
			flowID:  "  flow-id  ",
			wantErr: false,
		},
		{
			name:    "empty flow ID",
			flowID:  "",
			wantErr: true,
		},
		{
			name:    "whitespace-only flow ID",
			flowID:  "   ",
			wantErr: true,
		},
		{
			name:    "tabs and newlines only",
			flowID:  "\t\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlowID(tt.flowID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if err.Error() != "flow id is required" {
					t.Fatalf(
						"expected error %q, got %q",
						"flow id is required",
						err.Error(),
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// this tests vault request validation, not envelope format
func TestValidateVaultRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         VaultRequest
		wantErr     bool
		wantMessage string
	}{
		{
			name: "valid new vault",
			req: VaultRequest{
				EncryptedBlob:    validTestEnvelope(),
				ExpectedRevision: 0,
			},
			wantErr: false,
		},
		{
			name: "valid existing vault update",
			req: VaultRequest{
				EncryptedBlob:    validTestEnvelope(),
				ExpectedRevision: 5,
			},
			wantErr: false,
		},
		{
			name: "negative expected revision",
			req: VaultRequest{
				EncryptedBlob:    validTestEnvelope(),
				ExpectedRevision: -1,
			},
			wantErr:     true,
			wantMessage: "expected_revision cannot be negative",
		},
		{
			name: "invalid empty envelope",
			req: VaultRequest{
				ExpectedRevision: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVaultRequest(&tt.req)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantMessage != "" && err.Error() != tt.wantMessage {
				t.Fatalf(
					"expected error %q, got %q",
					tt.wantMessage,
					err.Error(),
				)
			}
		})
	}
}

// this tests whether valid JSON is accepted
func TestDecodeJSONAcceptsValidJSON(t *testing.T) {
	type requestBody struct {
		Username string `json:"username"`
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{"username":"testuser"}`),
	)
	rr := httptest.NewRecorder()

	var body requestBody

	err := decodeJSON(rr, req, &body, 1024)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if body.Username != "testuser" {
		t.Fatalf(
			"expected username %q, got %q",
			"testuser",
			body.Username,
		)
	}
}

// this tests whether invalid JSON is rejected
func TestDecodeJSONRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{"username":`),
	)
	rr := httptest.NewRecorder()

	var body struct {
		Username string `json:"username"`
	}

	err := decodeJSON(rr, req, &body, 1024)
	if err == nil {
		t.Fatal("expected invalid JSON error, got nil")
	}
}

// this tests whether unknown fields are rejected
func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{
			"username": "testuser",
			"role": "admin"
		}`),
	)
	rr := httptest.NewRecorder()

	var body struct {
		Username string `json:"username"`
	}

	err := decodeJSON(rr, req, &body, 1024)
	if err == nil {
		t.Fatal("expected unknown field error, got nil")
	}

	if !strings.Contains(err.Error(), `unknown field "role"`) {
		t.Fatalf(
			"expected unknown field error, got %q",
			err.Error(),
		)
	}
}

// this tests whether oversized bodies are rejected
func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{"value":"this body is too large"}`),
	)
	rr := httptest.NewRecorder()

	var body struct {
		Value string `json:"value"`
	}

	err := decodeJSON(rr, req, &body, 10)
	if err == nil {
		t.Fatal("expected body size error, got nil")
	}

	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		t.Fatalf(
			"expected *http.MaxBytesError, got %T: %v",
			err,
			err,
		)
	}
}

// this tests whether empty bodies are rejected
func TestDecodeJSONRejectsEmptyBody(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		http.NoBody,
	)
	rr := httptest.NewRecorder()

	var body struct {
		Username string `json:"username"`
	}

	err := decodeJSON(rr, req, &body, 1024)
	if err == nil {
		t.Fatal("expected empty body error, got nil")
	}
}